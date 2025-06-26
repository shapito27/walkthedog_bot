// Package mocks contains mock implementations for testing
package mocks

import (
	"fmt"
	"walkthedog/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
)

// MockTelegramBot implements TelegramBot interface for testing
type MockTelegramBot struct {
	SentMessages []tgbotapi.Chattable
	SendError    error
	UpdatesChan  chan tgbotapi.Update
}

func NewMockTelegramBot() *MockTelegramBot {
	return &MockTelegramBot{
		SentMessages: make([]tgbotapi.Chattable, 0),
		UpdatesChan:  make(chan tgbotapi.Update, 100),
	}
}

func (m *MockTelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if m.SendError != nil {
		return tgbotapi.Message{}, m.SendError
	}

	m.SentMessages = append(m.SentMessages, c)

	// Return a mock message based on the input type
	var chatID int64 = 12345
	var messageID int = len(m.SentMessages)

	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		chatID = msg.ChatID
	} else if poll, ok := c.(tgbotapi.SendPollConfig); ok {
		chatID = poll.ChatID
		// Return a poll message
		return tgbotapi.Message{
			MessageID: messageID,
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
			Poll: &tgbotapi.Poll{
				ID: fmt.Sprintf("poll_%d", messageID),
			},
		}, nil
	}

	return tgbotapi.Message{
		MessageID: messageID,
		Chat: &tgbotapi.Chat{
			ID: chatID,
		},
		Text: "Mock response",
	}, nil
}

func (m *MockTelegramBot) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return tgbotapi.UpdatesChannel(m.UpdatesChan)
}

func (m *MockTelegramBot) GetMe() (tgbotapi.User, error) {
	return tgbotapi.User{
		ID:        123456789,
		UserName:  "MockBot",
		FirstName: "Mock",
	}, nil
}

// Helper method to add updates for testing
func (m *MockTelegramBot) AddUpdate(update tgbotapi.Update) {
	m.UpdatesChan <- update
}

// Helper method to get the number of sent messages
func (m *MockTelegramBot) GetSentMessageCount() int {
	return len(m.SentMessages)
}

// MockGoogleSheetsService implements GoogleSheetsService interface for testing
type MockGoogleSheetsService struct {
	SaveError         error
	CreateSheetError  error
	HasSheetResponse  bool
	SavedTrips        []*models.TripToShelter
	CreatedSheets     []string
	SheetsWithHeaders []string
}

func NewMockGoogleSheetsService() *MockGoogleSheetsService {
	return &MockGoogleSheetsService{
		HasSheetResponse:  true,
		SavedTrips:        make([]*models.TripToShelter, 0),
		CreatedSheets:     make([]string, 0),
		SheetsWithHeaders: make([]string, 0),
	}
}

func (m *MockGoogleSheetsService) SaveTripToShelter(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error) {
	if m.SaveError != nil {
		return nil, m.SaveError
	}

	m.SavedTrips = append(m.SavedTrips, tripToShelter)

	return &sheets.AppendValuesResponse{
		ServerResponse: googleapi.ServerResponse{
			HTTPStatusCode: 200,
		},
	}, nil
}

func (m *MockGoogleSheetsService) SaveTripToShelterSystem(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error) {
	if m.SaveError != nil {
		return nil, m.SaveError
	}

	m.SavedTrips = append(m.SavedTrips, tripToShelter)

	return &sheets.AppendValuesResponse{
		ServerResponse: googleapi.ServerResponse{
			HTTPStatusCode: 200,
		},
	}, nil
}

func (m *MockGoogleSheetsService) CreateSheet(sheetName string) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	if m.CreateSheetError != nil {
		return nil, m.CreateSheetError
	}

	m.CreatedSheets = append(m.CreatedSheets, sheetName)

	return &sheets.BatchUpdateSpreadsheetResponse{}, nil
}

func (m *MockGoogleSheetsService) AddSheetHeaders(sheetName string) (*sheets.AppendValuesResponse, error) {
	if m.SaveError != nil {
		return nil, m.SaveError
	}

	m.SheetsWithHeaders = append(m.SheetsWithHeaders, sheetName)

	return &sheets.AppendValuesResponse{}, nil
}

func (m *MockGoogleSheetsService) HasSheet(sheetName string) bool {
	return m.HasSheetResponse
}

func (m *MockGoogleSheetsService) PrepareSheetForSavingData(sheetName string) error {
	if !m.HasSheet(sheetName) {
		_, err := m.CreateSheet(sheetName)
		if err != nil {
			return err
		}
		_, err = m.AddSheetHeaders(sheetName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper methods for testing
func (m *MockGoogleSheetsService) GetSavedTripsCount() int {
	return len(m.SavedTrips)
}

func (m *MockGoogleSheetsService) GetCreatedSheetsCount() int {
	return len(m.CreatedSheets)
}

func (m *MockGoogleSheetsService) SetSaveError(err error) {
	m.SaveError = err
}

func (m *MockGoogleSheetsService) SetCreateSheetError(err error) {
	m.CreateSheetError = err
}

func (m *MockGoogleSheetsService) SetHasSheetResponse(hasSheet bool) {
	m.HasSheetResponse = hasSheet
}
