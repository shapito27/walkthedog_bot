// Package models contains models are used in project
package models

// Shelter represent shelter information
type Shelter struct {
	ID         string          `yaml:"id"`
	Address    string          `yaml:"address"`
	DonateLink string          `yaml:"donate_link"`
	Title      string          `yaml:"title"`
	Link       string          `yaml:"link"`
	Guide      string          `yaml:"guide"`
	Schedule   ShelterSchedule `yaml:"schedule"`
}

// ShelterSchedule represents trips shedule to shelters
type ShelterSchedule struct {
	Type      string `yaml:"type"`
	Details   []int  `yaml:"details"`
	TimeStart string `yaml:"time_start"`
	TimeEnd   string `yaml:"time_end"`
}

// TripToShelter represents all important information about user's trip to shelter.
type TripToShelter struct {
	Username          string
	Shelter           *Shelter
	Date              string
	IsFirstTrip       bool
	Purpose           []string
	TripBy            string
	HowYouKnowAboutUs string
}

// State represents state of chat with user
type State struct {
	ChatId        int64
	LastMessage   string
	TripToShelter *TripToShelter
}

type TelegramConfig struct {
	APIToken string `yaml:"api_token"`
	Timeout  int    `yaml:"timeout"`
}
