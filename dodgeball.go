package dodgeball

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// APIVersion is the currently supported API version
	APIVersion = "v1"

	// APIURL is the URL of the service backend
	APIURL = "https://api.dodgeballhq.com/"

	// BaseCheckpointTimeout is the default timeout for a checkpoint
	BaseCheckpointTimeout = 100

	// MaxTimeout is the maximum timeout for a checkpoint
	MaxTimeout = 10000

	// MaxRetryCount is the maximum number of retries for a checkpoint
	MaxRetryCount = 3
)

var (
	ErrMissingCheckpointName = errors.New("checkpoint name is required")
	ErrMissingEventIP        = errors.New("event IP is required")
	ErrMissingDodgeballID    = errors.New("Dodgeball ID is required")

	internalOpts = &CheckpointResponseOptions{
		Sync:    true,
		Timeout: BaseCheckpointTimeout,
	}
)

// Config is the configuration for the Dodgeball client
type Config struct {
	APIVersion string
	APIURL     string
}

// NewConfig returns a new Config using defaults
func NewConfig() *Config {
	return &Config{
		APIVersion: APIVersion,
		APIURL:     APIURL,
	}
}

// Dodgeball is the client for the Dodgeball service
type Dodgeball struct {
	secret string
	config *Config
}

// Checkpoint will check with Dodgeball to verify if the request is allowed to proceed
func (d *Dodgeball) Checkpoint(request CheckpointRequest) (*CheckpointResponse, error) {
	if request.CheckpointName == "" {
		return nil, ErrMissingCheckpointName
	}

	if request.Event.IP == "" {
		return nil, ErrMissingEventIP
	}

	if request.DodgeballID == "" {
		return nil, ErrMissingDodgeballID
	}

	var checkpointResponse CheckpointResponse
	numRepeats := 0
	numFailures := 0

	for !checkpointResponse.Success && numRepeats < MaxRetryCount {
		resp, err := d.verify(&request)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(resp, &checkpointResponse)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling verify response %s", err.Error())
		}

		numRepeats++
	}

	if !checkpointResponse.Success {
		return nil, fmt.Errorf("verify failed after %d attempts", numRepeats)
	}

	isResolved := checkpointResponse.Verification.Status != VerificationStatusPending
	verificationID := checkpointResponse.Verification.ID
	var verificationResponse CheckpointResponse

	for !isResolved && numFailures < MaxRetryCount {
		// TODO: some timeout?
		time.Sleep(time.Millisecond * time.Duration(internalOpts.Timeout))

		resp, err := d.verification(&request, verificationID)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(resp, &verificationResponse)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling verification response %s", err.Error())
		}

		if !verificationResponse.Success {
			numFailures++
			continue
		}

		switch status := verificationResponse.Verification.Status; {
		case status == "":
			numFailures++
		case status != VerificationStatusPending:
			isResolved = true
		default:
			numRepeats++
		}
	}

	return &verificationResponse, nil
}

type requestParams struct {
	method   string
	endpoint string
	headers  map[string]string
	data     interface{}
}

func (d *Dodgeball) verify(request *CheckpointRequest) ([]byte, error) {
	params := requestParams{
		method:   http.MethodPost,
		endpoint: "/verify",
		headers: map[string]string{
			"Dodgeball-Verification-Id": request.UseVerificationID,
			"Dodgeball-Source-Id":       request.DodgeballID,
			"Dodgeball-Customer-Id":     request.UserID,
		},
		data: map[string]interface{}{
			"type":    request.CheckpointName,
			"event":   request.Event,
			"options": internalOpts,
		},
	}

	resp, err := d.request(params)
	if err != nil {
		return nil, fmt.Errorf("error calling verify %s", err.Error())
	}

	return resp, nil
}

func (d *Dodgeball) verification(request *CheckpointRequest, verificationID string) ([]byte, error) {
	params := requestParams{
		method:   http.MethodGet,
		endpoint: "/verification/" + verificationID,
		headers: map[string]string{
			"Dodgeball-Verification-Id": request.UseVerificationID,
			"Dodgeball-Source-Id":       request.DodgeballID,
			"Dodgeball-Customer-Id":     request.UserID,
		},
	}
	resp, err := d.request(params)
	if err != nil {
		return nil, fmt.Errorf("error calling verification %s", err.Error())
	}

	return resp, nil
}

func (d *Dodgeball) request(params requestParams) ([]byte, error) {
	client := &http.Client{
		// TODO: make timeout configurable
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(params.method, d.buildURL(params.endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Dodgeball-Secret-Key", d.secret)

	for k, v := range params.headers {
		req.Header.Add(k, v)
	}

	if params.data != nil {
		dataBytes, err := json.Marshal(params.data)
		if err != nil {
			return nil, fmt.Errorf("error marshaling data %s", err.Error())
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(dataBytes))
	}

	httpResponse, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling endpoint %s", err.Error())
	}
	defer httpResponse.Body.Close()

	return ioutil.ReadAll(httpResponse.Body)
}

func (d *Dodgeball) buildURL(endpoint string) string {
	return fmt.Sprintf("%s%s%s", d.config.APIURL, d.config.APIVersion, endpoint)
}

// New returns a new Dodgeball client
func New(secret string, config *Config) *Dodgeball {
	return &Dodgeball{
		secret: secret,
		config: config,
	}
}
