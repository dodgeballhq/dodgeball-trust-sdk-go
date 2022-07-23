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
		APIURL:     "http://localhost:3001/",
		APIVersion: "v1",
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"test1", fields{"secret", &config}, args{CheckpointRequest{
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
