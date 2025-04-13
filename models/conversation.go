package models

// Conversation represents a single conversation entry
type Conversation struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
