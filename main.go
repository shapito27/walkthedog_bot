package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"walkthedog/internal/dates"
	sheet "walkthedog/internal/google/sheet"
	"walkthedog/internal/models"

	"github.com/davecgh/go-spew/spew"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v2"
)

// Environments
const (
	productionEnv  = "production"
	testEnv        = "test"
	developmentEnv = "development"
)

// Commands
const (
	commandStart       = "/start"
	commandMasterclass = "/masterclass"
	commandError       = "/error"

	// Related to donation
	commandDonation            = "/donation"
	commandDonationShelterList = "/donation_shelter_list"

	// Related to Shelter trip process
	commandGoShelter          = "/go_shelter"
	commandChooseShelter      = "/choose_shelter"
	commandTripDates          = "/trip_dates"
	commandChooseDates        = "/choose_date"
	commandIsFirstTrip        = "/is_first_trip"
	commandTripPurpose        = "/trip_purpose"
	commandTripBy             = "/trip_by"
	commandHowYouKnowAboutUs  = "/how_you_know_about_us"
	commandSummaryShelterTrip = "/summary_shelter_trip"
)

// Answers
const (
	chooseByShelter = "Выбор по приюту"
	chooseByDate    = "Выбор по дате"
)

// Phrases
const (
	errorWrongShelterName = "не похоже на название приюта"
)

// purposes represents list of available purposes user can choose to going to shelter.
var purposes = []string{
	"Погулять с собаками",
	"Помочь приюту руками (прибрать, помыть, почесать :-)",
	"Пофотографировать животных для соц.сетей",
	"Привезти корм/медикаменты и т.п. для нужд приюта",
	"Перевести деньги для приюта",
	"Есть другие идеи (обязательно расскажите нам на выезде :-)",
}

// tripByOptions represents list of options to come to shelters.
var tripByOptions = []string{
	"Еду на своей машине или с кем-то на машине (мест больше нет)",
	"Еду на своей машине или с кем-то на машине (готов предложить места другим волонтерам)",
	"Еду общественным транспортом",
	"Ищу с кем поехать",
	"Какой-то другой магический вариант :)",
}

// sources represents list of available sources of information user knew about walkthedog.
var sources = []string{
	"Сарафанное радио (друзья, родственники, коллеги)",
	"Выставка или ярмарка",
	"Нашел в интернете",
	"Мосволонтер",
	"Вконтакте",
	"Наш канал в WhatsApp",
	"Наш канал в Telegram",
	"Другие социальные сети",
	"Авито/Юла",
	"Знаю вас уже давно :)",
	"Другой вариант",
}

// statePool store all chat states
var statePool = make(map[int64]*models.State)

//TODO: remove poll_id after answer.
// polls stores poll_id => chat_id
var polls = make(map[string]int64)

type EnvironmentConfig map[string]*models.TelegramConfig

// SheltersList represents list of Shelters
type SheltersList map[int]*models.Shelter

// NewTripToShelter initializes new object for storing user's trip information.
func NewTripToShelter(userName string) *models.TripToShelter {
	return &models.TripToShelter{
		Username: userName,
	}
}

func main() {
	// getting config by environment
	env := productionEnv //developmentEnv
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

	// getting shelters
	shelters, err := getShelters()
	if err != nil {
		log.Panic(err)
	}

	var newTripToShelter *models.TripToShelter

	// getting message
	for update := range updates {
		var chatId int64
		// extract chat id for different cases
		if update.Message != nil {
			chatId = update.Message.Chat.ID
		} else if update.PollAnswer != nil {
			chatId = polls[update.PollAnswer.PollID]
		}

		// fetching state or init new
		state, ok := statePool[chatId]
		log.Printf("**state**: %+v", state)
		if !ok {
			state = &models.State{
				ChatId:      chatId,
				LastMessage: "",
			}
			statePool[chatId] = state
		}
		// initilize last message and trip to shelter
		lastMessage = state.LastMessage
		newTripToShelter = state.TripToShelter

		if update.Message != nil { // If we got a message
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
			log.Printf("lastMessage: %s", lastMessage)

			var msgObj tgbotapi.MessageConfig
			//check for commands
			switch update.Message.Text {
			case commandStart:
				log.Println("[walkthedog_bot]: Send start message")
				msgObj = startMessage(chatId)
				msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msgObj)
				lastMessage = commandStart
			case commandGoShelter:
				log.Println("[walkthedog_bot]: Send appointmentOptionsMessage message")
				lastMessage = goShelterCommand(bot, &update)
			case commandChooseShelter:
				lastMessage = chooseShelterCommand(bot, &update, &shelters)
			case commandTripDates:
				lastMessage = tripDatesCommand(bot, &update, newTripToShelter, &shelters, lastMessage)
			case commandMasterclass:
				log.Println("[walkthedog_bot]: Send masterclass")
				msgObj = masterclass(chatId)
				bot.Send(msgObj)
				lastMessage = commandMasterclass
			case commandDonation:
				log.Println("[walkthedog_bot]: Send donation")
				lastMessage = donationCommand(bot, chatId)
			case commandDonationShelterList:
				log.Println("[walkthedog_bot]: Send donationShelterList")
				msgObj = donationShelterList(chatId, &shelters)
				bot.Send(msgObj)
				lastMessage = commandDonationShelterList
			default:
				switch lastMessage {
				case commandGoShelter:
					if update.Message.Text == chooseByShelter {
						lastMessage = chooseShelterCommand(bot, &update, &shelters)
					} else if update.Message.Text == chooseByDate {
						//lastMessage = tripDatesCommand(bot, &update, newTripToShelter, &shelters, lastMessage)
						ErrorFrontend(bot, &update, newTripToShelter, "Запись по Времени пока не доступна 😥")
						lastMessage = goShelterCommand(bot, &update)
						break
					} else {
						ErrorFrontend(bot, &update, newTripToShelter, fmt.Sprintf("Нажмите кноку \"%s\" или \"%s\"", chooseByDate, chooseByShelter))
						lastMessage = goShelterCommand(bot, &update)
						break
					}
				// when shelter was chosen next step to chose date
				case commandChooseShelter:
					if newTripToShelter == nil {
						newTripToShelter = NewTripToShelter(update.Message.From.UserName)
					}
					shelter, err := shelters.getShelterByNameID(update.Message.Text)

					if err != nil {
						ErrorFrontend(bot, &update, newTripToShelter, err.Error())
						chooseShelterCommand(bot, &update, &shelters)
						break
					}
					newTripToShelter.Shelter = shelter

					log.Println("[walkthedog_bot]: Send whichDate question")
					msgObj = whichDate(chatId, shelter)
					bot.Send(msgObj)
					lastMessage = commandChooseDates
				case commandChooseDates:
					if isTripDateValid(update.Message.Text, newTripToShelter) {
						lastMessage = isFirstTripCommand(bot, &update, newTripToShelter)
					} else {
						ErrorFrontend(bot, &update, newTripToShelter, "Кажется вы ошиблись с датой 🤔")
						lastMessage = tripDatesCommand(bot, &update, newTripToShelter, &shelters, lastMessage)
					}
				case commandIsFirstTrip:
					lastMessage, err = tripPurposeCommand(bot, &update, newTripToShelter)
					if err != nil {
						ErrorFrontend(bot, &update, newTripToShelter, err.Error())
						if isTripDateValid(update.Message.Text, newTripToShelter) {
							lastMessage = isFirstTripCommand(bot, &update, newTripToShelter)
						} else {
							lastMessage = tripDatesCommand(bot, &update, newTripToShelter, &shelters, lastMessage)
						}
					}
				case commandTripPurpose:
					ErrorFrontend(bot, &update, newTripToShelter, "Выберите цели поездки и нажмите кнопку голосовать")
				case commandTripBy:
					ErrorFrontend(bot, &update, newTripToShelter, "Расскажите как добираетесь до приюта")
				case commandHowYouKnowAboutUs:
					ErrorFrontend(bot, &update, newTripToShelter, "Расскажите как вы о нас узнали")
				default:
					log.Println("[walkthedog_bot]: Unknown command")

					message := "Не понимаю 🐶 Попробуй " + commandStart
					msgObj := tgbotapi.NewMessage(chatId, message)
					bot.Send(msgObj)
					lastMessage = commandChooseDates
				}
			}
		} else if update.Poll != nil {
			//log.Printf("[%s]: %s", update.FromChat().FirstName, "save poll id")
			//polls[update.Poll.ID] = update.FromChat().ID
		} else if update.PollAnswer != nil {
			log.Printf("[%s]: %v", update.PollAnswer.User.UserName, update.PollAnswer.OptionIDs)
			log.Printf("lastMessage: %s", lastMessage)

			switch lastMessage {
			case commandTripPurpose:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.Purpose = append(newTripToShelter.Purpose, purposes[option])
				}

				lastMessage = tripByCommand(bot, &update, newTripToShelter)
			case commandTripBy:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.TripBy = tripByOptions[option]
					break
				}
				lastMessage = howYouKnowAboutUsCommand(bot, &update, newTripToShelter)
			case commandHowYouKnowAboutUs:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.HowYouKnowAboutUs = sources[option]
					break
				}

				summaryCommand(bot, &update, newTripToShelter)
				lastMessage = donationCommand(bot, polls[update.PollAnswer.PollID])

				//save to google sheet
				srv, err := sheet.NewService()
				if err != nil {
					log.Fatalf(err.Error())
				} else {
					resp, err := sheet.SaveTripToShelter(srv, newTripToShelter)
					if err != nil {
						log.Fatalf("Unable to write data from sheet: %v", err)
					}
					if resp.ServerResponse.HTTPStatusCode != 200 {
						log.Fatalf("error: %+v", resp)
					}
				}
			}
		}
		// save state to pool
		state.LastMessage = lastMessage
		state.TripToShelter = newTripToShelter
		statePool[chatId] = state
		log.Println("[trip_state]: ", newTripToShelter)
	}
}

// goShelterCommand prepares message about available options to start appointment to shelter and then sends it and returns last command.
func goShelterCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) string {
	msgObj := appointmentOptionsMessage(update.Message.Chat.ID)
	bot.Send(msgObj)
	return commandGoShelter
}

// chooseShelterCommand prepares message about available shelters and then sends it and returns last command.
func chooseShelterCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, shelters *SheltersList) string {
	log.Println("[walkthedog_bot]: Send whichShelter question")
	msgObj := whichShelter(update.Message.Chat.ID, shelters)
	bot.Send(msgObj)
	return commandChooseShelter
}

// isFirstTripCommand prepares message with question "is your first trip?" and then sends it and returns last command.
func isFirstTripCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	newTripToShelter.Date = update.Message.Text
	msgObj := isFirstTrip(update.Message.Chat.ID)
	bot.Send(msgObj)
	return commandIsFirstTrip
}

// isTripDateValid return true if it's one of the available dates of shelter trip.
func isTripDateValid(date string, newTripToShelter *models.TripToShelter) bool {
	isCorrectDate := false

	if newTripToShelter == nil {
		return false
	}
	if newTripToShelter.Shelter == nil {
		return false
	}

	shelterDates := getDatesByShelter(newTripToShelter.Shelter)
	for _, v := range shelterDates {
		if v == date {
			isCorrectDate = true
			break
		}
	}
	return isCorrectDate
}

// tripPurposeCommand prepares poll with question about your purpose for this trip and then sends it and returns last command.
func tripPurposeCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter) (string, error) {
	if update.Message.Text == "Да" {
		newTripToShelter.IsFirstTrip = true
	} else if update.Message.Text == "Нет" {
		newTripToShelter.IsFirstTrip = false
	} else {
		return commandIsFirstTrip, errors.New("доступные ответы \"Да\" и \"Нет\"")
	}

	msgObj := tripPurpose(update.Message.Chat.ID)

	responseMessage, _ := bot.Send(msgObj)
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID

	return commandTripPurpose, nil
}

// tripByCommand prepares poll with question about how he going to come to shelter and then sends it and returns last command.
func tripByCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	msgObj := tripBy(polls[update.PollAnswer.PollID])
	responseMessage, _ := bot.Send(msgObj)
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID
	return commandTripBy
}

// howYouKnowAboutUsCommand prepares poll with question about where did you know about us and then sends it and returns last command.
func howYouKnowAboutUsCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	msgObj := howYouKnowAboutUs(polls[update.PollAnswer.PollID])
	responseMessage, _ := bot.Send(msgObj)
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID
	return commandHowYouKnowAboutUs
}

// summaryCommand prepares message with summary and then sends it and returns last command.
func summaryCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	msgObj := summary(polls[update.PollAnswer.PollID], newTripToShelter)
	bot.Send(msgObj)
	return commandSummaryShelterTrip
}

// donationCommand prepares message with availabele ways to dontate us or shelters and then sends it and returns last command.
func donationCommand(bot *tgbotapi.BotAPI, chatId int64) string {
	msgObj := donation(chatId)
	bot.Send(msgObj)
	return commandDonation
}

// tripDatesCommand prepares message with availabele dates to go to shelters and then sends it and returns last command.
func tripDatesCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter, shelters *SheltersList, lastMessage string) string {
	if newTripToShelter == nil {
		message := "По времени записаться пока нельзя :("
		msgObj := tgbotapi.NewMessage(update.Message.Chat.ID, message)
		bot.Send(msgObj)
		return commandGoShelter
		/* newTripToShelter = NewTripToShelter(update.Message.From.UserName)
		if lastMessage == commandChooseShelter {
			panic("change it if I use it")
			shelter, err := shelters.getShelterByNameID(update.Message.Text)

			if err != nil {
				ErrorFrontend(bot, update, newTripToShelter, err.Error())
				chooseShelterCommand(bot, update, shelters)
			}
			newTripToShelter.Shelter = shelter
		} else if lastMessage == commandGoShelter {

		} */
	}
	log.Println("[walkthedog_bot]: Send whichDate question")
	msgObj := whichDate(update.Message.Chat.ID, newTripToShelter.Shelter)
	bot.Send(msgObj)
	return commandChooseDates
}

// getShelterByNameID returns Shelter and error using given shelter name in following format:
// 1. Хаски Хелп (Истра)
// it substr string before dot and try to find shelter by ID.
func (shelters SheltersList) getShelterByNameID(name string) (*models.Shelter, error) {
	dotPosition := strings.Index(name, ".")
	if dotPosition == -1 {
		//log.Println(errors.New(fmt.Sprintf("message %s don't contain dot", name)))
		return nil, errors.New(errorWrongShelterName)
	}
	shelterId, err := strconv.Atoi(name[0:dotPosition])
	if err != nil {
		log.Println(err)
		return nil, errors.New(errorWrongShelterName)
	}
	//log.Println("id part", update.Message.Text[0:strings.Index(update.Message.Text, ".")])
	shelter, ok := shelters[shelterId]
	if !ok {
		log.Println(fmt.Errorf("shelter name \"%s\", extracted id=\"%d\" is not found", name, shelterId))
		return nil, errors.New(errorWrongShelterName)
	}

	return shelter, nil
}

// ErrorFrontend sends error message to user and returns last command.
func ErrorFrontend(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *models.TripToShelter, errMessage string) string {
	log.Println("[walkthedog_bot]: Send ERROR")
	if errMessage == "" {
		errMessage = "Error"
	}
	msgObj := errorMessage(update.Message.Chat.ID, errMessage)
	bot.Send(msgObj)
	return commandError
}

// getConfig returns config by environment.
func getConfig(environment string) (*models.TelegramConfig, error) {
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

// getShelters returns list of shelters with information about them.
func getShelters() (SheltersList, error) {
	yamlFile, err := ioutil.ReadFile("configs/shelters.yml")
	if err != nil {
		return nil, err
	}
	var sheltersList SheltersList
	err = yaml.Unmarshal(yamlFile, &sheltersList)
	if err != nil {
		return nil, err
	}

	return sheltersList, nil
}

// masterclass returns masterclasses.
func masterclass(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := `Запись на мастер-классы скоро здесь появится, а пока вы можете записаться на ближайший на walkthedog.ru/cages`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// donationShelterList returns information about donations.
func donationShelterList(chatId int64, shelters *SheltersList) tgbotapi.MessageConfig {
	message := "Пожертвовать в приют:\n"

	for i := 1; i <= len(*shelters); i++ {
		if len((*shelters)[i].DonateLink) == 0 {
			continue
		}
		message += fmt.Sprintf("%s. %s\n %s\n", (*shelters)[i].ID, (*shelters)[i].Title, (*shelters)[i].DonateLink)
	}
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.DisableWebPagePreview = true
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	return msgObj
}

// startMessage returns first message with all available commands.
func startMessage(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := `🐕 /go_shelter Записаться на выезд в приют

📐 /masterclass Записаться на мастер-класс по изготовлению будок и котодомиков для приютов

❤️ /donation Сделать пожертвование

@walkthedog_support Задать вопрос или предложить добрую идею

@walkthedog Подписаться на наш телеграм канал`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// appointmentOptionsMessage returns message with 2 options.
func appointmentOptionsMessage(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "Вы можете записаться на выезд в приют исходя из даты (напр. хотите поехать в ближайшие выходные) или выбрать конкретный приют и записаться на ближайший выезд в него. На страничке walkthedog.ru/shelters есть удобная карта, которая покажет ближайший к вам приют."
	msgObj := tgbotapi.NewMessage(chatId, message)

	var numericKeyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(chooseByDate),
		tgbotapi.NewKeyboardButton(chooseByShelter),
	))
	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// whichShelter returns message with question "Which Shelter you want go" and button options.
func whichShelter(chatId int64, shelters *SheltersList) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "В какой приют желаете записаться?"
	msgObj := tgbotapi.NewMessage(chatId, message)

	var sheltersButtons [][]tgbotapi.KeyboardButton
	log.Println("shelters before range", shelters)

	for i := 1; i <= len(*shelters); i++ {
		buttonRow := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(fmt.Sprintf("%s. %s", (*shelters)[i].ID, (*shelters)[i].Title)),
		)

		sheltersButtons = append(sheltersButtons, buttonRow)
	}
	log.Println("sheltersButtons", sheltersButtons)
	var numericKeyboard = tgbotapi.NewReplyKeyboard(sheltersButtons...)
	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// whichShelter returns message with question "Which Shelter you want go" and button options.
func errorMessage(chatId int64, message string) tgbotapi.MessageConfig {
	msgObj := tgbotapi.NewMessage(chatId, message)
	return msgObj
}

// whichDate returns object including message text "Which Date you want to go" and other message config.
func whichDate(chatId int64, shelter *models.Shelter) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "Выберите дату выезда:"
	msgObj := tgbotapi.NewMessage(chatId, message)

	var numericKeyboard tgbotapi.ReplyKeyboardMarkup
	var dateButtons [][]tgbotapi.KeyboardButton

	shelterDates := getDatesByShelter(shelter)
	for _, value := range shelterDates {
		buttonRow := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(value),
		)
		dateButtons = append(dateButtons, buttonRow)
	}
	numericKeyboard = tgbotapi.NewReplyKeyboard(dateButtons...)

	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// getDatesByShelter return list of dates.
func getDatesByShelter(shelter *models.Shelter) []string {
	var shedule []string
	now := time.Now()
	spew.Dump(shelter)
	if shelter.Schedule.Type == "regularly" {

		scheduleWeek := shelter.Schedule.Details[0]
		scheduleDay := shelter.Schedule.Details[1]
		scheduleTime := shelter.Schedule.TimeStart
		for i := 0; i < 6; i++ {
			month := time.Month(int(now.Month()) + i)
			day := calculateDay(scheduleDay, scheduleWeek, month)
			if i == 0 && now.Day() > day.Day() {
				continue
			}
			shedule = append(shedule, dates.WeekDaysRu[day.Weekday()]+" "+day.Format("02.01.2006")+" "+scheduleTime)

		}
	} else if shelter.Schedule.Type == "everyday" {
		//TODO: finish everyday type
	}

	return shedule
}

// isFirstTrip returns object including message text "is your first trip" and other message config.
func isFirstTrip(chatId int64) tgbotapi.MessageConfig {
	message := "Это ваша первая поездка?"
	msgObj := tgbotapi.NewMessage(chatId, message)

	var numericKeyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Да"),
		tgbotapi.NewKeyboardButton("Нет"),
	))
	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

// tripPurpose returns object including poll about trip purpose and other poll config.
func tripPurpose(chatId int64) tgbotapi.SendPollConfig {
	message := "🎯 Чем хочу помочь"
	options := purposes
	msgObj := tgbotapi.NewPoll(chatId, message, options...)
	msgObj.AllowsMultipleAnswers = true
	msgObj.IsAnonymous = false
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return msgObj
}

// tripBy returns object including poll about how he/she going to come to shelter poll config.
func tripBy(chatId int64) tgbotapi.SendPollConfig {
	message := "🚗 Как добираетесь до приюта?"
	options := tripByOptions
	msgObj := tgbotapi.NewPoll(chatId, message, options...)
	msgObj.IsAnonymous = false
	msgObj.AllowsMultipleAnswers = false
	return msgObj
}

// howYouKnowAboutUs returns object including poll about how he/she know about us and other poll config.
func howYouKnowAboutUs(chatId int64) tgbotapi.SendPollConfig {
	message := "🤫 Как вы о нас узнали?"
	options := sources
	msgObj := tgbotapi.NewPoll(chatId, message, options...)
	msgObj.IsAnonymous = false
	msgObj.AllowsMultipleAnswers = false
	return msgObj
}

// summary returns object including message text with summary of user's answers and other message config.
func summary(chatId int64, newTripToShelter *models.TripToShelter) tgbotapi.MessageConfig {
	guide := `Место проведения: %s (точный адрес приюта отправляется в чат после регистрации)

📎 За 5-7 дней до выезда мы пришлем вам ссылку для добавления в Whats App чат, где расскажем все детали и ответим на вопросы. До встречи!
	`
	if newTripToShelter.Shelter.Guide != "" {
		guide = "Все детали о выезде в приют включая адрес, как доехать, что взять и потребности приюта: " + newTripToShelter.Shelter.Guide
	} else {
		guide = fmt.Sprintf(guide, newTripToShelter.Shelter.Address)
	}
	message := fmt.Sprintf(`Регистрация прошла успешно.
	
ℹ️ Информация о событии
Выезд в приют: <a href="%s">%s</a>
Дата и время: %s
%s

❤️ Напоминаем, что участие в выезде в приют является бесплатным. При этом вы можете сделать добровольное пожертвование.

💬 За 5 дней до выезда мы добавим вас в телеграм-чат выезда, где можно будет задать вопросы и уточнить о волонтерах, у кого будут места в машине.

Если у вас появятся вопросы до добавления в чат - пишите @walkthedog_support
`, newTripToShelter.Shelter.Link,
		newTripToShelter.Shelter.Title,
		newTripToShelter.Date,
		guide)
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML

	return msgObj
}

// donation set donation text and message options and returns MessageConfig.
func donation(chatId int64) tgbotapi.MessageConfig {
	message :=
		`Добровольное пожертвование в 500 рублей и более осчастливит 1 собаку (500 рублей = 2 недели питания одной собаки в приюте). Все собранные деньги будут переданы в приют.

📍 /donation_shelter_list - пожертвовать в конкретный приют

📍 Перевод по номеру телефона +7 916 085 1342 (Михайлов Дмитрий) - укажите "пожертвование".

📍 <a href="https://yoomoney.ru/to/410015848442299">Яндекс.Деньги</a>
`
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML
	msgObj.DisableWebPagePreview = true
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	return msgObj
}

// calculateDay returns the date of by given day of week, week number and month.
func calculateDay(dayOfWeek int, week int, month time.Month) time.Time {
	firstDayOfMonth := time.Date(time.Now().Year(), month, 1, 0, 0, 0, 0, time.UTC)
	//currentDay := (8 - int(firstDayOfMonth.Weekday())) % 7

	currentDay := int(firstDayOfMonth.Weekday())
	if currentDay == 0 {
		currentDay = 7
	}
	var resultDay int
	if dayOfWeek == currentDay {
		resultDay = 1 + 7*(week-1)
	} else if dayOfWeek > currentDay {
		resultDay = 1 + (dayOfWeek - currentDay) + (week-1)*7
	} else {
		resultDay = 1 + (7 - currentDay + dayOfWeek) + (week-1)*7
	}

	return time.Date(time.Now().Year(), month, resultDay, 0, 0, 0, 0, time.UTC)
}
