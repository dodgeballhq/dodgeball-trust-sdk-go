package dodgeball

import "testing"

const DODGEBALL_SECRET_KEY = "1c29d5d6593011ec9412470128c0fd71"

func TestDodgeball_Checkpoint(t *testing.T) {
	type fields struct {
		secret string
		config *Config
	}
	type args struct {
		request CheckpointRequest
	}
	config := Config{
		APIURL:     "http://localhost:3001/",
		APIVersion: "v1",
		IsEnabled:  true,
	}
	configDisabled := Config{
		APIURL:     "http://localhost:3001/",
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
				IP:   "123.123.123.123",
				Data: map[string]interface{}{},
			},
			Options: CheckpointResponseOptions{
				Timeout: 1000,
			},
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
		APIURL:     "http://localhost:3001/",
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
