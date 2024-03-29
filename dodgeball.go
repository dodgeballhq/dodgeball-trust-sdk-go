package dodgeball

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	// Maximal Sleep in between polling invocations
	MaxPollingSleep = 1000

	// MaxRetryCount is the maximum number of retries for a checkpoint
	MaxRetryCount = 3
)

var (
	ErrMissingCheckpointName         = errors.New("checkpoint name is required")
	ErrMissingEventIP                = errors.New("event IP is required")
	ErrMissingSessionIDOrSourceToken = errors.New("either session ID or sourceToken is required")
)

// Config is the configuration for the Dodgeball client
type Config struct {
	APIVersion string
	APIURL     string
	IsEnabled  bool
}

// NewConfig returns a new Config using defaults
func NewConfig() *Config {
	return &Config{
		APIVersion: APIVersion,
		APIURL:     APIURL,
		IsEnabled:  true,
	}
}

// Dodgeball is the client for the Dodgeball service
type Dodgeball struct {
	secret string
	config *Config
}

// Track will add additional information about a user's journey by submitting events from your server
func (d *Dodgeball) Event(options TrackOptions) error {
	if options.Event.EventTime == 0 {
		options.Event.EventTime = time.Now().UnixMilli()
	}

	resp, err := d.event(&options)
	if err != nil {
		return err
	}

	var trackResponse TrackResponse
	err = json.Unmarshal(resp, &trackResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling track response %s", err.Error())
	}

	if !trackResponse.Success {
		return fmt.Errorf("track failed")
	}

	return nil
}

// Checkpoint will check with Dodgeball to verify if the request is allowed to proceed
func (d *Dodgeball) Checkpoint(request CheckpointRequest) (*CheckpointResponse, error) {
	if request.CheckpointName == "" {
		return nil, ErrMissingCheckpointName
	}

	if request.Event.IP == "" {
		return nil, ErrMissingEventIP
	}

	if request.SessionID == "" && request.SourceToken == "" {
		return nil, ErrMissingSessionIDOrSourceToken
	}

	if !d.config.IsEnabled {
		return &CheckpointResponse{
			Success: true,
			Errors:  []CheckpointResponseError{},
			Version: d.config.APIVersion,
			Verification: CheckpointResponseVerification{
				ID:      "DODGEBALL_DISABLED",
				Status:  VerificationStatusComplete,
				Outcome: VerificationOutcomeApproved,
			},
		}, nil
	}

	request.Event.Type = request.CheckpointName

	trivialTimeout := request.Options.Timeout <= 0
	largeTimeout := request.Options.Timeout > 5*BaseCheckpointTimeout
	mustPoll := trivialTimeout || largeTimeout
	activeTimeout := BaseCheckpointTimeout
	switch {
	case mustPoll:
		activeTimeout = BaseCheckpointTimeout
	case !trivialTimeout:
		activeTimeout = request.Options.Timeout
	}

	internalOpts := &CheckpointResponseOptions{
		Sync:    false, // TODO: make configurable
		Timeout: activeTimeout,
		Webhook: request.Options.Webhook,
	}

	var verificationResponse CheckpointResponse
	numRepeats := 0
	numFailures := 0

	for !verificationResponse.Success && numRepeats < MaxRetryCount {
		resp, err := d.verify(&request, internalOpts)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(resp, &verificationResponse)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling verify response %s", err.Error())
		}

		numRepeats++
	}

	if !verificationResponse.Success {
		return nil, fmt.Errorf("verify failed after %d attempts", numRepeats)
	}

	isResolved := verificationResponse.Verification.Status != VerificationStatusPending
	verificationID := verificationResponse.Verification.ID
	elapsedTime := activeTimeout
	pollingSleep := 100

	for (trivialTimeout || request.Options.Timeout > elapsedTime) && !isResolved && (numFailures < MaxRetryCount) {
		time.Sleep(time.Millisecond * time.Duration(pollingSleep))

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

		elapsedTime += pollingSleep
		if pollingSleep < MaxPollingSleep {
			if (2 * pollingSleep) <= MaxPollingSleep {
				pollingSleep = 2 * pollingSleep
			}
		}
	}

	if numFailures >= MaxRetryCount {
		verificationResponse.Success = false
		verificationResponse.timedOut = true
		verificationResponse.Errors = append(verificationResponse.Errors, CheckpointResponseError{Code: 503, Message: "Service Unavailable: Maximum retry count exceeded"})
	}

	return &verificationResponse, nil
}

type requestParams struct {
	method   string
	endpoint string
	headers  map[string]string
	data     interface{}
}

func (d *Dodgeball) event(request *TrackOptions) ([]byte, error) {
	params := requestParams{
		method:   http.MethodPost,
		endpoint: "/track",
		headers: map[string]string{
			"Dodgeball-Source-Token": request.SourceToken,
			"Dodgeball-Customer-Id":  request.UserID,
			"Dodgeball-Session-Id":   request.SessionID,
		},
		data: map[string]interface{}{
			"type":      request.Event.Type,
			"data":      request.Event.Data,
			"eventTime": request.Event.EventTime,
		},
	}

	resp, err := d.request(params)
	if err != nil {
		return nil, fmt.Errorf("error calling track %s", err.Error())
	}

	return resp, nil
}

func (d *Dodgeball) verify(request *CheckpointRequest, internalOpts *CheckpointResponseOptions) ([]byte, error) {
	headers := map[string]string{
		"Dodgeball-Verification-Id": request.UseVerificationID,
		"Dodgeball-Customer-Id":     request.UserID,
		"Dodgeball-Session-Id":      request.SessionID,
	}

	if request.SourceToken != "" {
		headers["Dodgeball-Source-Token"] = request.SourceToken
	}

	params := requestParams{
		method:   http.MethodPost,
		endpoint: "/checkpoint",
		headers:  headers,
		data: map[string]interface{}{
			"event":   request.Event,
			"options": internalOpts,
		},
	}

	resp, err := d.request(params)
	if err != nil {
		return nil, fmt.Errorf("error calling checkpoint %s", err.Error())
	}

	return resp, nil
}

func (d *Dodgeball) verification(request *CheckpointRequest, verificationID string) ([]byte, error) {
	params := requestParams{
		method:   http.MethodGet,
		endpoint: "/verification/" + verificationID,
		headers: map[string]string{
			"Dodgeball-Verification-Id": request.UseVerificationID,
			"Dodgeball-Source-Token":    request.SourceToken,
			"Dodgeball-Customer-Id":     request.UserID,
			"Dodgeball-Session-Id":      request.SessionID,
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
		req.Body = io.NopCloser(bytes.NewReader(dataBytes))
	}

	httpResponse, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling endpoint %s", err.Error())
	}
	defer httpResponse.Body.Close()

	return io.ReadAll(httpResponse.Body)
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
