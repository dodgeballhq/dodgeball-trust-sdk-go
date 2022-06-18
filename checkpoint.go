package dodgeball

type CheckpointEvent struct {
	IP   string      `json:"ip"`
	Data interface{} `json:"data"`
}

type CheckpointRequest struct {
	CheckpointName    string                    `json:"checkpointName"`
	Event             CheckpointEvent           `json:"event"`
	DodgeballID       string                    `json:"dodgeballId"`
	UserID            string                    `json:"userId"`
	UseVerificationID string                    `json:"useVerificationId"`
	Options           CheckpointResponseOptions `json:"options"`
}

type CheckpointResponseOptions struct {
	Sync    bool   `json:"sync"`
	Timeout int    `json:"timeout"`
	Webhook string `json:"webhook"`
}

type CheckpointResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CheckpointResponse struct {
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

// IsRunning checks to see if the verification is still running
func (cr *CheckpointResponse) IsRunning() bool {
	if cr.Success {
		switch cr.Verification.Status {
		case VerificationStatusPending:
			return true
		case VerificationStatusBlocked:
			return true
		default:
			return false
		}
	}
	return false
}

// IsALlowed checks to see if the verification has been approved
func (cr *CheckpointResponse) IsAllowed() bool {
	if cr.Success && cr.Verification.Status == VerificationStatusComplete {
		switch cr.Verification.Outcome {
		case VerificationOutcomeApproved:
			return true
		default:
			return false
		}
	}
	return false
}

// IsDenied checks to see if the verification has been denied
func (cr *CheckpointResponse) IsDenied() bool {
	if cr.Success {
		switch cr.Verification.Outcome {
		case VerificationOutcomeDenied:
			return true
		default:
			return false
		}
	}
	return false
}

// IsUndecided checks to see if the verification has completed but undecided
func (cr *CheckpointResponse) IsUndecided() bool {
	if cr.Success && cr.Verification.Status == VerificationStatusComplete {
		switch cr.Verification.Outcome {
		case VerificationOutcomePending:
			return true
		default:
			return false
		}
	}
	return false
}

// HasError checks to see if the verification has encountered an error
func (cr *CheckpointResponse) HasError() bool {
	if !cr.Success || len(cr.Errors) > 0 {
		return true
	}
	return false
}

// IsTimeout checks to see if the verification has timed out
func (cr *CheckpointResponse) IsTimeout() bool {
	if !cr.Success && cr.timedOut {
		return true
	}
	return false
}
