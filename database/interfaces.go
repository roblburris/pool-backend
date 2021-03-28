package database

// User - 
type User struct {
	Destination string	`json:"destination"`
	Driver bool `json:"driver"`
	ID int64 `json:"id"`
	Name string `json:"name"`
	Opted bool `json:"opted"`
	Partner int64 `json:"partner"`
	TargetCity string `json:"target-city"`
}