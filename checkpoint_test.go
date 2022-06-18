package dodgeball

import (
	"testing"
)

type fields struct {
	Success      bool                      `json:"success"`
	Errors       []CheckpointResponseError `json:"errors"`
	Version      string                    `json:"version"`
	Verification struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Outcome string `json:"outcome"`
	} `json:"verification"`
	timedOut bool `json:"-"`
}

func mockResponse(success, timedOut bool, status string, outcome string) *fields {
	return &fields{
		Success: success,
		Errors:  nil,
		Version: "v1",
		Verification: struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Outcome string `json:"outcome"`
		}{
			ID:      "verifyID",
			Status:  status,
			Outcome: outcome,
		},
		timedOut: timedOut,
	}
}

func TestCheckpointResponse_IsRunning(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"running: pending", *mockResponse(true, false, "PENDING", "PENDING"), true},
		{"running: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), true},
		{"running: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), false},
		{"running: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), false},
		{"running: failed", *mockResponse(false, false, "FAILED", "ERROR"), false},
		{"running: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), false},
		{"running: timedout", *mockResponse(false, true, "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.IsRunning(); got != tt.want {
				t.Errorf("CheckpointResponse.IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointResponse_IsAllowed(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"allowed: pending", *mockResponse(true, false, "PENDING", "PENDING"), false},
		{"allowed: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), false},
		{"allowed: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), true},
		{"allowed: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), false},
		{"allowed: failed", *mockResponse(false, false, "FAILED", "ERROR"), false},
		{"allowed: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), false},
		{"allowed: timedout", *mockResponse(false, true, "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.IsAllowed(); got != tt.want {
				t.Errorf("CheckpointResponse.IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointResponse_IsDenied(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"denied: pending", *mockResponse(true, false, "PENDING", "PENDING"), false},
		{"denied: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), false},
		{"denied: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), false},
		{"denied: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), true},
		{"denied: failed", *mockResponse(false, false, "FAILED", "ERROR"), false},
		{"denied: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), false},
		{"denied: timedout", *mockResponse(false, true, "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.IsDenied(); got != tt.want {
				t.Errorf("CheckpointResponse.IsDenied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointResponse_IsUndecided(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"undecided: pending", *mockResponse(true, false, "PENDING", "PENDING"), false},
		{"undecided: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), false},
		{"undecided: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), false},
		{"undecided: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), false},
		{"undecided: failed", *mockResponse(false, false, "FAILED", "ERROR"), false},
		{"undecided: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), true},
		{"undecided: timedout", *mockResponse(false, true, "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.IsUndecided(); got != tt.want {
				t.Errorf("CheckpointResponse.IsUndecided() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointResponse_HasError(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"error: pending", *mockResponse(true, false, "PENDING", "PENDING"), false},
		{"error: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), false},
		{"error: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), false},
		{"error: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), false},
		{"error: failed", *mockResponse(false, false, "FAILED", "ERROR"), true},
		{"error: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), false},
		{"error: timedout", *mockResponse(false, true, "", ""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.HasError(); got != tt.want {
				t.Errorf("CheckpointResponse.HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckpointResponse_IsTimeout(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"timedout: pending", *mockResponse(true, false, "PENDING", "PENDING"), false},
		{"timedout: blocked", *mockResponse(true, false, "BLOCKED", "PENDING"), false},
		{"timedout: approved", *mockResponse(true, false, "COMPLETE", "APPROVED"), false},
		{"timedout: denied", *mockResponse(true, false, "COMPLETE", "DENIED"), false},
		{"timedout: failed", *mockResponse(false, false, "FAILED", "ERROR"), false},
		{"timedout: undecided", *mockResponse(true, false, "COMPLETE", "PENDING"), false},
		{"timedout: timedout", *mockResponse(false, true, "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &CheckpointResponse{
				Success:      tt.fields.Success,
				Errors:       tt.fields.Errors,
				Version:      tt.fields.Version,
				Verification: tt.fields.Verification,
			}
			if got := cr.IsTimeout(); got != tt.want {
				t.Errorf("CheckpointResponse.IsTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
