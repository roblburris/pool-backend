package database

// User - 
type User struct {
	Destination string	`json:"destination,omitempty"`
	Driver bool `json:"driver,omitempty"`
	ID int64 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Opted bool `json:"opted,omitempty"`
	Partner int64 `json:"partner,omitempty"`
	TargetCity string `json:"target-city,omitempty"`
}