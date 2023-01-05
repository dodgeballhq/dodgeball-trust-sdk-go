package dodgeball

import "testing"

func TestDodgeball_Checkpoint(t *testing.T) {
	type fields struct {
		secret string
		config *Config
	}
	type args struct {
		request CheckpointRequest
	}
	config := Config{
		APIURL:     "https://api.dev.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  true,
	}
	configDisabled := Config{
		APIURL:     "https://api.dev.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  false,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"checkpoint test1", fields{"secret", &config}, args{CheckpointRequest{
			CheckpointName: "test",
			Event: CheckpointEvent{
				IP:   "123.123.123.123",
				Data: nil,
			},
			SourceToken:       "dodgeballID",
			SessionID:         "sessionID",
			UserID:            "userID",
			UseVerificationID: "verifyID",
		}}, false},
		{"checkpoint test2", fields{"secret", &configDisabled}, args{CheckpointRequest{
			CheckpointName: "test",
			Event: CheckpointEvent{
				IP:   "123.123.123.123",
				Data: nil,
			},
			SourceToken:       "dodgeballID",
			SessionID:         "sessionID",
			UserID:            "userID",
			UseVerificationID: "verifyID",
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

func TestDodgeball_Track(t *testing.T) {
	type fields struct {
		secret string
		config *Config
	}
	type args struct {
		request TrackOptions
	}
	config := Config{
		APIURL:     "https://api.dev.dodgeballhq.com/",
		APIVersion: "v1",
		IsEnabled:  true,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"track test1", fields{"secret", &config}, args{TrackOptions{
			Event: TrackEvent{
				Type: "TEST_TRACK_EVENT",
				Data: nil,
			},
			SourceToken: "dodgeballID",
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
			if err := d.Track(tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("Dodgeball.Track() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
