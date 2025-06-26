// Package sheet helps to read/write Google sheets
package sheet

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"walkthedog/internal/interfaces"
	"walkthedog/internal/models"
)

type googleSheet struct {
	SpreadsheetID string
	Service       *sheets.Service
}

func NewGoogleSpreadsheet(google models.Google) (interfaces.GoogleSheetsService, error) {
	//save to google sheet
	srv, err := NewService()
	if err != nil {
		return nil, err
	}

	return &googleSheet{
		SpreadsheetID: google.SpreadsheetID,
		Service:       srv,
	}, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		return nil, err
	}

	return config.Client(context.Background(), tok), nil
}

/* // Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
} */

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//send url to user
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// NewService reads creadentials, prepares config, creates client and creates new service.
func RequestAuthCodeURL() (string, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return "", fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return "", fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return authURL, nil
}

// AuthorizationCodeToToken gets auth code and try to exchange it into token.
func AuthorizationCodeToToken(authCode string) error {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("Unable to retrieve token from web: %v", err)
	}
	saveToken(tokFile, tok)

	return nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// NewService reads creadentials, prepares config, creates client and creates new service.
func NewService() (*sheets.Service, error) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %v", err)
	}

	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

// SaveTripToShelter saves information about trip to google sheet.
func (googleSheetService googleSheet) SaveTripToShelter(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error) {
	var vr sheets.ValueRange
	now := time.Now()
	tripToShelterInfo := []interface{}{
		tripToShelter.Username,
		tripToShelter.Shelter.Title,
		tripToShelter.Date,
		strconv.FormatBool(tripToShelter.IsFirstTrip),
		strings.Join(tripToShelter.Purpose, ","),
		tripToShelter.TripBy,
		strings.Join(tripToShelter.HowYouKnowAboutUs, ","),
		now.Format("02.01.2006 15:04:05"),
	}
	vr.Values = append(vr.Values, tripToShelterInfo)

	readRange := fmt.Sprintf("%s!A2:H", sheetName)

	return googleSheetService.Service.Spreadsheets.Values.Append(googleSheetService.SpreadsheetID, readRange, &vr).ValueInputOption("RAW").Do()
}

// SaveTripToShelter saves information about trip in short format to System sheet to google sheet.
func (googleSheetService googleSheet) SaveTripToShelterSystem(sheetName string, tripToShelter *models.TripToShelter) (*sheets.AppendValuesResponse, error) {
	var vr sheets.ValueRange
	now := time.Now()
	tripToShelterInfo := []interface{}{
		tripToShelter.Username,
		tripToShelter.Shelter.ShortTitle,
		tripToShelter.Date,
		now.Format("02.01.2006 15:04:05"),
	}
	vr.Values = append(vr.Values, tripToShelterInfo)

	readRange := fmt.Sprintf("%s!A1:D", sheetName)

	return googleSheetService.Service.Spreadsheets.Values.Append(googleSheetService.SpreadsheetID, readRange, &vr).ValueInputOption("RAW").Do()
}

// CreateSheet creates sheet.
func (googleSheetService googleSheet) CreateSheet(sheetName string) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	req := sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: sheetName,
			},
		},
	}

	rbb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&req},
	}

	return googleSheetService.Service.Spreadsheets.BatchUpdate(googleSheetService.SpreadsheetID, rbb).Context(context.Background()).Do()
}

// AddSheetHeaders adds headers for new sheet.
func (googleSheetService googleSheet) AddSheetHeaders(sheetName string) (*sheets.AppendValuesResponse, error) {
	//User	Приют	Дата	Первый раз	Цели	Как добирается	Откуда узнал	Дата регистрации на выезд (UTC +8)	Статус
	var vr sheets.ValueRange
	headers := []interface{}{
		"User",
		"Приют",
		"Дата",
		"Первый раз",
		"Цели",
		"Как добирается",
		"Откуда узнал",
		"Дата регистрации на выезд (UTC +8)",
		"Статус",
	}
	vr.Values = append(vr.Values, headers)

	readRange := fmt.Sprintf("%s!A1:I", sheetName)

	return googleSheetService.Service.Spreadsheets.Values.Append(googleSheetService.SpreadsheetID, readRange, &vr).ValueInputOption("RAW").Do()
}

// HasSheet checks is sheet exist
func (googleSheetService googleSheet) HasSheet(sheetName string) bool {
	_, err := googleSheetService.Service.Spreadsheets.Values.Get(googleSheetService.SpreadsheetID, fmt.Sprintf("%s!A1:B1", sheetName)).Do()

	return err == nil
}

// PrepareSheetForSavingData check if sheet exists. If no create it and headers
func (googleSheetService googleSheet) PrepareSheetForSavingData(sheetName string) error {
	if !googleSheetService.HasSheet(sheetName) {
		_, err := googleSheetService.CreateSheet(sheetName)
		if err != nil {
			return err
		}
		_, err = googleSheetService.AddSheetHeaders(sheetName)
		if err != nil {
			return err
		}
	}

	return nil
}

/* func readSheet() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1sGRvjVEP4OvT4JubFmn9pp_h72vwO5aSHvZdUFoEFsY"
	readRange := "Class Data!A2:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[4])
		}
	}
} */
