package main

import (
	"fmt"
	"log"
	"testing"
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
