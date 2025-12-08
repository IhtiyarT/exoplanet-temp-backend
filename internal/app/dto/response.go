package dto

type SystemListItem struct {
	ID              uint    `json:"id"`
	DateCreated     string  `json:"date_created"`
	Status          string  `json:"status"`
	UserLogin       string  `json:"user_login"`
	ModerID         *uint   `json:"moder_id,omitempty"`
	StarName        string  `json:"star_name"`
	StarType        string  `json:"star_type"`
	StarLuminosity  float64 `json:"star_luminosity"`
	PlanetCount     int     `json:"planet_count"`
	PlanetTempCount int     `json:"planet_temp_count"`
}