package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v2"
)

const (
	productionEnv  = "production"
	testEnv        = "test"
	developmentEnv = "development"
)

type EnvironmentConfig map[string]*TelegramConfig

type TelegramConfig struct {
	APIToken string `yaml:"api_token"`
	Timeout  int    `yaml:"timeout"`
}

type SheltersList map[int]Shelter

type Shelter struct {
	Title string `yaml:"title"`
	Link  string `yaml:"link"`
}

type TripToShelter struct {
	Username string
	Shelter *Shelter
	Date string
	IsFirstTrip bool
	UserPurpose string
	TripBy string
	HowYouKnowAboutUs string
}

func main() {
	// getting config by environment
	env := developmentEnv
	config, err := getConfig(env)
	if err != nil {
		log.Panic(err)
	}

	// bot init
	bot, err := tgbotapi.NewBotAPI(config.APIToken)
	if err != nil {
		log.Panic(err)
	}

	if env == developmentEnv {
		bot.Debug = true
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// set how often check for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = config.Timeout

	updates := bot.GetUpdatesChan(u)

	var lastMessage string
	shelters, err := getShelters()
	if err != nil {
		log.Panic(err)
	}
	// getting message
	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)

			var msgObj tgbotapi.MessageConfig
			//check for commands
			switch update.Message.Text {
			case "/start":
				log.Println("[walkthedog_bot]: Send start message")
				msgObj = startMessage(update.Message.Chat.ID)
				bot.Send(msgObj)
				lastMessage = "/start"
			case "/go_shelter":
				log.Println("[walkthedog_bot]: Send whichShelter question")
				msgObj = whichShelter(update.Message.Chat.ID, shelters)
				bot.Send(msgObj)
				lastMessage = "/go_shelter"
			case "/choose_shelter":
				log.Println("[walkthedog_bot]: Send whichShelter question")
				msgObj = whichShelter(update.Message.Chat.ID, shelters)
				bot.Send(msgObj)
				lastMessage = "/choose_shelter"
			case "/masterclass":
				log.Println("[walkthedog_bot]: Send masterclass")
				msgObj = masterclass(update.Message.Chat.ID)
				bot.Send(msgObj)
				lastMessage = "/masterclass"
			case "/donation":
				log.Println("[walkthedog_bot]: Send donation")
				msgObj = donation(update.Message.Chat.ID)
				bot.Send(msgObj)
				lastMessage = "/donation"
			}

			log.Println("lastMessage", lastMessage, "shelter_name", update.Message.Text == "Хаски Хелп (Истра)")
			var shelter string
			if lastMessage == "/choose_shelter" && (update.Message.Text == "Хаски Хелп (Истра)" || update.Message.Text == "Приют \"Ника\" (Зеленоград)") {
				shelter = update.Message.Text

				message := "Хороший выбор!\n%s будет рад вам.\nАдрес: Московская область, городской округ Истра, деревня Карцево.\nО приюте: https://walkthedog.ru/huskyhelp"
				msgObj := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(message, shelter))
				bot.Send(msgObj)

				log.Println("[walkthedog_bot]: Send whichDate question")
				msgObj = whichDate(update.Message.Chat.ID)
				bot.Send(msgObj)
				lastMessage = "/choose_date"
			}

			//lastMessage = update.Message.Text
		}
	}
}

// getConfig return config by environment.
func getConfig(environment string) (*TelegramConfig, error) {
	yamlFile, err := ioutil.ReadFile("configs/telegram.yml")
	if err != nil {
		return nil, err
	}
	var environmentConfig EnvironmentConfig
	err = yaml.Unmarshal(yamlFile, &environmentConfig)
	if err != nil {
		return nil, err
	}

	if environmentConfig[environment] == nil {
		return nil, errors.New("wrong environment set")
	}

	log.Println(environmentConfig[environment])

	return environmentConfig[environment], nil
}

// getShelters return list of shelters with information about them.
func getShelters() (*SheltersList, error) {
	yamlFile, err := ioutil.ReadFile("configs/shelters.yml")
	if err != nil {
		return nil, err
	}
	var sheltersList SheltersList
	err = yaml.Unmarshal(yamlFile, &sheltersList)
	if err != nil {
		return nil, err
	}

	log.Println("sheltersList", sheltersList)

	return &sheltersList, nil
}

// masterclass return message with options.
func masterclass(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := `TODO masterclass message`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// donation return message with options.
func donation(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := `Пожертвовать можно в:
	1. Хаски Хелп (Истра) https://www.tinkoff.ru/sl/1msxKU5XTyS`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// startMessage return message with options.
func startMessage(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := `- /go_shelter Записаться на выезд в приют
- /masterclass Записаться на мастерклас
- /donation Сделать пожертвование`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// whichShelter return message with question "Which Shelter you want go" and button options.
func whichShelter(chatId int64, shelters *SheltersList) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "В какой приют желаете записаться?"
	msgObj := tgbotapi.NewMessage(chatId, message)

	var sheltersButtons [][]tgbotapi.KeyboardButton
	for _, v := range *shelters {
		buttonRow := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(v.Title),
		)
		log.Println("v.Title", v.Title)
		sheltersButtons = append(sheltersButtons, buttonRow)
	}
	log.Println("sheltersButtons", sheltersButtons)
	var numericKeyboard = tgbotapi.NewReplyKeyboard(sheltersButtons...)
	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// whichDate return message with question "Which Date you want go" and button options
func whichDate(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "Выберите дату выезда"
	msgObj := tgbotapi.NewMessage(chatId, message)

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	fmt.Println(firstOfMonth)

	for i := 0; i < 6; i++ {
		month := time.Month(int(time.Now().Month()) + i)
		//@todo finish this function. Need to calculate first saturday, sunday for each month correctly
		day := calculateDay(6, 1, month)
		log.Println(strconv.Itoa(day) + " " + strconv.Itoa(int(month)) + " суббота")
	}
	var numericKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(firstOfMonth.Format("Mon 2 Jan") + " 11:00"),
		),
	)
	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// FirstMonday returns the day of the first Monday in the given month.
func calculateDay(dayOfWeek int, week int, month time.Month) int {
	t := time.Date(time.Now().Year(), month, 1, 0, 0, 0, 0, time.UTC)
	return (8-int(t.Weekday()))%7 + (week-1)*7 + dayOfWeek
}
