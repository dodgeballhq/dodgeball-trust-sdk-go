package dodgeball

import "testing"

// Set the secret key appropriately when executing tests
const DODGEBALL_SECRET_KEY = ""

func TestDodgeball_Checkpoint(t *testing.T) {
	type fields struct {
		secret string
		config *Config
	}
	type args struct {
		request CheckpointRequest
	}
	config := Config{
		APIURL:     "https://api.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  true,
	}
	configDisabled := Config{
		APIURL:     "https:/api.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  false,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"checkpoint test1", fields{DODGEBALL_SECRET_KEY, &config}, args{CheckpointRequest{
			CheckpointName: "TEST_CHECKPOINT",
			Event: CheckpointEvent{
				IP: "123.123.123.123",
				Data: map[string]interface{}{
					"mfa": map[string]interface{}{
						"phoneNumbers": "+16175551212",
					},
					"customer": map[string]interface{}{
						"firstName":    "John",
						"middleName":   "A",
						"lastName":     "Smith",
						"primaryEmail": "john.a.smith@example.com",
						"primaryPhone": "+1 2049381968",
						"dateOfBirth":  "2003-02-14",
						"taxId":        "111111100",
						"createdAt":    "2010-01-01",
					},
					"transaction": map[string]interface{}{
						"externalId": "badExternalId",
						"currency":   "USD",
						"amount":     1000,
					},
					"deduce": map[string]interface{}{"isTest": true}},
			},
			Options:           CheckpointResponseOptions{},
			SessionID:         "64de1794-8bb9-11ed-a1eb-0242ac120004",
			UserID:            "64de1794-8bb9-11ed-a1eb-0242ac120002",
			UseVerificationID: "",
		}}, false},
		{"checkpoint test2", fields{DODGEBALL_SECRET_KEY, &configDisabled}, args{CheckpointRequest{
			CheckpointName: "TEST_CHECKPOINT2",
			Event: CheckpointEvent{
				IP:   "123.123.123.123",
				Data: map[string]interface{}{},
			},
			SessionID:         "64de1794-8bb9-11ed-a1eb-0242ac120004",
			UserID:            "64de1794-8bb9-11ed-a1eb-0242ac120002",
			UseVerificationID: "",
		}}, false},
		{"checkpoint test3", fields{DODGEBALL_SECRET_KEY, &configDisabled}, args{CheckpointRequest{
			CheckpointName: "TEST_CHECKPOINT2",
			Event: CheckpointEvent{
				IP:   "123.123.123.123",
				Data: map[string]interface{}{},
			},
			SourceToken:       "64de1794-8bb9-11ed-a1eb-0242ac120004",
			UserID:            "64de1794-8bb9-11ed-a1eb-0242ac120002",
			UseVerificationID: "",
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dodgeball{
				secret: tt.fields.secret,
				config: tt.fields.config,
			}
			if _, err := d.Checkpoint(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("Dodgeball.Checkpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDodgeball_Event(t *testing.T) {
	type fields struct {
		secret string
		config *Config
	}
	type args struct {
		request TrackOptions
	}
	config := Config{
		APIURL:     "https://api.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  true,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"track test1", fields{DODGEBALL_SECRET_KEY, &config}, args{TrackOptions{
			Event: TrackEvent{
				Type: "TEST_TRACK_EVENT",
				Data: nil,
			},
			SourceToken: "",
			SessionID:   "64de1794-8bb9-11ed-a1eb-0242ac120004",
			UserID:      "64de1794-8bb9-11ed-a1eb-0242ac120002",
		}}, false},
		{"track test2", fields{DODGEBALL_SECRET_KEY, &config}, args{TrackOptions{
			Event: TrackEvent{
				Type: "TEST_TRACK_EVENT",
				Data: nil,
			},
			SourceToken: "64de1794-8bb9-11ed-a1eb-0242ac120004",
			UserID:      "64de1794-8bb9-11ed-a1eb-0242ac120002",
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dodgeball{
				secret: tt.fields.secret,
				config: tt.fields.config,
			}
			if err := d.Event(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("Dodgeball.Event() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
