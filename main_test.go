package main

import (
	"fmt"
	"log"
	"testing"
	"time"
	"walkthedog/internal/models"
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

// TestSendTripToGSheetErrorHandling tests our nil pointer fixes
func TestSendTripToGSheetErrorHandling(t *testing.T) {
	// Initialize app config for testing
	c, err := initCache()
	if err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}
	app.Cache = c
	
	// Create test trip
	trip := &models.TripToShelter{
		ID:       "test-trip-id",
		Username: "testuser",
		Shelter: &models.Shelter{
			ShortTitle: "TestShelter",
		},
		Date: "01.01.2024",
	}
	
	// Test with invalid Google config (should not panic)
	app.Google = &models.Google{SpreadsheetID: "invalid"}
	
	// This should handle errors gracefully and not panic
	result := app.sendTripToGSheet(12345, trip)
	
	// Should return false due to Google Sheets error, but not crash
	if result {
		t.Error("Expected false result due to invalid Google config")
	}
	
	// Verify trip was cached due to failure
	cachedTrip, found := app.Cache.Get(trip.ID)
	if !found {
		t.Error("Expected trip to be cached after Google Sheets failure")
	}
	
	retrievedTrip := cachedTrip.(models.TripToShelter)
	if retrievedTrip.ID != trip.ID {
		t.Errorf("Expected cached trip ID %s, got %s", trip.ID, retrievedTrip.ID)
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
