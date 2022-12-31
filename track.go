package dodgeball

type TrackEvent struct {
	Type      string      `json:"type"`      // The name of the event, may be any string under 256 characters, that indicates what took place
	Data      interface{} `json:"data"`      // Any arbitrary data they want to track. Will be digested into the Dodgeball Vocabulary
	EventTime int64       `json:"eventTime"` // The time the event occurred, in milliseconds since the epoch
}

type TrackOptions struct {
	Event       TrackEvent `json:"event"`
	SourceToken string     `json:"sourceToken"`
	UserID      string     `json:"userId"`
	SessionID   string     `json:"sessionId"`
}
