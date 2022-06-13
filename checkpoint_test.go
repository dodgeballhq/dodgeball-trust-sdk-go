package dodgeball

import (
	"testing"
)

type fields struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Version      string `json:"version"`
	Verification struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Outcome string `json:"outcome"`
	} `json:"verification"`
}

func mockResponse(success bool, status string, outcome string) *fields {
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
	}
}

func TestCheckpointResponse_IsRunning(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"running: pending", *mockResponse(true, "PENDING", "PENDING"), true},
		{"running: blocked", *mockResponse(true, "BLOCKED", "PENDING"), true},
		{"running: approved", *mockResponse(true, "COMPLETE", "APPROVED"), false},
		{"running: denied", *mockResponse(true, "COMPLETE", "DENIED"), false},
		{"running: failed", *mockResponse(false, "FAILED", "ERROR"), false},
		{"running: undecided", *mockResponse(true, "COMPLETE", "PENDING"), false},
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
		{"allowed: pending", *mockResponse(true, "PENDING", "PENDING"), false},
		{"allowed: blocked", *mockResponse(true, "BLOCKED", "PENDING"), false},
		{"allowed: approved", *mockResponse(true, "COMPLETE", "APPROVED"), true},
		{"allowed: denied", *mockResponse(true, "COMPLETE", "DENIED"), false},
		{"allowed: failed", *mockResponse(false, "FAILED", "ERROR"), false},
		{"allowed: undecided", *mockResponse(true, "COMPLETE", "PENDING"), false},
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
		{"denied: pending", *mockResponse(true, "PENDING", "PENDING"), false},
		{"denied: blocked", *mockResponse(true, "BLOCKED", "PENDING"), false},
		{"denied: approved", *mockResponse(true, "COMPLETE", "APPROVED"), false},
		{"denied: denied", *mockResponse(true, "COMPLETE", "DENIED"), true},
		{"denied: failed", *mockResponse(false, "FAILED", "ERROR"), false},
		{"denied: undecided", *mockResponse(true, "COMPLETE", "PENDING"), false},
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
		{"undecided: pending", *mockResponse(true, "PENDING", "PENDING"), false},
		{"undecided: blocked", *mockResponse(true, "BLOCKED", "PENDING"), false},
		{"undecided: approved", *mockResponse(true, "COMPLETE", "APPROVED"), false},
		{"undecided: denied", *mockResponse(true, "COMPLETE", "DENIED"), false},
		{"undecided: failed", *mockResponse(false, "FAILED", "ERROR"), false},
		{"undecided: undecided", *mockResponse(true, "COMPLETE", "PENDING"), true},
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
		{"error: pending", *mockResponse(true, "PENDING", "PENDING"), false},
		{"error: blocked", *mockResponse(true, "BLOCKED", "PENDING"), false},
		{"error: approved", *mockResponse(true, "COMPLETE", "APPROVED"), false},
		{"error: denied", *mockResponse(true, "COMPLETE", "DENIED"), false},
		{"error: failed", *mockResponse(false, "FAILED", "ERROR"), true},
		{"error: undecided", *mockResponse(true, "COMPLETE", "PENDING"), false},
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
