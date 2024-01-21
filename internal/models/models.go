// Package models contains models are used in project
package models

// Shelter represent shelter information
type Shelter struct {
	ID          string          `yaml:"id"`
	Address     string          `yaml:"address"`
	DonateLink  string          `yaml:"donate_link"`
	Title       string          `yaml:"title"`
	LongTitle   string          `yaml:"long_title"`
	ShortTitle  string          `yaml:"short_title"`
	Link        string          `yaml:"link"`
	Guide       string          `yaml:"guide"`
	PeopleLimit int32           `yaml:"people_limit"`
	Schedule    ShelterSchedule `yaml:"schedule"`
}

// ShelterSchedule represents trips shedule to shelters
type ShelterSchedule struct {
	Type            string   `yaml:"type"`
	Details         [][]int    `yaml:"details"`
	DatesExceptions []string `yaml:"dates_exceptions"`
	TimeStart       string   `yaml:"time_start"`
	TimeEnd         string   `yaml:"time_end"`
}

// TripToShelter represents all important information about user's trip to shelter.
type TripToShelter struct {
	ID                string
	Username          string
	Shelter           *Shelter
	Date              string
	IsFirstTrip       bool
	Purpose           []string
	TripBy            string
	HowYouKnowAboutUs []string
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
type TelegramEnvironment struct {
	Environment    string                     `yaml:"environment"`
	TelegramConfig map[string]*TelegramConfig `yaml:"environments"`
}
type Administration struct {
	Admin string `yaml:"admin"`
}
type Google struct {
	SpreadsheetID string `yaml:"spreadsheet_id"`
}
type ConfigFile struct {
	TelegramEnvironment *TelegramEnvironment `yaml:"telegram"`
	Administration      *Administration      `yaml:"administration"`
	Google              *Google              `yaml:"google"`
}
