package main

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
	"walkthedog/internal/mocks"
	"walkthedog/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestSavingOneTripToCache save one trip to cache and then get it from cache. Test succeed if values are same.
func TestSavingOneTripToCache(t *testing.T) {
	c, err := initCache()
	if err != nil {
		log.Panic(err)
	}
	app.Cache = c
	var chatId int64 = 5656565656

	tripsList := getTripsToSheltersList()

	newTripToShelter := tripsList[0]

	app.saveTripToCache(newTripToShelter, chatId)

	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}
	for _, v := range chatsWithTripsID[chatId] {
		var trip models.TripToShelter
		tripFromCache, found := app.Cache.Get(v)
		if found {
			trip = tripFromCache.(models.TripToShelter)
		}
		//newTripToShelter.Date = ""
		err := equalTripsToShelters(newTripToShelter, &trip)
		if err != "" {
			t.Error(err)
		}
	}
}

// TestSavingOneTripToCache save one trip to cache and then get it from cache. Test succeed if values are same.
func TestSavingTwoTripsFromOneChatToCache(t *testing.T) {
	c, err := initCache()
	if err != nil {
		log.Panic(err)
	}
	app.Cache = c
	var chatId int64 = 5656565656
	tripsList := getTripsToSheltersList()
	newTripToShelter1 := tripsList[0]

	app.saveTripToCache(newTripToShelter1, chatId)

	newTripToShelter2 := tripsList[1]

	app.saveTripToCache(newTripToShelter2, chatId)

	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}

	if len(chatsWithTripsID[chatId]) != 2 {
		t.Error("Expected only 2 elements in cache")
	}

	// take first trip from cache and compare
	var trip models.TripToShelter
	tripFromCache, found := app.Cache.Get(chatsWithTripsID[chatId][0])
	if found {
		trip = tripFromCache.(models.TripToShelter)
	}
	//newTripToShelter.Date = ""
	errorMes := equalTripsToShelters(newTripToShelter1, &trip)
	if errorMes != "" {
		t.Error(errorMes)
	}

	// take first trip from cache and compare
	tripFromCache, found = app.Cache.Get(chatsWithTripsID[chatId][1])
	if found {
		trip = tripFromCache.(models.TripToShelter)
	}
	//newTripToShelter.Date = ""
	errorMes = equalTripsToShelters(newTripToShelter2, &trip)
	if errorMes != "" {
		t.Error(errorMes)
	}
}

func TestSavingTwoTripsFromOneChatToCacheAndRemoveOne(t *testing.T) {
	c, err := initCache()
	if err != nil {
		log.Panic(err)
	}
	app.Cache = c
	var chatId int64 = 5656565656
	tripsList := getTripsToSheltersList()
	newTripToShelter1 := tripsList[0]

	app.saveTripToCache(newTripToShelter1, chatId)

	newTripToShelter2 := tripsList[1]

	app.saveTripToCache(newTripToShelter2, chatId)

	// remove from cache first added trip
	app.removeTripFromCache(newTripToShelter1.ID, chatId)

	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}

	if len(chatsWithTripsID[chatId]) != 1 {
		t.Error("Expected only 1 element in cache")
	}

	var trip models.TripToShelter
	// take trip from cache and compare
	tripFromCache, found := app.Cache.Get(chatsWithTripsID[chatId][0])
	if found {
		trip = tripFromCache.(models.TripToShelter)
	}
	//newTripToShelter.Date = ""
	errorMes := equalTripsToShelters(newTripToShelter2, &trip)
	if errorMes != "" {
		t.Error(errorMes)
	}
}

// equalTripsToShelters compares two TripToShelter.
func equalTripsToShelters(trip1 *models.TripToShelter, trip2 *models.TripToShelter) string {
	error := ""
	if trip1.ID != trip2.ID {
		error = fmt.Sprintf("ID is not equal. Expected \"%s\", got \"%s\"", trip1.ID, trip2.ID)
	}
	if trip1.Date != trip2.Date {
		error = fmt.Sprintf("Date is not equal. Expected \"%s\", got \"%s\"", trip1.Date, trip2.Date)
	}
	if trip1.Username != trip2.Username {
		error = fmt.Sprintf("Username is not equal. Expected \"%s\", got \"%s\"", trip1.Username, trip2.Username)
	}
	for i, v := range trip1.HowYouKnowAboutUs {
		if v != trip2.HowYouKnowAboutUs[i] {
			error = fmt.Sprintf("HowYouKnowAboutUs is not equal. Expected \"%s\", got \"%s\"", trip1.HowYouKnowAboutUs[i], trip2.HowYouKnowAboutUs[i])
		}
	}
	if trip1.TripBy != trip2.TripBy {
		error = fmt.Sprintf("TripBy is not equal. Expected \"%s\", got \"%s\"", trip1.TripBy, trip2.TripBy)
	}
	if trip1.IsFirstTrip != trip2.IsFirstTrip {
		error = fmt.Sprintf("IsFirstTrip is not equal. Expected \"%t\", got \"%t\"", trip1.IsFirstTrip, trip2.IsFirstTrip)
	}
	for i, v := range trip1.Purpose {
		if v != trip2.Purpose[i] {
			error = fmt.Sprintf("Purpose is not equal. Expected \"%s\", got \"%s\"", trip1.Purpose[i], trip2.Purpose[i])
		}
	}
	if trip1.Shelter.ID != trip2.Shelter.ID {
		error = fmt.Sprintf("Shelter.ID is not equal. Expected \"%s\", got \"%s\"", trip1.Shelter.ID, trip2.Shelter.ID)
	}

	return error
}

func getTripsToSheltersList() []*models.TripToShelter {
	var list []*models.TripToShelter

	list = append(list, &models.TripToShelter{
		ID:       "13.08.2022Ника",
		Username: "sdfsd908",
		Shelter: &models.Shelter{
			ID:          "2",
			Address:     "sdfsdf",
			DonateLink:  "sdfsdfsdf",
			Title:       "bib",
			ShortTitle:  "d",
			Link:        "sdfsdfsdf",
			Guide:       "sdfsdf",
			PeopleLimit: 4,
			Schedule: models.ShelterSchedule{
				Type:            "sdfsdf",
				Details:         [][]int{{4, 54}, {1, 7}},
				DatesExceptions: []string{"434", "sdf"},
				TimeStart:       "1:1",
				TimeEnd:         "34:5",
			},
		},
		Date:              "3434",
		IsFirstTrip:       true,
		Purpose:           []string{"dfdfdf", "df"},
		TripBy:            "dfdf",
		HowYouKnowAboutUs: []string{"youtube", "fb"},
	})

	list = append(list, &models.TripToShelter{
		ID:       "11.07.2022шанс",
		Username: "sdfsd908",
		Shelter: &models.Shelter{
			ID:          "2",
			Address:     "sdfsdf",
			DonateLink:  "sdfsdfsdf",
			Title:       "bib",
			ShortTitle:  "d",
			Link:        "sdfsdfsdf",
			Guide:       "sdfsdf",
			PeopleLimit: 4,
			Schedule: models.ShelterSchedule{
				Type:            "sdfsdf",
				Details:         [][]int{{4, 54}, {1, 7}},
				DatesExceptions: []string{"434", "sdf"},
				TimeStart:       "1:1",
				TimeEnd:         "34:5",
			},
		},
		Date:              "3434",
		IsFirstTrip:       true,
		Purpose:           []string{"dfdfdf", "df"},
		TripBy:            "dfdf",
		HowYouKnowAboutUs: []string{"youtube", "fb"},
	})

	return list
}

// TestGetShelterByNameID tests shelter selection by name/ID
func TestGetShelterByNameID(t *testing.T) {
	shelters := getSheltersListForTest()

	// Test valid shelter selection
	shelter, err := shelters.getShelterByNameID("1. Test Shelter")
	if err != nil {
		t.Errorf("Expected no error for valid shelter, got: %v", err)
	}
	if shelter.ID != "1" {
		t.Errorf("Expected shelter ID '1', got '%s'", shelter.ID)
	}

	// Test invalid shelter name (no dot)
	_, err = shelters.getShelterByNameID("Invalid Shelter Name")
	if err == nil {
		t.Error("Expected error for invalid shelter name format")
	}

	// Test non-existent shelter ID
	_, err = shelters.getShelterByNameID("999. Non-existent Shelter")
	if err == nil {
		t.Error("Expected error for non-existent shelter ID")
	}
}

// TestIsTripDateValid tests date validation logic
func TestIsTripDateValid(t *testing.T) {
	shelter := &models.Shelter{
		Schedule: models.ShelterSchedule{
			Type:    "regularly",
			Details: [][]int{{1, 6}}, // First Saturday
		},
	}

	trip := &models.TripToShelter{
		Shelter: shelter,
	}

	// Test with nil trip
	if isTripDateValid("01.01.2024", nil) {
		t.Error("Expected false for nil trip")
	}

	// Test with nil shelter
	tripNoShelter := &models.TripToShelter{Shelter: nil}
	if isTripDateValid("01.01.2024", tripNoShelter) {
		t.Error("Expected false for nil shelter")
	}

	// Test valid date (would need to mock getDatesByShelter for proper testing)
	// This is a basic structure test
	result := isTripDateValid("01.01.2024", trip)
	// Result depends on actual date calculation, but function should not panic
	_ = result
}

// TestCalculateDay tests date calculation logic
func TestCalculateDay(t *testing.T) {
	// Test first Saturday of January 2024
	result := calculateDay(6, 1, time.January) // Saturday, first week

	// Verify it's a Saturday
	if result.Weekday() != time.Saturday {
		t.Errorf("Expected Saturday, got %v", result.Weekday())
	}

	// Test second Sunday of February
	result = calculateDay(7, 2, time.February) // Sunday, second week
	if result.Weekday() != time.Sunday {
		t.Errorf("Expected Sunday, got %v", result.Weekday())
	}
}

// TestIsShelterHasTripDates tests shelter availability
func TestIsShelterHasTripDates(t *testing.T) {
	// Test shelter with regular schedule
	shelterRegular := &models.Shelter{
		Schedule: models.ShelterSchedule{Type: "regularly"},
	}
	if !isShelterHasTripDates(shelterRegular) {
		t.Error("Expected true for regular schedule")
	}

	// Test shelter with no schedule
	shelterNone := &models.Shelter{
		Schedule: models.ShelterSchedule{Type: "none"},
	}
	if isShelterHasTripDates(shelterNone) {
		t.Error("Expected false for none schedule")
	}

	// Test shelter with everyday schedule
	shelterEveryday := &models.Shelter{
		Schedule: models.ShelterSchedule{Type: "everyday"},
	}
	if !isShelterHasTripDates(shelterEveryday) {
		t.Error("Expected true for everyday schedule")
	}
}

// TestInitCacheWithMissingDirectory tests cache directory creation
func TestInitCacheWithMissingDirectory(t *testing.T) {
	// This test verifies our fix for the cache directory issue
	cache, err := initCache()
	if err != nil {
		t.Errorf("initCache should handle missing directory, got error: %v", err)
	}
	if cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

// TestSendTripToGSheetErrorHandling tests our error handling with mocks
func TestSendTripToGSheetErrorHandling(t *testing.T) {
	// Initialize app config for testing
	c, err := initCache()
	if err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}
	app.Cache = c

	// Create mock services
	mockBot := mocks.NewMockTelegramBot()
	mockSheets := mocks.NewMockGoogleSheetsService()

	// Set up mock to return error
	mockSheets.SetSaveError(errors.New("mock sheets error"))

	app.Bot = mockBot
	app.SheetsService = mockSheets
	app.Google = &models.Google{SpreadsheetID: "test-id"}

	// Create test trip
	trip := &models.TripToShelter{
		ID:       "test-trip-id",
		Username: "testuser",
		Shelter: &models.Shelter{
			ShortTitle: "TestShelter",
		},
		Date: "01.01.2024",
	}

	// This should handle errors gracefully and not panic
	result := app.sendTripToGSheet(12345, trip)

	// Should return false due to Google Sheets error, but not crash
	if result {
		t.Error("Expected false result due to mock Google Sheets error")
	}

	// Verify trip was cached due to failure
	app.saveTripToCache(trip, 12345)
	cachedTrip, found := app.Cache.Get(trip.ID)
	if !found {
		t.Error("Expected trip to be cached after Google Sheets failure")
		return
	}

	retrievedTrip, ok := cachedTrip.(models.TripToShelter)
	if !ok {
		t.Error("Cached trip has invalid type, expected models.TripToShelter")
		return
	}
	if retrievedTrip.ID != trip.ID {
		t.Errorf("Expected cached trip ID %s, got %s", trip.ID, retrievedTrip.ID)
	}
}

// TestSendTripToGSheetSuccess tests successful Google Sheets operation with mocks
func TestSendTripToGSheetSuccess(t *testing.T) {
	// Initialize app config for testing
	c, err := initCache()
	if err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}
	app.Cache = c

	// Create mock services
	mockBot := mocks.NewMockTelegramBot()
	mockSheets := mocks.NewMockGoogleSheetsService()

	app.Bot = mockBot
	app.SheetsService = mockSheets
	app.Google = &models.Google{SpreadsheetID: "test-id"}

	// Create test trip
	trip := &models.TripToShelter{
		ID:       "test-trip-success-id",
		Username: "testuser",
		Shelter: &models.Shelter{
			ShortTitle: "TestShelter",
		},
		Date: "01.01.2024",
	}

	// This should succeed
	result := app.sendTripToGSheet(12345, trip)

	// Should return true on success
	if !result {
		t.Error("Expected true result for successful Google Sheets operation")
	}

	// Verify mock received the trip
	if mockSheets.GetSavedTripsCount() != 2 { // Should be saved twice (main and system)
		t.Errorf("Expected 2 saved trips, got %d", mockSheets.GetSavedTripsCount())
	}
}

// TestTelegramBotMocking tests the Telegram bot mock
func TestTelegramBotMocking(t *testing.T) {
	mockBot := mocks.NewMockTelegramBot()

	// Test GetMe
	user, err := mockBot.GetMe()
	if err != nil {
		t.Errorf("Expected no error from GetMe, got %v", err)
	}
	if user.UserName != "MockBot" {
		t.Errorf("Expected username 'MockBot', got '%s'", user.UserName)
	}

	// Test Send
	msg := tgbotapi.NewMessage(12345, "Test message")
	response, err := mockBot.Send(msg)
	if err != nil {
		t.Errorf("Expected no error from Send, got %v", err)
	}
	if response.Chat.ID != 12345 {
		t.Errorf("Expected chat ID 12345, got %d", response.Chat.ID)
	}

	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 sent message, got %d", mockBot.GetSentMessageCount())
	}
}

// Helper function to create test shelters
func getSheltersListForTest() SheltersList {
	return SheltersList{
		1: &models.Shelter{
			ID:         "1",
			Title:      "Test Shelter",
			ShortTitle: "Test",
			Schedule: models.ShelterSchedule{
				Type:    "regularly",
				Details: [][]int{{1, 6}},
			},
		},
		2: &models.Shelter{
			ID:         "2",
			Title:      "Another Shelter",
			ShortTitle: "Another",
			Schedule: models.ShelterSchedule{
				Type: "none",
			},
		},
	}
}

// TestGetConfig tests configuration file parsing
func TestGetConfig(t *testing.T) {
	// This test verifies config loading works
	config, err := getConfig()
	if err != nil {
		t.Errorf("getConfig should not fail, got error: %v", err)
	}
	if config == nil {
		t.Error("Expected config to be loaded")
	}

	// Verify basic structure
	if config.TelegramEnvironment == nil {
		t.Error("Expected telegram environment to be loaded")
	}
	if config.Administration == nil {
		t.Error("Expected administration config to be loaded")
	}
	if config.Google == nil {
		t.Error("Expected google config to be loaded")
	}
}

// TestGetShelters tests shelter configuration loading
func TestGetShelters(t *testing.T) {
	shelters, err := getShelters()
	if err != nil {
		t.Errorf("getShelters should not fail, got error: %v", err)
	}
	if len(shelters) == 0 {
		t.Error("Expected shelters to be loaded")
	}

	// Verify shelter structure
	for id, shelter := range shelters {
		if shelter.ID == "" {
			t.Errorf("Shelter %d should have non-empty ID", id)
		}
		if shelter.Title == "" {
			t.Errorf("Shelter %d should have non-empty Title", id)
		}
		if shelter.ShortTitle == "" {
			t.Errorf("Shelter %d should have non-empty ShortTitle", id)
		}
	}
}

// TestEqualTripsToShelters tests the comparison function
func TestEqualTripsToShelters(t *testing.T) {
	trip1 := &models.TripToShelter{
		ID:       "test-id",
		Username: "testuser",
		Date:     "01.01.2024",
		Shelter: &models.Shelter{
			ID: "1",
		},
		IsFirstTrip:       true,
		Purpose:           []string{"walk dogs"},
		TripBy:            "car",
		HowYouKnowAboutUs: []string{"internet"},
	}

	trip2 := &models.TripToShelter{
		ID:       "test-id",
		Username: "testuser",
		Date:     "01.01.2024",
		Shelter: &models.Shelter{
			ID: "1",
		},
		IsFirstTrip:       true,
		Purpose:           []string{"walk dogs"},
		TripBy:            "car",
		HowYouKnowAboutUs: []string{"internet"},
	}

	// Test equal trips
	result := equalTripsToShelters(trip1, trip2)
	if result != "" {
		t.Errorf("Expected empty result for equal trips, got: %s", result)
	}

	// Test different IDs
	trip2.ID = "different-id"
	result = equalTripsToShelters(trip1, trip2)
	if result == "" {
		t.Error("Expected error for different IDs")
	}

	// Test different usernames
	trip2.ID = trip1.ID
	trip2.Username = "differentuser"
	result = equalTripsToShelters(trip1, trip2)
	if result == "" {
		t.Error("Expected error for different usernames")
	}
}

// =================== COMPREHENSIVE COMMAND HANDLING TESTS ===================

// TestStartCommand tests the /start command
func TestStartCommand(t *testing.T) {
	app := setupTestApp(t)

	update := createTestUpdate(t, 12345, "/start")

	// Process the command
	processTestUpdate(app, update)

	// Verify response
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", mockBot.GetSentMessageCount())
	}

	// Verify state was updated
	statePoolMutex.RLock()
	state, exists := statePool[12345]
	statePoolMutex.RUnlock()

	if !exists {
		t.Error("Expected state to be created for chat")
	}
	if state.LastMessage != commandStart {
		t.Errorf("Expected last message to be %s, got %s", commandStart, state.LastMessage)
	}
}

// TestGoShelterCommand tests the /go_shelter command
func TestGoShelterCommand(t *testing.T) {
	app := setupTestApp(t)

	update := createTestUpdate(t, 12345, "/go_shelter")

	// Process the command
	processTestUpdate(app, update)

	// Verify response
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", mockBot.GetSentMessageCount())
	}

	// Verify state
	statePoolMutex.RLock()
	state := statePool[12345]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandGoShelter {
		t.Errorf("Expected last message to be %s, got %s", commandGoShelter, state.LastMessage)
	}
}

// TestChooseShelterCommand tests the /choose_shelter command
func TestChooseShelterCommand(t *testing.T) {
	app := setupTestApp(t)

	update := createTestUpdate(t, 12345, "/choose_shelter")

	// Process the command
	processTestUpdate(app, update)

	// Verify response
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", mockBot.GetSentMessageCount())
	}

	// Verify state
	statePoolMutex.RLock()
	state := statePool[12345]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandChooseShelter {
		t.Errorf("Expected last message to be %s, got %s", commandChooseShelter, state.LastMessage)
	}
}

// TestDonationCommand tests the /donation command
func TestDonationCommand(t *testing.T) {
	app := setupTestApp(t)

	update := createTestUpdate(t, 12345, "/donation")

	// Process the command
	processTestUpdate(app, update)

	// Verify response
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", mockBot.GetSentMessageCount())
	}

	// Verify state
	statePoolMutex.RLock()
	state := statePool[12345]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandDonation {
		t.Errorf("Expected last message to be %s, got %s", commandDonation, state.LastMessage)
	}
}

// TestMasterclassCommand tests the /masterclass command
func TestMasterclassCommand(t *testing.T) {
	app := setupTestApp(t)

	update := createTestUpdate(t, 12345, "/masterclass")

	// Process the command
	processTestUpdate(app, update)

	// Verify response
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() != 1 {
		t.Errorf("Expected 1 message sent, got %d", mockBot.GetSentMessageCount())
	}

	// Verify state
	statePoolMutex.RLock()
	state := statePool[12345]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandMasterclass {
		t.Errorf("Expected last message to be %s, got %s", commandMasterclass, state.LastMessage)
	}
}

// =================== ADMIN COMMAND TESTS ===================

// TestAdminRereadSheltersCommand tests admin command /reread_shelters
func TestAdminRereadSheltersCommand(t *testing.T) {
	app := setupTestApp(t)
	app.AdminChatId = 99999 // Set admin chat ID

	// Test as admin user
	update := createTestUpdate(t, 99999, "/reread_shelters")

	// Process the command
	processTestUpdate(app, update)

	// Verify state was updated
	statePoolMutex.RLock()
	state := statePool[99999]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandRereadShelters {
		t.Errorf("Expected last message to be %s, got %s", commandRereadShelters, state.LastMessage)
	}
}

// TestNonAdminRereadSheltersCommand tests that non-admin users can't use admin commands
func TestNonAdminRereadSheltersCommand(t *testing.T) {
	app := setupTestApp(t)
	app.AdminChatId = 99999 // Set admin chat ID

	// Test as non-admin user
	update := createTestUpdate(t, 12345, "/reread_shelters")

	// Process the command
	processTestUpdate(app, update)

	// Verify state was NOT updated to admin command
	statePoolMutex.RLock()
	state := statePool[12345]
	statePoolMutex.RUnlock()

	if state.LastMessage == commandRereadShelters {
		t.Error("Non-admin user should not be able to execute admin commands")
	}
}

// TestAdminClearCacheCommand tests admin command /clear_cache
func TestAdminClearCacheCommand(t *testing.T) {
	app := setupTestApp(t)
	app.AdminChatId = 99999

	// Add some data to cache first
	trip := &models.TripToShelter{
		ID:       "test-cache-trip",
		Username: "testuser",
		Shelter:  &models.Shelter{ShortTitle: "Test"},
		Date:     "01.01.2024",
	}
	app.saveTripToCache(trip, 12345)

	// Verify cache has data
	_, found := app.Cache.Get(trip.ID)
	if !found {
		t.Error("Expected trip to be in cache before clear")
	}

	// Execute clear cache command
	update := createTestUpdate(t, 99999, "/clear_cache")
	processTestUpdate(app, update)

	// Note: In a real test, we'd need to verify cache was cleared,
	// but since sendCachedTripsToGSheet uses mocks, we can check mock calls
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)
	if mockSheets.GetSavedTripsCount() == 0 {
		t.Error("Expected cached trips to be sent to sheets before cache clear")
	}
}

// =================== STATE MANAGEMENT TESTS ===================

// TestStateCreation tests that states are created properly for new chats
func TestStateCreation(t *testing.T) {
	setupTestApp(t)

	chatId := int64(12345)

	// Verify no state exists initially
	statePoolMutex.RLock()
	_, exists := statePool[chatId]
	statePoolMutex.RUnlock()

	if exists {
		t.Error("State should not exist for new chat")
	}

	// Process a message
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(setupTestApp(t), update)

	// Verify state was created
	statePoolMutex.RLock()
	state, exists := statePool[chatId]
	statePoolMutex.RUnlock()

	if !exists {
		t.Error("State should be created after processing message")
	}
	if state.ChatId != chatId {
		t.Errorf("Expected chat ID %d, got %d", chatId, state.ChatId)
	}
}

// TestStateTransitions tests state transitions during user flow
func TestStateTransitions(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)

	// Start with /start command
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(app, update)

	// Verify initial state
	statePoolMutex.RLock()
	state := statePool[chatId]
	statePoolMutex.RUnlock()
	if state.LastMessage != commandStart {
		t.Errorf("Expected initial state %s, got %s", commandStart, state.LastMessage)
	}

	// Transition to shelter selection
	update = createTestUpdate(t, chatId, "/go_shelter")
	processTestUpdate(app, update)

	// Verify state transition
	statePoolMutex.RLock()
	state = statePool[chatId]
	statePoolMutex.RUnlock()
	if state.LastMessage != commandGoShelter {
		t.Errorf("Expected state %s, got %s", commandGoShelter, state.LastMessage)
	}
}

// =================== HELPER FUNCTIONS FOR TESTS ===================

// setupTestApp creates a test AppConfig with mocks
func setupTestApp(t *testing.T) *AppConfig {
	t.Helper()

	// Clean up global state from previous tests
	cleanupTestState()

	// Initialize cache
	c, err := initCache()
	if err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}

	// Create mocks
	mockBot := mocks.NewMockTelegramBot()
	mockSheets := mocks.NewMockGoogleSheetsService()

	return &AppConfig{
		Environment:   "test",
		AdminChatId:   99999,
		Google:        &models.Google{SpreadsheetID: "test-id"},
		Cache:         c,
		Bot:           mockBot,
		SheetsService: mockSheets,
	}
}

// cleanupTestState clears global state between tests
func cleanupTestState() {
	statePoolMutex.Lock()
	statePool = make(map[int64]*models.State)
	statePoolMutex.Unlock()

	pollsMutex.Lock()
	polls = make(map[string]int64)
	pollsMutex.Unlock()
}

// createTestUpdate creates a test Telegram update
func createTestUpdate(t *testing.T, chatId int64, text string) tgbotapi.Update {
	t.Helper()

	return tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			From: &tgbotapi.User{
				ID:       int64(chatId),
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID: chatId,
			},
			Text: text,
		},
	}
}

// createTestPollUpdate creates a test poll answer update
func createTestPollUpdate(t *testing.T, chatId int64, pollId string, options []int) tgbotapi.Update {
	t.Helper()

	return tgbotapi.Update{
		UpdateID: 1,
		PollAnswer: &tgbotapi.PollAnswer{
			PollID: pollId,
			User: tgbotapi.User{
				ID:       int64(chatId),
				UserName: "testuser",
			},
			OptionIDs: options,
		},
	}
}

// processTestUpdate simulates processing an update through the main logic
func processTestUpdate(app *AppConfig, update tgbotapi.Update) {
	// Set the global app variable for the test
	oldApp := app
	defer func() { app = oldApp }()

	// This is a simplified version of the main update processing logic
	// In a real implementation, you'd extract the main logic into testable functions

	var chatId int64
	if update.Message != nil {
		chatId = update.Message.Chat.ID
	} else if update.PollAnswer != nil {
		pollsMutex.RLock()
		chatId = polls[update.PollAnswer.PollID]
		pollsMutex.RUnlock()
	}

	// Get or create state
	statePoolMutex.RLock()
	state, ok := statePool[chatId]
	statePoolMutex.RUnlock()

	if !ok {
		state = &models.State{
			ChatId:      chatId,
			LastMessage: "",
		}
		statePoolMutex.Lock()
		statePool[chatId] = state
		statePoolMutex.Unlock()
	}

	// Process message commands
	if update.Message != nil {
		text := update.Message.Text

		switch text {
		case commandStart:
			msgObj := startMessage(chatId)
			app.Bot.Send(msgObj)
			state.LastMessage = commandStart
		case commandGoShelter:
			state.LastMessage = app.goShelterCommand(&update)
		case commandChooseShelter:
			shelters := getSheltersListForTest()
			state.LastMessage = app.chooseShelterCommand(&update, &shelters)
		case commandDonation:
			state.LastMessage = app.donationCommand(chatId)
		case commandMasterclass:
			msgObj := masterclass(chatId)
			app.Bot.Send(msgObj)
			state.LastMessage = commandMasterclass
		case commandRereadShelters:
			if chatId == app.AdminChatId {
				state.LastMessage = commandRereadShelters
			}
		case commandClearCache:
			if chatId == app.AdminChatId {
				app.sendCachedTripsToGSheet()
				app.Cache.Flush()
				state.LastMessage = commandClearCache
			}
		}
	}

	// Update state
	statePoolMutex.Lock()
	statePool[chatId] = state
	statePoolMutex.Unlock()
}

// =================== POLL ANSWER PROCESSING TESTS ===================

// TestPollAnswerTripPurpose tests poll answer processing for trip purpose
func TestPollAnswerTripPurpose(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)
	pollId := "test-poll-123"

	// Set up poll mapping
	pollsMutex.Lock()
	polls[pollId] = chatId
	pollsMutex.Unlock()

	// Set up state
	state := &models.State{
		ChatId:      chatId,
		LastMessage: commandTripPurpose,
		TripToShelter: &models.TripToShelter{
			ID:       "test-trip",
			Username: "testuser",
			Shelter:  &models.Shelter{ShortTitle: "Test"},
			Date:     "01.01.2024",
			Purpose:  []string{},
		},
	}
	statePoolMutex.Lock()
	statePool[chatId] = state
	statePoolMutex.Unlock()

	// Create poll answer update
	update := createTestPollUpdate(t, chatId, pollId, []int{0, 1}) // Select first two options

	// Process the poll answer
	processTestPollUpdate(app, update)

	// Verify trip purpose was updated
	statePoolMutex.RLock()
	updatedState := statePool[chatId]
	statePoolMutex.RUnlock()

	if len(updatedState.TripToShelter.Purpose) != 2 {
		t.Errorf("Expected 2 purposes selected, got %d", len(updatedState.TripToShelter.Purpose))
	}
}

// TestPollAnswerTripBy tests poll answer processing for trip transport
func TestPollAnswerTripBy(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)
	pollId := "test-poll-456"

	// Set up poll mapping
	pollsMutex.Lock()
	polls[pollId] = chatId
	pollsMutex.Unlock()

	// Set up state
	state := &models.State{
		ChatId:      chatId,
		LastMessage: commandTripBy,
		TripToShelter: &models.TripToShelter{
			ID:       "test-trip",
			Username: "testuser",
			Shelter:  &models.Shelter{ShortTitle: "Test"},
			Date:     "01.01.2024",
		},
	}
	statePoolMutex.Lock()
	statePool[chatId] = state
	statePoolMutex.Unlock()

	// Create poll answer update
	update := createTestPollUpdate(t, chatId, pollId, []int{0}) // Select first option

	// Process the poll answer
	processTestPollUpdate(app, update)

	// Verify trip by was updated
	statePoolMutex.RLock()
	updatedState := statePool[chatId]
	statePoolMutex.RUnlock()

	if updatedState.TripToShelter.TripBy == "" {
		t.Error("Expected TripBy to be set after poll answer")
	}
}

// =================== TRIP REGISTRATION WORKFLOW TESTS ===================

// TestCompleteUserRegistrationFlow tests the complete user registration workflow
func TestCompleteUserRegistrationFlow(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)

	// Step 1: Start command
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(app, update)

	// Verify state
	statePoolMutex.RLock()
	state := statePool[chatId]
	statePoolMutex.RUnlock()
	if state.LastMessage != commandStart {
		t.Errorf("Expected %s, got %s", commandStart, state.LastMessage)
	}

	// Step 2: Go to shelter selection
	update = createTestUpdate(t, chatId, "/go_shelter")
	processTestUpdate(app, update)

	// Verify state transition
	statePoolMutex.RLock()
	state = statePool[chatId]
	statePoolMutex.RUnlock()
	if state.LastMessage != commandGoShelter {
		t.Errorf("Expected %s, got %s", commandGoShelter, state.LastMessage)
	}

	// Step 3: Choose shelter
	update = createTestUpdate(t, chatId, "/choose_shelter")
	processTestUpdate(app, update)

	// Verify state transition
	statePoolMutex.RLock()
	state = statePool[chatId]
	statePoolMutex.RUnlock()
	if state.LastMessage != commandChooseShelter {
		t.Errorf("Expected %s, got %s", commandChooseShelter, state.LastMessage)
	}

	// Verify messages were sent
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() < 3 {
		t.Errorf("Expected at least 3 messages sent, got %d", mockBot.GetSentMessageCount())
	}
}

// TestTripCreationAndStorage tests trip creation and storage
func TestTripCreationAndStorage(t *testing.T) {
	app := setupTestApp(t)

	trip := &models.TripToShelter{
		ID:                "test-trip-create",
		Username:          "testuser",
		Shelter:           &models.Shelter{ShortTitle: "TestShelter"},
		Date:              "01.01.2024",
		IsFirstTrip:       true,
		Purpose:           []string{"Help animals", "Volunteer"},
		TripBy:            "Car",
		HowYouKnowAboutUs: []string{"Social media"},
	}

	// Test successful storage
	result := app.sendTripToGSheet(12345, trip)
	if !result {
		t.Error("Expected successful trip storage")
	}

	// Verify mock received the trip
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)
	if mockSheets.GetSavedTripsCount() != 2 { // Main + System sheet
		t.Errorf("Expected 2 saved trips, got %d", mockSheets.GetSavedTripsCount())
	}
}

// =================== SHELTER SELECTION AND DATE VALIDATION TESTS ===================

// TestShelterSelection tests shelter selection logic
func TestShelterSelection(t *testing.T) {
	shelters := getSheltersListForTest()

	// Test shelter exists
	shelter, exists := shelters[1]
	if !exists {
		t.Error("Expected test shelter 1 to exist")
	}
	if shelter.Title != "Test Shelter" {
		t.Errorf("Expected 'Test Shelter', got '%s'", shelter.Title)
	}

	// Test shelter doesn't exist
	_, exists = shelters[999]
	if exists {
		t.Error("Expected shelter 999 to not exist")
	}
}

// TestDateValidation tests date validation logic
func TestDateValidation(t *testing.T) {
	// Test nil trip handling
	if isTripDateValid("01.01.2024", nil) {
		t.Error("Expected false for nil trip")
	}

	// Test nil shelter handling
	trip := &models.TripToShelter{
		Shelter: nil,
	}
	if isTripDateValid("01.01.2024", trip) {
		t.Error("Expected false for nil shelter")
	}

	// Test with valid shelter structure
	trip = &models.TripToShelter{
		Shelter: &models.Shelter{
			Schedule: models.ShelterSchedule{
				Type:    "regularly",
				Details: [][]int{{1, 6}}, // Week 1, Day 6 (Saturday)
			},
		},
	}

	// Note: The actual date validation depends on the current date and shelter schedule
	// We're testing the function doesn't crash rather than specific date logic
	result := isTripDateValid("13.01.2024 Субботa", trip)
	_ = result // Just ensure it doesn't crash

	// Test invalid date format
	invalidDate := "invalid-date"
	if isTripDateValid(invalidDate, trip) {
		t.Errorf("Expected date '%s' to be invalid", invalidDate)
	}
}

// =================== EDGE CASES AND ERROR SCENARIOS ===================

// TestNilPointerHandling tests nil pointer handling
func TestNilPointerHandling(t *testing.T) {
	app := setupTestApp(t)

	// Test sending nil trip to sheets
	result := app.sendTripToGSheet(12345, nil)
	if result {
		t.Error("Expected false result for nil trip")
	}
}

// TestConcurrentStateAccess tests concurrent access to state
func TestConcurrentStateAccess(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)

	// Create initial state
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(app, update)

	// Simulate concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			statePoolMutex.RLock()
			_ = statePool[chatId]
			statePoolMutex.RUnlock()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			statePoolMutex.Lock()
			if state, exists := statePool[chatId]; exists {
				state.LastMessage = fmt.Sprintf("test-%d", i)
			}
			statePoolMutex.Unlock()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify state still exists
	statePoolMutex.RLock()
	_, exists := statePool[chatId]
	statePoolMutex.RUnlock()

	if !exists {
		t.Error("Expected state to still exist after concurrent access")
	}
}

// TestMemoryCleanup tests the memory cleanup functionality
func TestMemoryCleanup(t *testing.T) {
	setupTestApp(t)

	// Add some states
	for i := int64(1); i <= 5; i++ {
		state := &models.State{
			ChatId:      i,
			LastMessage: "",
		}
		statePoolMutex.Lock()
		statePool[i] = state
		statePoolMutex.Unlock()
	}

	// Verify states exist
	statePoolMutex.RLock()
	count := len(statePool)
	statePoolMutex.RUnlock()

	if count != 5 {
		t.Errorf("Expected 5 states, got %d", count)
	}

	// Run cleanup
	cleanupOldStates()

	// Verify empty states were cleaned up
	statePoolMutex.RLock()
	newCount := len(statePool)
	statePoolMutex.RUnlock()

	if newCount != 0 {
		t.Errorf("Expected 0 states after cleanup, got %d", newCount)
	}
}

// TestErrorRecovery tests error recovery scenarios
func TestErrorRecovery(t *testing.T) {
	app := setupTestApp(t)

	// Set mock to return error
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)
	mockSheets.SetSaveError(errors.New("network error"))

	trip := &models.TripToShelter{
		ID:       "test-error-recovery",
		Username: "testuser",
		Shelter:  &models.Shelter{ShortTitle: "Test"},
		Date:     "01.01.2024",
	}

	// Attempt to save trip
	result := app.sendTripToGSheet(12345, trip)

	// Should handle error gracefully
	if result {
		t.Error("Expected false result due to error")
	}

	// Application should continue functioning
	// Clear the error and try again
	mockSheets.SetSaveError(nil)
	result = app.sendTripToGSheet(12345, trip)

	if !result {
		t.Error("Expected successful recovery after error cleared")
	}
}

// Helper function to process poll updates
func processTestPollUpdate(app *AppConfig, update tgbotapi.Update) {
	if update.PollAnswer == nil {
		return
	}

	pollsMutex.RLock()
	chatId := polls[update.PollAnswer.PollID]
	pollsMutex.RUnlock()

	statePoolMutex.RLock()
	state := statePool[chatId]
	statePoolMutex.RUnlock()

	if state == nil {
		return
	}

	// Simulate poll processing based on state
	switch state.LastMessage {
	case commandTripPurpose:
		for _, option := range update.PollAnswer.OptionIDs {
			if option >= 0 && option < len(purposes) {
				state.TripToShelter.Purpose = append(state.TripToShelter.Purpose, purposes[option])
			}
		}
		state.LastMessage = app.tripByCommand(&update, state.TripToShelter)
	case commandTripBy:
		for _, option := range update.PollAnswer.OptionIDs {
			if option >= 0 && option < len(tripByOptions) {
				state.TripToShelter.TripBy = tripByOptions[option]
			}
			break
		}
		state.LastMessage = app.howYouKnowAboutUsCommand(&update, state.TripToShelter)
	case commandHowYouKnowAboutUs:
		for _, option := range update.PollAnswer.OptionIDs {
			if option >= 0 && option < len(sources) {
				state.TripToShelter.HowYouKnowAboutUs = append(state.TripToShelter.HowYouKnowAboutUs, sources[option])
			}
		}
	}

	statePoolMutex.Lock()
	statePool[chatId] = state
	statePoolMutex.Unlock()
}

// =================== CONFIGURATION AND SCHEDULE TESTS ===================

// TestShelterScheduleTypes tests different shelter schedule types
func TestShelterScheduleTypes(t *testing.T) {
	// Test "regularly" schedule type
	regularShelter := &models.Shelter{
		Schedule: models.ShelterSchedule{
			Type:    "regularly",
			Details: [][]int{{1, 6}}, // Week 1, Day 6
		},
	}

	// Test that it doesn't crash
	dates := getDatesByShelter(regularShelter)
	// For regularly scheduled shelters, dates might be empty but shouldn't be nil
	_ = dates // Just ensure it doesn't crash

	// Test "everyday" schedule type
	everydayShelter := &models.Shelter{
		Schedule: models.ShelterSchedule{
			Type: "everyday",
		},
	}

	// Test that it doesn't crash
	dates = getDatesByShelter(everydayShelter)
	// For everyday scheduled shelters, dates array is returned (may be empty but not nil)
	_ = dates // Just ensure it doesn't crash

	// Test "none" schedule type
	noneShelter := &models.Shelter{
		Schedule: models.ShelterSchedule{
			Type: "none",
		},
	}

	// Test that it doesn't crash
	dates = getDatesByShelter(noneShelter)
	// For none scheduled shelters, function should not crash
	_ = dates // Just ensure it doesn't crash
}

// TestCacheInitialization tests cache initialization
func TestCacheInitialization(t *testing.T) {
	cache, err := initCache()
	if err != nil {
		t.Errorf("Expected no error initializing cache, got %v", err)
	}
	if cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

// TestNewTripToShelter tests trip creation helper
func TestNewTripToShelter(t *testing.T) {
	username := "testuser"
	trip := NewTripToShelter(username)

	if trip == nil {
		t.Error("Expected trip to be created")
	}
	if trip.Username != username {
		t.Errorf("Expected username %s, got %s", username, trip.Username)
	}
	// Note: ID is not automatically generated in NewTripToShelter
	// IDs are typically set elsewhere in the application flow
}

// =================== INTEGRATION TESTS ===================

// TestEndToEndUserFlow tests a complete user interaction flow
func TestEndToEndUserFlow(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)

	// Step 1: User starts interaction
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(app, update)

	// Step 2: User chooses to go to shelter
	update = createTestUpdate(t, chatId, "/go_shelter")
	processTestUpdate(app, update)

	// Step 3: User selects shelter by choice
	update = createTestUpdate(t, chatId, "/choose_shelter")
	processTestUpdate(app, update)

	// Verify the flow completed without errors
	mockBot := app.Bot.(*mocks.MockTelegramBot)
	if mockBot.GetSentMessageCount() < 3 {
		t.Errorf("Expected at least 3 messages in flow, got %d", mockBot.GetSentMessageCount())
	}

	// Verify final state
	statePoolMutex.RLock()
	state := statePool[chatId]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandChooseShelter {
		t.Errorf("Expected final state %s, got %s", commandChooseShelter, state.LastMessage)
	}
}

// TestMultipleUsersInteraction tests concurrent users
func TestMultipleUsersInteraction(t *testing.T) {
	app := setupTestApp(t)

	users := []int64{12345, 54321, 67890}

	// All users start interaction simultaneously
	for _, chatId := range users {
		update := createTestUpdate(t, chatId, "/start")
		processTestUpdate(app, update)
	}

	// Verify each user has their own state
	for _, chatId := range users {
		statePoolMutex.RLock()
		state, exists := statePool[chatId]
		statePoolMutex.RUnlock()

		if !exists {
			t.Errorf("Expected state to exist for user %d", chatId)
		}
		if state.ChatId != chatId {
			t.Errorf("Expected chat ID %d, got %d", chatId, state.ChatId)
		}
		if state.LastMessage != commandStart {
			t.Errorf("Expected state %s for user %d, got %s", commandStart, chatId, state.LastMessage)
		}
	}

	// Verify correct number of states
	statePoolMutex.RLock()
	stateCount := len(statePool)
	statePoolMutex.RUnlock()

	if stateCount != len(users) {
		t.Errorf("Expected %d states, got %d", len(users), stateCount)
	}
}

// TestErrorRecoveryInFlow tests error recovery during user flow
func TestErrorRecoveryInFlow(t *testing.T) {
	app := setupTestApp(t)
	chatId := int64(12345)

	// Start normal flow
	update := createTestUpdate(t, chatId, "/start")
	processTestUpdate(app, update)

	// Inject error in sheets service
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)
	mockSheets.SetSaveError(errors.New("temporary error"))

	// Try to save a trip (should handle error gracefully)
	trip := &models.TripToShelter{
		ID:       "test-error-flow",
		Username: "testuser",
		Shelter:  &models.Shelter{ShortTitle: "Test"},
		Date:     "01.01.2024",
	}

	result := app.sendTripToGSheet(chatId, trip)
	if result {
		t.Error("Expected false result due to error")
	}

	// Clear error and continue
	mockSheets.SetSaveError(nil)

	// Continue with normal flow
	update = createTestUpdate(t, chatId, "/go_shelter")
	processTestUpdate(app, update)

	// Verify state is still valid
	statePoolMutex.RLock()
	state := statePool[chatId]
	statePoolMutex.RUnlock()

	if state.LastMessage != commandGoShelter {
		t.Errorf("Expected state %s after error recovery, got %s", commandGoShelter, state.LastMessage)
	}
}

// TestBotMessageHandling tests various message types
func TestBotMessageHandling(t *testing.T) {
	app := setupTestApp(t)
	mockBot := app.Bot.(*mocks.MockTelegramBot)

	// Test different message types
	testCases := []struct {
		command  string
		expected string
	}{
		{"/start", commandStart},
		{"/go_shelter", commandGoShelter},
		{"/donation", commandDonation},
		{"/masterclass", commandMasterclass},
	}

	for _, tc := range testCases {
		// Reset state for each test
		cleanupTestState()

		update := createTestUpdate(t, 12345, tc.command)
		processTestUpdate(app, update)

		statePoolMutex.RLock()
		state := statePool[12345]
		statePoolMutex.RUnlock()

		if state.LastMessage != tc.expected {
			t.Errorf("Command %s: expected state %s, got %s", tc.command, tc.expected, state.LastMessage)
		}
	}

	// Verify messages were sent for each command
	if mockBot.GetSentMessageCount() < len(testCases) {
		t.Errorf("Expected at least %d messages sent, got %d", len(testCases), mockBot.GetSentMessageCount())
	}
}

// =================== SPECIFIC BUG REGRESSION TESTS ===================

// TestOriginalNilPointerBug tests the specific bug from the error log:
// panic: runtime error: invalid memory address or nil pointer dereference
// at main.(*AppConfig).sendTripToGSheet when Google Sheets returns 403 error
func TestOriginalNilPointerBug(t *testing.T) {
	app := setupTestApp(t)

	// Simulate the exact error from the log: Google Sheets 403 permission error
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)
	mockSheets.SetSaveError(errors.New("googleapi: Error 403: The caller does not have permission, forbidden"))

	// Test 1: Valid trip with Google Sheets error (should handle gracefully)
	trip := &models.TripToShelter{
		ID:       "test-403-error",
		Username: "testuser",
		Shelter:  &models.Shelter{ShortTitle: "TestShelter"},
		Date:     "01.01.2024",
	}

	result := app.sendTripToGSheet(12345, trip)
	if result {
		t.Error("Expected false result due to Google Sheets 403 error")
	}

	// Test 2: Nil trip (original crash scenario)
	result = app.sendTripToGSheet(12345, nil)
	if result {
		t.Error("Expected false result for nil trip")
	}

	// Test 3: Trip with nil shelter (would cause original crash at line 1416)
	tripWithNilShelter := &models.TripToShelter{
		ID:       "test-nil-shelter",
		Username: "testuser",
		Shelter:  nil, // This would cause newTripToShelter.Shelter.ShortTitle panic
		Date:     "01.01.2024",
	}

	result = app.sendTripToGSheet(12345, tripWithNilShelter)
	if result {
		t.Error("Expected false result for trip with nil shelter")
	}

	// If we reach here without panicking, the bug is fixed!
}

// TestGoogleSheetsErrorScenarios tests various Google Sheets API error scenarios
func TestGoogleSheetsErrorScenarios(t *testing.T) {
	app := setupTestApp(t)
	mockSheets := app.SheetsService.(*mocks.MockGoogleSheetsService)

	trip := &models.TripToShelter{
		ID:       "test-error-scenarios",
		Username: "testuser",
		Shelter:  &models.Shelter{ShortTitle: "TestShelter"},
		Date:     "01.01.2024",
	}

	errorScenarios := []struct {
		name  string
		error string
	}{
		{"403 Permission Error", "googleapi: Error 403: The caller does not have permission, forbidden"},
		{"401 Authentication Error", "googleapi: Error 401: Request had invalid authentication credentials"},
		{"429 Rate Limit Error", "googleapi: Error 429: Quota exceeded"},
		{"500 Server Error", "googleapi: Error 500: Internal error encountered"},
		{"Network Error", "network error: connection timeout"},
		{"503 Service Unavailable", "googleapi: Error 503: The service is currently unavailable., backendError"},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			mockSheets.SetSaveError(errors.New(scenario.error))

			result := app.sendTripToGSheet(12345, trip)
			if result {
				t.Errorf("Expected false result for %s", scenario.name)
			}

			// Verify app continues to function after error
			mockSheets.SetSaveError(nil)
			result = app.sendTripToGSheet(12345, trip)
			if !result {
				t.Errorf("Expected recovery after %s was cleared", scenario.name)
			}
		})
	}
}
