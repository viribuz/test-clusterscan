package main

// Agent represents the structure of our resource
type Agent struct {
	ID        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Token     string `json:"token,omitempty"`
	TokenAge  string `json:"tokenAge,omitempty"`
}

// TimeCardDetail represents the structure of a time card entry
type TimeCardDetail struct {
	ID         int     `json:"id"`
	CampaignId int     `json:"CampaignId"`
	FirmId     int     `json:"FirmId"`
	Days       string  `json:"Days"`
	Regular    float64 `json:"Regular"`
	Holidays   float64 `json:"Holidays"`
	Vacation   float64 `json:"Vacation"`
	Date       string  `json:"Date"`
}

// ScheduleSubmission represents the structure of a schedule submission
type ScheduleSubmission struct {
	PayPeriod       string           `json:"PayPeriod"`
	Year            string           `json:"Year"`
	TimeCardDetails []TimeCardDetail `json:"TimeCardDetails"`
}

// Schedule represents the structure of a schedule
type Schedule struct {
	PayPeriod       string           `json:"PayPeriod"`
	Month           string           `json:"Month"`
	Year            string           `json:"Year"`
	Agent           int              `json:"Agent"`
	TimeCardDetails []TimeCardDetail `json:"TimeCardDetails"`
}
type TimeCardDetails struct {
	ID         int     `json:"id"`
	CampaignId int     `json:"CampaignId"`
	FirmId     int     `json:"FirmId"`
	Days       string  `json:"Days"`
	Regular    float64 `json:"Regular"`
	Holidays   float64 `json:"Holidays"`
	Vacation   float64 `json:"Vacation"`
	Date       string  `json:"Date"`
}
