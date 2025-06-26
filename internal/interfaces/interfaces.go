// Package interfaces contains interfaces for external dependencies
package interfaces

import (
	"walkthedog/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/sheets/v4"
)

// TelegramBot interface for Telegram Bot API operations
type TelegramBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	GetMe() (tgbotapi.User, error)
}

// GoogleSheetsService interface for Google Sheets operations
type GoogleSheetsService interface {
	SaveTripToShelter(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error)
	SaveTripToShelterSystem(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error)
	CreateSheet(sheetName string) (*sheets.BatchUpdateSpreadsheetResponse, error)
	AddSheetHeaders(sheetName string) (*sheets.AppendValuesResponse, error)
	HasSheet(sheetName string) bool
	PrepareSheetForSavingData(sheetName string) error
}
