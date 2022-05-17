package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v2"
)

const (
	productionEnv  = "production"
	testEnv        = "test"
	developmentEnv = "development"
)

var purposes = []string{
	"Погулять с собаками",
	"Помочь приюту руками (прибрать, помыть, почесать :-)",
	"Пофотографировать животных для соц.сетей",
	"Привезти корм/медикаменты и т.п. для нужд приюта",
	"Перевести деньги для приюта",
	"Есть другие идеи (обязательно расскажите нам на выезде :-)",
}

var sources = []string{
	"Сарафанное радио или друзья, родственники, коллеги",
	"Сайт walkthedog.ru (погуляйсобаку.рф)",
	"Выставка или ярмарка",
	"Нашел в интернете",
	"Радио или ТВ",
	"Вконтакте",
	"Facebook",
	"Instagram",
	"Наш канал в WhatsApp",
	"Наш канал в Telegram",
}

//TODO: remove poll_id after anser.
// polls stores poll_id => chat_id
var polls = make(map[string]int64)

type EnvironmentConfig map[string]*TelegramConfig

type TelegramConfig struct {
	APIToken string `yaml:"api_token"`
	Timeout  int    `yaml:"timeout"`
}

type SheltersList map[int]*Shelter

type ShelterSchedule struct {
	Type      string `yaml:"type"`
	Details   []int  `yaml:"details"`
	TimeStart string `yaml:"time_start"`
	TimeEnd   string `yaml:"time_end"`
}

type Shelter struct {
	ID         string          `yaml:"id"`
	Address    string          `yaml:"address"`
	DonateLink string          `yaml:"donate_link"`
	Title      string          `yaml:"title"`
	Link       string          `yaml:"link"`
	Schedule   ShelterSchedule `yaml:"schedule"`
}

type TripToShelter struct {
	Username          string
	Shelter           *Shelter
	Date              string
	IsFirstTrip       bool
	Purpose           []string
	TripBy            string
	HowYouKnowAboutUs string
}

func NewTripToShelter() *TripToShelter {
	return &TripToShelter{}
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

	var newTripToShelter *TripToShelter

	// getting message
	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
			log.Printf("lastMessage: %s", lastMessage)

			var msgObj tgbotapi.MessageConfig
			//check for commands
			switch update.Message.Text {
			case "/start":
				log.Println("[walkthedog_bot]: Send start message")
				msgObj = startMessage(update.Message.Chat.ID)
				msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msgObj)
				lastMessage = "/start"
			case "/go_shelter":
				log.Println("[walkthedog_bot]: Send appointmentOptionsMessage message")
				lastMessage = goShelterCommand(bot, &update)
			case "/choose_shelter":
				lastMessage = chooseShelterCommand(bot, &update, newTripToShelter, &shelters)
			case "/trip_dates":
				lastMessage = tripDates(bot, &update, newTripToShelter, &shelters, lastMessage)
			case "/masterclass":
				log.Println("[walkthedog_bot]: Send masterclass")
				msgObj = masterclass(update.Message.Chat.ID)
				bot.Send(msgObj)
				lastMessage = "/masterclass"
			case "/donation":
				log.Println("[walkthedog_bot]: Send donation")
				lastMessage = donationCommand(bot, update.Message.Chat.ID)
			case "/donation_shelter_list":
				log.Println("[walkthedog_bot]: Send donationShelterList")
				msgObj = donationShelterList(update.Message.Chat.ID, &shelters)
				bot.Send(msgObj)
				lastMessage = "/donation_shelter_list"
			default:
				switch lastMessage {
				case "/go_shelter":
					if update.Message.Text == "Приют" {
						lastMessage = chooseShelterCommand(bot, &update, newTripToShelter, &shelters)
					} else if update.Message.Text == "Время" {
						lastMessage = tripDates(bot, &update, newTripToShelter, &shelters, lastMessage)
					} else {
						Error(bot, &update, newTripToShelter, "Нажмите кноку \"Приют\" или \"Время\"")
						lastMessage = goShelterCommand(bot, &update)
						break
					}
				// when shelter was chosen next step to chose date
				case "/choose_shelter":
					if newTripToShelter == nil {
						newTripToShelter = NewTripToShelter()
					}
					shelter, err := shelters.getShelterByNameID(update.Message.Text)

					if err != nil {
						Error(bot, &update, newTripToShelter, err.Error())
						chooseShelterCommand(bot, &update, newTripToShelter, &shelters)
						break
					}
					newTripToShelter.Shelter = shelter

					log.Println("[walkthedog_bot]: Send whichDate question")
					msgObj = whichDate(update.Message.Chat.ID, shelter)
					bot.Send(msgObj)
					lastMessage = "/choose_date"
				case "/choose_date":
					lastMessage = isFirstTripCommand(bot, &update, newTripToShelter)
				case "/is_first_trip":
					lastMessage, err = tripPurposeCommand(bot, &update, newTripToShelter)
					if err != nil {
						Error(bot, &update, newTripToShelter, err.Error())
						lastMessage = isFirstTripCommand(bot, &update, newTripToShelter)
					}
				case "/trip_purpose":
					Error(bot, &update, newTripToShelter, "Выберите цели поездки и нажмите кнопку голосовать")
				case "/how_you_know_about_us":
					Error(bot, &update, newTripToShelter, "Расскажите как вы о нас узнали")
				default:
					log.Println("[walkthedog_bot]: Unknown command")

					message := "Не понимаю 🐶 Попробуй /start"
					msgObj := tgbotapi.NewMessage(update.Message.Chat.ID, message)
					bot.Send(msgObj)
					lastMessage = "/choose_date"
				}
			}
		} else if update.Poll != nil {
			//log.Printf("[%s]: %s", update.FromChat().FirstName, "save poll id")
			//polls[update.Poll.ID] = update.FromChat().ID
		} else if update.PollAnswer != nil {
			log.Printf("[%s]: %s", update.PollAnswer.User.UserName, update.PollAnswer.OptionIDs)
			log.Printf("lastMessage: %s", lastMessage)
			switch lastMessage {
			case "/trip_purpose":
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.Purpose = append(newTripToShelter.Purpose, purposes[option])
				}

				lastMessage = howYouKnowAboutUsCommand(bot, &update, newTripToShelter)
			case "/how_you_know_about_us":
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.HowYouKnowAboutUs = sources[option]
					break
				}
				summaryCommand(bot, &update, newTripToShelter)
				lastMessage = donationCommand(bot, polls[update.PollAnswer.PollID])
			}
		}
		log.Println("[trip_state]: ", newTripToShelter)
	}
}

func goShelterCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) string {
	msgObj := appointmentOptionsMessage(update.Message.Chat.ID)
	bot.Send(msgObj)
	return "/go_shelter"
}

func chooseShelterCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter, shelters *SheltersList) string {
	if newTripToShelter == nil {
		newTripToShelter = NewTripToShelter()
	}
	log.Println("[walkthedog_bot]: Send whichShelter question")
	msgObj := whichShelter(update.Message.Chat.ID, shelters)
	bot.Send(msgObj)
	return "/choose_shelter"
}

func isFirstTripCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter) string {
	newTripToShelter.Date = update.Message.Text
	msgObj := isFirstTrip(update.Message.Chat.ID)
	bot.Send(msgObj)
	return "/is_first_trip"
}

func tripPurposeCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter) (string, error) {
	if update.Message.Text == "Да" {
		newTripToShelter.IsFirstTrip = true
	} else if update.Message.Text == "Нет" {
		newTripToShelter.IsFirstTrip = false
	} else {
		return "/is_first_trip", errors.New("Доступные ответы \"Да\" и \"Нет\"")
	}

	msgObj := tripPurpose(update.Message.Chat.ID)

	responseMessage, _ := bot.Send(msgObj)
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID

	return "/trip_purpose", nil
}

func howYouKnowAboutUsCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter) string {
	log.Println("-----------------------",
		polls,
		update.PollAnswer.PollID,
	)
	msgObj := howYouKnowAboutUs(polls[update.PollAnswer.PollID])
	responseMessage, _ := bot.Send(msgObj)
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID
	return "/how_you_know_about_us"
}

func summaryCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter) string {
	msgObj := summary(polls[update.PollAnswer.PollID], newTripToShelter)
	bot.Send(msgObj)
	return "/summary"
}

func donationCommand(bot *tgbotapi.BotAPI, chatId int64) string {
	msgObj := donation(chatId)
	bot.Send(msgObj)
	return "/donation"
}

func tripDates(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter, shelters *SheltersList, lastMessage string) string {
	if newTripToShelter == nil {
		newTripToShelter = NewTripToShelter()
		if lastMessage == "/choose_shelter" {
			panic("change it if I use it")
			shelter, err := shelters.getShelterByNameID(update.Message.Text)

			if err != nil {
				Error(bot, update, newTripToShelter, err.Error())
				chooseShelterCommand(bot, update, newTripToShelter, shelters)
			}
			newTripToShelter.Shelter = shelter
		} else if lastMessage == "/go_shelter" {

		}
	}
	log.Println("[walkthedog_bot]: Send whichDate question")
	msgObj := whichDate(update.Message.Chat.ID, nil)
	bot.Send(msgObj)
	return "/trip_dates"
}

func (shelters SheltersList) getShelterByNameID(name string) (*Shelter, error) {
	dotPosition := strings.Index(name, ".")
	if dotPosition == -1 {
		//log.Println(errors.New(fmt.Sprintf("message %s don't contain dot", name)))
		return nil, errors.New("Не похоже на название приюта")
	}
	shelterId, err := strconv.Atoi(name[0:dotPosition])
	if err != nil {
		log.Println(err)
		return nil, errors.New("Не похоже на название приюта")
	}
	//log.Println("id part", update.Message.Text[0:strings.Index(update.Message.Text, ".")])
	shelter, ok := shelters[shelterId]
	if !ok {
		log.Println(errors.New(fmt.Sprintf("shelter name \"%s\", extracted id=\"%d\" is not found", name, shelterId)))
		return nil, errors.New("Не похоже на название приюта")
	}

	return shelter, nil
}

func Error(bot *tgbotapi.BotAPI, update *tgbotapi.Update, newTripToShelter *TripToShelter, errMessage string) string {
	log.Println("[walkthedog_bot]: Send ERROR")
	if errMessage == "" {
		errMessage = "Error"
	}
	msgObj := errorMessage(update.Message.Chat.ID, errMessage)
	bot.Send(msgObj)
	return "/error"
}

// getConfig returns config by environment.
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
	message := `TODO masterclass message`
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
	message := `- /go_shelter Записаться на выезд в приют
- /masterclass Записаться на мастерклас
- /donation Сделать пожертвование`
	msgObj := tgbotapi.NewMessage(chatId, message)

	return msgObj
}

// appointmentOptionsMessage returns message with 2 options.
func appointmentOptionsMessage(chatId int64) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "Вы можете выбрать время выезда или конкретный приют"
	msgObj := tgbotapi.NewMessage(chatId, message)

	var numericKeyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Время"),
		tgbotapi.NewKeyboardButton("Приют"),
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

// whichDate returns message with question "Which Date you want go" and button options
func whichDate(chatId int64, shelter *Shelter) tgbotapi.MessageConfig {
	//ask about what shelter are you going
	message := "Выберите дату выезда:"
	msgObj := tgbotapi.NewMessage(chatId, message)

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	fmt.Println(firstOfMonth)

	var numericKeyboard tgbotapi.ReplyKeyboardMarkup
	log.Println(shelter)
	log.Println(*shelter)
	log.Println(shelter.Schedule)

	if shelter.Schedule.Type == "regularly" {

		scheduleWeek := shelter.Schedule.Details[0]
		scheduleDay := shelter.Schedule.Details[1]
		scheduleTime := shelter.Schedule.TimeStart
		var dateButtons [][]tgbotapi.KeyboardButton
		for i := 0; i < 6; i++ {
			month := time.Month(int(time.Now().Month()) + i)

			day := calculateDay(scheduleDay, scheduleWeek, month)
			//TODO:display on russian lang
			log.Println(day.Format("Mon 2 Jan") + " " + scheduleTime)
			if i == 0 && time.Now().Day() > day.Day() {
				continue
			}
			buttonRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(day.Format("Mon 02.01.2006") + " " + scheduleTime),
			)
			dateButtons = append(dateButtons, buttonRow)

		}
		numericKeyboard = tgbotapi.NewReplyKeyboard(dateButtons...)
		//}
	} else if shelter.Schedule.Type == "everyday" {
		//TODO: finish everyday type
	}

	msgObj.ReplyMarkup = numericKeyboard
	return msgObj
}

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

func tripPurpose(chatId int64) tgbotapi.SendPollConfig {
	message := "Цель поездки"
	options := purposes
	msgObj := tgbotapi.NewPoll(chatId, message, options...)
	msgObj.AllowsMultipleAnswers = true
	msgObj.IsAnonymous = false
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return msgObj
}

func howYouKnowAboutUs(chatId int64) tgbotapi.SendPollConfig {
	message := "Как вы о нас узнали?"
	options := sources
	msgObj := tgbotapi.NewPoll(chatId, message, options...)
	msgObj.IsAnonymous = false
	msgObj.AllowsMultipleAnswers = false
	return msgObj
}

func summary(chatId int64, newTripToShelter *TripToShelter) tgbotapi.MessageConfig {

	message := fmt.Sprintf(`Регистрация прошла успешно.
	
Информация о событии
Выезд в приют: <a href="%s">%s</a>
Дата и время: %s
Место проведения: %s (точный адрес приюта отправляется в чат после регистрации)

За 5-7 дней до выезда мы пришлем вам ссылку для добавления в Whats App чат, где расскажем все детали и ответим на вопросы. До встречи!

Напоминаем, что участие в выезде в приют является бесплатным. При этом вы можете сделать добровольное пожертвование.
`, newTripToShelter.Shelter.Link,
		newTripToShelter.Shelter.Title,
		newTripToShelter.Date,
		newTripToShelter.Shelter.Address)
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML

	return msgObj
}

// donation set donation text and message options and returns MessageConfig
func donation(chatId int64) tgbotapi.MessageConfig {
	message :=
		`Добровольное пожертвование в 500 рублей и более осчастливит 1 собаку (500 рублей = 2 недели питания одной собаки в приюте). Все собранные деньги будут переданы в приют.
📍 /donation_shelter_list - пожертвовать в конкретный приют
📍 Перевод по номеру телефона +7 916 085 1342 (Михайлов Дмитрий) - укажите название приюта.
📍 Перевод по номеру карты 4377 7314 2793 9183 (Тинькоф), 5336 6903 0880 6803 (Сбербанк), 5559 4928 1417 6700 (Альфабанк) - укажите название приюта.
📍 <a href="https://yoomoney.ru/to/410015848442299">Яндекс.Деньги</a>
📍 <a href="https://www.paypal.com/paypalme/monblan">PayPal</a>
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
