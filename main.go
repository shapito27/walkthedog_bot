package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"walkthedog/internal/dates"
	sheet "walkthedog/internal/google/sheet"
	"walkthedog/internal/models"

	"github.com/davecgh/go-spew/spew"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/patrickmn/go-cache"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Environment string
	AdminChatId int64
	Google      *models.Google
	Cache       *cache.Cache
	Bot         *tgbotapi.BotAPI
}

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
	commandSendUserContact    = "/provide_user_contact"
	commandSummaryShelterTrip = "/summary_shelter_trip"

	// System
	commandRereadShelters   = "/reread_shelters"
	commandRereadConfigFile = "/reread_app_config"
	commandUpdateGoogleAuth = "/update_google_auth"
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

const (
	cacheDir      = "cache/"
	cacheFileName = "cache.dat"
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
	"Нашел в интернете",
	"Telegram",
	"WhatsApp",
	"Вконтакте",
	"Другие социальные сети",
	"Авито/Юла",
	"Мосволонтер",
	"Знаю вас уже давно",
	"Другой вариант",
}

// statePool store all chat states
var statePool = make(map[int64]*models.State)

//TODO: remove poll_id after answer.
// polls stores poll_id => chat_id
var polls = make(map[string]int64)

// SheltersList represents list of Shelters
type SheltersList map[int]*models.Shelter

// NewTripToShelter initializes new object for storing user's trip information.
func NewTripToShelter(userName string) *models.TripToShelter {
	return &models.TripToShelter{
		Username: userName,
	}
}

//var tempCacheFileName string
var app AppConfig

func main() {
	/*
			@TODO cache mechanism:
			1. save to temp files. No need gob.Register!
			2. no need second cache var "chats_have_trips" because each file cache comes from one session
			3. if data sent to google sheet, then remove temprorary file
			4. after app fall down, I need load files, sent to google sheet, remove temp file one by one.

		-------------------------
		No, not like this.
		Because I have cache in memory. If app is not fall down I should restore data from inmemory cache.
		In case app fall down I need store all value to file before:
			- panic. How to catch example: @see https://stackoverflow.com/questions/55923271/how-to-write-to-the-console-before-a-function-panics
															defer func() {
																r := recover()
																if r != nil {
																	fmt.Println("Recovered", r)
																}
															}()

															panic("panic here")
			- manual app termination. https://stackoverflow.com/questions/37798572/exiting-go-applications-gracefully

		So,
			1. Add command send cached trips. When I got new token, need to reread cache and try to send data. If done clear from cache
			2. I'll save Data to file before panic or exit
				1) not sent trips details(from cache)
				2) ids of not sent trips(from cache)
				3) polls(memory)
				4) statePool(memory)
				5) not finished registration to the trip(memory)
			3. Make it possible to restore data when start app
			4.
	*/

	c, err := initCache()
	if err != nil {
		log.Panic(err)
	}
	app.Cache = c
	/* trip := models.TripToShelter{
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
				Details:         []int{4, 54},
				DatesExceptions: []string{"434", "sdf"},
				TimeStart:       "1:1",
				TimeEnd:         "34:5",
			},
		},
		Date:              "3434",
		IsFirstTrip:       true,
		Purpose:           []string{"dfdfdf", "df"},
		TripBy:            "dfdf",
		HowYouKnowAboutUs: "dfdfdf",
	} */
	//c.Set("test", &trip, cache.NoExpiration)
	/* saveTripToCache(c, &trip, 3453453453453) */
	/* spew.Dump(c.Get("chats_have_trips"))
	spew.Dump(c.Get("3453453453453")) */

	/* c, err = initCache()
	if err != nil {
		log.Panic(err)
	} */

	//panic("end")
	config, err := getConfig()
	if err != nil {
		log.Panic(err)
	}

	app.Environment = config.TelegramEnvironment.Environment
	telegramConfig := config.TelegramEnvironment.TelegramConfig[app.Environment]

	/* app.Environment = curEnvironment */
	//app.Administration = config.Administration
	app.Google = config.Google

	// @TODO remove bot var. Use app.Bot
	// bot init
	app.Bot, err = tgbotapi.NewBotAPI(telegramConfig.APIToken)
	if err != nil {
		log.Panic(err)
	}

	if app.Environment == developmentEnv {
		app.Bot.Debug = true
	}

	log.Printf("Authorized on account %s", app.Bot.Self.UserName)

	// set how often check for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = telegramConfig.Timeout

	updates := app.Bot.GetUpdatesChan(u)

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
		var isAdmin bool

		var adminChatId int64 = 0
		if config.Administration.Admin == "" {
			log.Println("config.Administration.Admin is empty!")
		} else {
			adminChatIdTmp, err := strconv.Atoi(config.Administration.Admin)
			if err != nil {
				log.Println(err)
			}
			// @TODO remove adminChatId
			adminChatId = int64(adminChatIdTmp)
			app.AdminChatId = int64(adminChatIdTmp)
		}

		// If we got a message
		if update.Message != nil {
			isAdmin = update.Message.Chat.ID == adminChatId
			log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)
			log.Printf("lastMessage: %s", lastMessage)

			var msgObj tgbotapi.MessageConfig
			//check for commands
			switch update.Message.Text {
			case "/sh":
				//for testing
				spew.Dump("start")

				spew.Dump("end")
			case commandStart:
				log.Println("[walkthedog_bot]: Send start message")
				msgObj = startMessage(chatId)
				msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				app.Bot.Send(msgObj)
				lastMessage = commandStart
			case commandGoShelter:
				log.Println("[walkthedog_bot]: Send appointmentOptionsMessage message")
				lastMessage = app.goShelterCommand(&update)
			case commandChooseShelter:
				lastMessage = app.chooseShelterCommand(&update, &shelters)
			case commandTripDates:
				lastMessage = app.tripDatesCommand(&update, newTripToShelter, &shelters, lastMessage)
			case commandMasterclass:
				log.Println("[walkthedog_bot]: Send masterclass")
				msgObj = masterclass(chatId)
				app.Bot.Send(msgObj)
				lastMessage = commandMasterclass
			case commandDonation:
				log.Println("[walkthedog_bot]: Send donation")
				lastMessage = app.donationCommand(chatId)
			case commandDonationShelterList:
				log.Println("[walkthedog_bot]: Send donationShelterList")
				msgObj = donationShelterList(chatId, &shelters)
				app.Bot.Send(msgObj)
				lastMessage = commandDonationShelterList
			//system commands
			case commandRereadShelters:
				if isAdmin {
					// getting shelters again
					shelters, err = getShelters()
					if err != nil {
						log.Panic(err)
					}
					log.Println("[walkthedog_bot]: Shelters list was reread")
					lastMessage = commandRereadShelters
				}
			case commandRereadConfigFile:
				if isAdmin {
					config, err = getConfig()
					if err != nil {
						log.Panic(err)
					}
					log.Println("[walkthedog_bot]: App config was reread")
					lastMessage = commandRereadConfigFile
				}
			case commandUpdateGoogleAuth:
				if isAdmin {
					//googleSpreadsheet := sheet.NewGoogleSpreadsheet(*config.Google)

					var message string
					authURL, err := sheet.RequestAuthCodeURL()
					if err != nil {
						message = err.Error()
					} else {
						message = authURL + " \r\n Необходимо перейти по ссылке дать разрешения в гугле, после редиректа скопировать ссылку и отправить боту"
					}
					msgObj := tgbotapi.NewMessage(adminChatId, message)
					app.Bot.Send(msgObj)
					lastMessage = commandUpdateGoogleAuth
				}
			default:
				switch lastMessage {
				case commandGoShelter:
					if update.Message.Text == chooseByShelter {
						lastMessage = app.chooseShelterCommand(&update, &shelters)
					} else if update.Message.Text == chooseByDate {
						//lastMessage = tripDatesCommand(&update, newTripToShelter, &shelters, lastMessage)
						app.ErrorFrontend(&update, "Запись по Времени пока не доступна 😥")
						lastMessage = app.goShelterCommand(&update)
						break
					} else {
						app.ErrorFrontend(&update, fmt.Sprintf("Нажмите кноку \"%s\" или \"%s\"", chooseByDate, chooseByShelter))
						lastMessage = app.goShelterCommand(&update)
						break
					}
				// when shelter was chosen next step to chose date
				case commandChooseShelter:
					if newTripToShelter == nil {
						newTripToShelter = NewTripToShelter(update.Message.From.UserName)
					}
					shelter, err := shelters.getShelterByNameID(update.Message.Text)

					if err != nil {
						app.ErrorFrontend(&update, err.Error())
						app.chooseShelterCommand(&update, &shelters)
						break
					}
					newTripToShelter.Shelter = shelter
					if shelter.ID == "8" {
						message := `<b>Зоотель "Лемур" находится в г. Воскресенск на юго-востоке от Москвы (80 км от МКАД по Новорязанское шоссе).</b>
В этом районе нет приютов, а только стационары двух ветклиник. Здесь содержатся до 30 бездомных кошек и до 8 собак. Большинство имеют те или иные заболевания и травмы. В зоотеле животные проходят полный курс лечения и стерилизации. Вот примерная точка (https://yandex.ru/maps/-/CCUNFHxqCB) на город Воскресенск.
						
Мы сейчас не организуем групповые выезды туда, так как на передержке обычно немного собак, с которыми могло бы погулять большое количество людей. 
						
При этом любой человек может самостоятельно приехать в Лемур. Также в Лемуре стоит «Корзина добра» для сбора помощи бездомным животным Воскресенского района. 
						 
Приехать в Лемур можно в любой день с 10 до 18. 
Перед тем как поехать - напишите нам в чат @walkthedog_lemur c датой когда хотите приехать (в ответ мы пришлем все детали).
						
Подробнее про Лемур: walkthedog.ru/lemur`
						msgObj := tgbotapi.NewMessage(chatId, message)

						msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						msgObj.ParseMode = tgbotapi.ModeHTML
						app.Bot.Send(msgObj)
						break
					}
					log.Println("[walkthedog_bot]: Send whichDate question")
					msgObj = whichDate(chatId, shelter)
					app.Bot.Send(msgObj)
					lastMessage = commandChooseDates
				case commandChooseDates:
					if isTripDateValid(update.Message.Text, newTripToShelter) {
						lastMessage = app.isFirstTripCommand(&update, newTripToShelter)
					} else {
						app.ErrorFrontend(&update, "Кажется вы ошиблись с датой 🤔")
						lastMessage = app.tripDatesCommand(&update, newTripToShelter, &shelters, lastMessage)
					}
				case commandIsFirstTrip:
					lastMessage, err = app.tripPurposeCommand(&update, newTripToShelter)
					if err != nil {
						app.ErrorFrontend(&update, err.Error())
						if isTripDateValid(update.Message.Text, newTripToShelter) {
							lastMessage = app.isFirstTripCommand(&update, newTripToShelter)
						} else {
							lastMessage = app.tripDatesCommand(&update, newTripToShelter, &shelters, lastMessage)
						}
					}
				case commandSendUserContact:
					// set username.
					newTripToShelter.Username = update.Message.Text
					app.registrationFinished(chatId, newTripToShelter)
				case commandTripPurpose:
					app.ErrorFrontend(&update, "Выберите цели поездки и нажмите кнопку голосовать")
				case commandTripBy:
					app.ErrorFrontend(&update, "Расскажите как добираетесь до приюта")
				case commandHowYouKnowAboutUs:
					app.ErrorFrontend(&update, "Расскажите как вы о нас узнали")
				case commandUpdateGoogleAuth:
					if isAdmin {
						//extract code from url
						u, err := url.Parse(update.Message.Text)
						if err != nil {
							lastMessage = app.ErrorFrontend(&update, err.Error())
							break
						}
						m, err := url.ParseQuery(u.RawQuery)
						if err != nil {
							lastMessage = app.ErrorFrontend(&update, err.Error())
							break
						}
						/* // @TODO send request for auth again (probably need to remove token.json first)
						e := os.Remove("token.json")
						if e != nil {
							log.Fatal(e)
						} */
						// save new token by parsed auth code
						err = sheet.AuthorizationCodeToToken(m["code"][0])
						if err != nil {
							lastMessage = app.ErrorFrontend(&update, err.Error())
							break
						}
						message := "G.Sheet токен авторизации обновлен"
						msgObj := tgbotapi.NewMessage(adminChatId, message)
						app.Bot.Send(msgObj)

						//@TODO try to send cached trips
						app.sendCachedTripsToGSheet()
					}
				default:
					log.Println("[walkthedog_bot]: Unknown command")

					message := "Не понимаю 🐶 Попробуй " + commandStart
					msgObj := tgbotapi.NewMessage(chatId, message)
					app.Bot.Send(msgObj)
					lastMessage = commandChooseDates
				}
			}
		} else if update.Poll != nil {
			//log.Printf("[%s]: %s", update.FromChat().FirstName, "save poll id")
			//polls[update.Poll.ID] = update.FromChat().ID
		} else if update.PollAnswer != nil {
			isAdmin = update.PollAnswer.User.UserName == config.Administration.Admin
			log.Printf("[%s]: %v", update.PollAnswer.User.UserName, update.PollAnswer.OptionIDs)
			log.Printf("lastMessage: %s", lastMessage)

			switch lastMessage {
			case commandTripPurpose:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.Purpose = append(newTripToShelter.Purpose, purposes[option])
				}

				lastMessage = app.tripByCommand(&update, newTripToShelter)
			case commandTripBy:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.TripBy = tripByOptions[option]
					break
				}
				lastMessage = app.howYouKnowAboutUsCommand(&update, newTripToShelter)
			case commandHowYouKnowAboutUs:
				for _, option := range update.PollAnswer.OptionIDs {
					newTripToShelter.HowYouKnowAboutUs = sources[option]
					break
				}

				// if user dont set username
				if update.PollAnswer.User.UserName == "" {
					lastMessage = app.askForContactCommand(polls[update.PollAnswer.PollID])
					break
				}

				app.registrationFinished(chatId, newTripToShelter)
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
func (app *AppConfig) goShelterCommand(update *tgbotapi.Update) string {
	msgObj := appointmentOptionsMessage(update.Message.Chat.ID)
	app.Bot.Send(msgObj)
	return commandGoShelter
}

// chooseShelterCommand prepares message about available shelters and then sends it and returns last command.
func (app *AppConfig) chooseShelterCommand(update *tgbotapi.Update, shelters *SheltersList) string {
	log.Println("[walkthedog_bot]: Send whichShelter question")
	msgObj := whichShelter(update.Message.Chat.ID, shelters)
	app.Bot.Send(msgObj)
	return commandChooseShelter
}

// isFirstTripCommand prepares message with question "is your first trip?" and then sends it and returns last command.
func (app *AppConfig) isFirstTripCommand(update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	newTripToShelter.Date = update.Message.Text
	msgObj := isFirstTrip(update.Message.Chat.ID)
	app.Bot.Send(msgObj)
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
func (app *AppConfig) tripPurposeCommand(update *tgbotapi.Update, newTripToShelter *models.TripToShelter) (string, error) {
	if update.Message.Text == "Да" {
		newTripToShelter.IsFirstTrip = true
	} else if update.Message.Text == "Нет" {
		newTripToShelter.IsFirstTrip = false
	} else {
		return commandIsFirstTrip, errors.New("доступные ответы \"Да\" и \"Нет\"")
	}

	msgObj := tripPurpose(update.Message.Chat.ID)

	responseMessage, err := app.Bot.Send(msgObj)
	if err != nil {
		log.Fatalln(err)
	}
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID

	return commandTripPurpose, nil
}

// tripByCommand prepares poll with question about how he going to come to shelter and then sends it and returns last command.
func (app *AppConfig) tripByCommand(update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	msgObj := tripBy(polls[update.PollAnswer.PollID])
	responseMessage, err := app.Bot.Send(msgObj)
	if err != nil {
		log.Fatalln(err)
	}
	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID
	return commandTripBy
}

// howYouKnowAboutUsCommand prepares poll with question about where did you know about us and then sends it and returns last command.
func (app *AppConfig) howYouKnowAboutUsCommand(update *tgbotapi.Update, newTripToShelter *models.TripToShelter) string {
	msgObj := howYouKnowAboutUs(polls[update.PollAnswer.PollID])
	responseMessage, err := app.Bot.Send(msgObj)

	if err != nil {
		//@TODO if i got error here I don't have chat id in response(but have PollAnswer.PollID and PollAnswer.User). So need to get chat id and display error that bot is broken.
		log.Fatalln(err)
		/* app.ErrorFrontend(update, newTripToShelter, "У бота временные проблемы 😥")
		return commandError */
	}

	polls[responseMessage.Poll.ID] = responseMessage.Chat.ID
	return commandHowYouKnowAboutUs
}

// summaryCommand prepares message with summary and then sends it and returns last command.
func (app *AppConfig) summaryCommand(chatId int64, newTripToShelter *models.TripToShelter) string {
	msgObj := summary(chatId, newTripToShelter)
	app.Bot.Send(msgObj)
	return commandSummaryShelterTrip
}

// donationCommand prepares message with availabele ways to dontate us or shelters and then sends it and returns last command.
func (app *AppConfig) donationCommand(chatId int64) string {
	msgObj := donation(chatId)
	app.Bot.Send(msgObj)
	return commandDonation
}

// tripDatesCommand prepares message with availabele dates to go to shelters and then sends it and returns last command.
func (app *AppConfig) tripDatesCommand(update *tgbotapi.Update, newTripToShelter *models.TripToShelter, shelters *SheltersList, lastMessage string) string {
	if newTripToShelter == nil {
		message := "По времени записаться пока нельзя :("
		msgObj := tgbotapi.NewMessage(update.Message.Chat.ID, message)
		app.Bot.Send(msgObj)
		return commandGoShelter
		/* newTripToShelter = NewTripToShelter(update.Message.From.UserName)
		if lastMessage == commandChooseShelter {
			panic("change it if I use it")
			shelter, err := shelters.getShelterByNameID(update.Message.Text)

			if err != nil {
				app.ErrorFrontend(update, newTripToShelter, err.Error())
				chooseShelterCommand(update, shelters)
			}
			newTripToShelter.Shelter = shelter
		} else if lastMessage == commandGoShelter {

		} */
	}
	log.Println("[walkthedog_bot]: Send whichDate question")
	msgObj := whichDate(update.Message.Chat.ID, newTripToShelter.Shelter)
	app.Bot.Send(msgObj)
	return commandChooseDates
}

// askForContactCommand prepares message with question about user contact and returns last command.
func (app *AppConfig) askForContactCommand(chatId int64) string {
	msgObj := askContact(chatId)
	app.Bot.Send(msgObj)
	return commandSendUserContact
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
func (app *AppConfig) ErrorFrontend(update *tgbotapi.Update, errMessage string) string {
	log.Println("[walkthedog_bot]: Send ERROR")
	if errMessage == "" {
		errMessage = "Error"
	}
	msgObj := errorMessage(update.Message.Chat.ID, errMessage)
	app.Bot.Send(msgObj)
	return commandError
}

// getConfig returns config by environment.
func getConfig() (*models.ConfigFile, error) {
	yamlFile, err := ioutil.ReadFile("configs/app.yml")
	if err != nil {
		return nil, err
	}

	var configFile models.ConfigFile
	err = yaml.Unmarshal(yamlFile, &configFile)
	if err != nil {
		return nil, err
	}
	return &configFile, nil
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
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

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
			tgbotapi.NewKeyboardButton(fmt.Sprintf("%s. %s", (*shelters)[i].ID, (*shelters)[i].LongTitle)),
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
			formatedDate := day.Format("02.01.2006")
			isException := false
			//check for exceptions
			for _, v := range shelter.Schedule.DatesExceptions {
				if v == formatedDate {
					isException = true
					break
				}
			}
			if isException {
				continue
			}

			shedule = append(shedule, dates.WeekDaysRu[day.Weekday()]+" "+formatedDate+" "+scheduleTime)

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
	message := fmt.Sprintf(`Регистрация прошла успешно.
	
ℹ️ Информация о событии
Выезд в приют: <a href="%s">%s</a>
Дата и время: %s

❤️ Напоминаем, что участие в выезде в приют является бесплатным. При этом вы можете сделать добровольное пожертвование.

💬 За 5 дней до выезда мы добавим вас в чат, где можно будет узнать все детали о выезде в приют включая адрес, как доехать, что взять, потребности приюта и задать вопросы.

Если у вас появятся вопросы до добавления в чат - пишите @walkthedog_support
`, newTripToShelter.Shelter.Link,
		newTripToShelter.Shelter.Title,
		newTripToShelter.Date)
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML

	return msgObj
}

// donation set donation text and message options and returns MessageConfig.
func donation(chatId int64) tgbotapi.MessageConfig {
	message :=
		`Добровольное пожертвование в 500 рублей и более осчастливит 1 собаку (500 рублей = 2 недели питания одной собаки в приюте). На собранные пожертвования мы строим теплые будки для приютов, покупаем корм и медикаменты.

📍 /donation_shelter_list - пожертвовать в конкретный приют

📍 Перевод по номеру телефона +79160851342 (Михайлов Дмитрий) - укажите "пожертвование"

📍 Сбор пожертвований через <a href="https://www.tinkoff.ru/sl/72xLdsZQp6">Тинькоф банк</a>

📍 <a href="https://yoomoney.ru/to/410015848442299">Яндекс.Деньги</a>
`
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML
	msgObj.DisableWebPagePreview = true
	msgObj.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	return msgObj
}

// askContact returns object including message text with summary of user's answers and other message config.
func askContact(chatId int64) tgbotapi.MessageConfig {
	message := fmt.Sprintf(`Регистрация почти завершена 👍

Но мы не можем определить ваше имя пользователя Телеграм. 
Пожалуйста напишите в следующем сообщении email или номер телефона, чтобы мы смогли добавить вас в чат выезда в приют.

Если возникли проблемы напишите нам @walkthedog_support
`)
	msgObj := tgbotapi.NewMessage(chatId, message)
	msgObj.ParseMode = tgbotapi.ModeHTML

	return msgObj
}

func (app *AppConfig) registrationFinished(chatId int64, newTripToShelter *models.TripToShelter) string {
	app.summaryCommand(chatId, newTripToShelter)
	lastMessage := app.donationCommand(chatId)

	// generate uniq ID for trip to shelter
	date := newTripToShelter.Date
	date = date[strings.Index(date, " ")+1 : strings.Index(date, " ")+10]
	newTripToShelter.ID = date + newTripToShelter.Shelter.ShortTitle

	app.saveTripToCache(newTripToShelter, chatId)

	isTripSent := app.sendTripToGSheet(chatId, newTripToShelter)
	if !isTripSent {
		// send message to the admin. G.Sheet auth expired.
		message := "G.Sheet auth expired."
		msgObj := tgbotapi.NewMessage(app.AdminChatId, message)
		app.Bot.Send(msgObj)
	}

	return lastMessage
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

// initCache init cache based on file or creates new.
func initCache() (*cache.Cache, error) {
	c := cache.New(5*time.Hour, 10*time.Hour)

	err := c.LoadFile(cacheDir + cacheFileName)
	if err != nil {
		if err.Error() == "EOF" {
			log.Println("Cache file is empty.")
		} else {
			log.Println(err.Error())
		}
	} else {
		log.Println("cache from file")
	}

	return c, nil
}

// saveCacheToFile saves cache to file.
func saveCacheToFile(cache *cache.Cache) error {
	/* f, err := ioutil.TempFile(cacheDir, cacheFileName)
	if err != nil {
		return err
		//t.Fatal("Couldn't create cache file:", err)
	}
	fname := f.Name()
	tempCacheFileName = fname
	f.Close()
	err = cache.SaveFile(fname) */
	err := cache.SaveFile(cacheDir + cacheFileName)
	if err != nil {
		return err
	}

	return nil
}

// saveTripToCache saves trip to cache.
func (app *AppConfig) saveTripToCache(newTripToShelter *models.TripToShelter, chatId int64) {
	//save newTripToShelter pointer to the object to the cache
	app.Cache.Set(newTripToShelter.ID, *newTripToShelter, cache.NoExpiration)

	// take all trips registrations that were not save to G.Sheets from cache
	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}
	// add new registration chat id
	chatsWithTripsID[chatId] = append(chatsWithTripsID[chatId], newTripToShelter.ID)
	// save to the cache
	app.Cache.Set("chats_have_trips", chatsWithTripsID, cache.NoExpiration)
	/*
		no need, i'll save to file before panic or exit
		err := saveCacheToFile(ca)
		if err != nil {
			log.Printf("Unable save Cache To File: %v", err)
		} */
}

// removeTripFromCache removes trip from cache.
func (app *AppConfig) removeTripFromCache(newTripToShelterId string, chatId int64) {
	//delete newTripToShelter from cache by chatId
	app.Cache.Delete(newTripToShelterId)

	/* // take all trips registrations that were not save to G.Sheets from cache
	var tripIds []int64
	var tripIdsResult []string

	tripIdsFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		tripIds = tripIdsFromCache.([]int64)
	} */

	var tripIdsResult []string
	// remove trips from array
	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}
	tripsByChatId, ok := chatsWithTripsID[chatId]
	// exit if don't have such a key in array
	if !ok {
		return
	}
	for i, v := range tripsByChatId {
		if v == newTripToShelterId {
			if len(tripsByChatId) == 1 {
				// we need to delete trip from array by chat id and for this chat id only one trip exist. So remove chat id from map and underlying value.
				delete(chatsWithTripsID, chatId)
			} else {
				tripIdsResult = append(tripIdsResult, tripsByChatId[:i]...)
				tripIdsResult = append(tripIdsResult, tripsByChatId[i+1:]...)
				chatsWithTripsID[chatId] = tripIdsResult
			}
			break
		}
	}
	// save to the cache
	app.Cache.Set("chats_have_trips", chatsWithTripsID, cache.NoExpiration)
}

// sendTextMessage sends message
func (app *AppConfig) sendTextMessage(chatId int64, message string) (tgbotapi.Message, error) {
	msgObj := tgbotapi.NewMessage(chatId, message)
	return app.Bot.Send(msgObj)
}

// sendCachedTripsToGSheet
func (app *AppConfig) sendCachedTripsToGSheet() {
	chatsWithTripsID := make(map[int64][]string)
	chatsWithTripsIDFromCache, found := app.Cache.Get("chats_have_trips")
	if found {
		chatsWithTripsID = chatsWithTripsIDFromCache.(map[int64][]string)
	}
	for chatId, TripsIDs := range chatsWithTripsID {
		for _, v := range TripsIDs {
			var tripToShelter models.TripToShelter
			tripFromCache, found := app.Cache.Get(v)
			if found {
				tripToShelter = tripFromCache.(models.TripToShelter)
			}
			isTripSent := app.sendTripToGSheet(chatId, &tripToShelter)
			if isTripSent {
				log.Printf("Trip %s from cache sent to G.Sheet", tripToShelter.ID)
				app.removeTripFromCache(tripToShelter.ID, chatId)
			} else {
				log.Println("Can't send trip to GSheet, so strop loop")
				break
			}
		}
	}

	/* // try to find trip details by trip's id
	for _, v := range tripIds {
		//get newTripToShelter pointer
		var tripToShelter *models.TripToShelter
		tripToShelterFromCache, found := app.Cache.Get(fmt.Sprintf("%d", v))
		if found {
			tripToShelter = tripToShelterFromCache.(*models.TripToShelter)
		}

		isTripSent := app.sendTripToGSheet(v, tripToShelter)
		if isTripSent {
			app.removeTripFromCache(v)
		} else {
			log.Println("Can't send trip to GSheet, so strop loop")
			break
		}
	} */
}

// sendTripToGSheet.
func (app *AppConfig) sendTripToGSheet(chatId int64, newTripToShelter *models.TripToShelter) bool {
	savingError := false
	googleSpreadsheet, err := sheet.NewGoogleSpreadsheet(*app.Google)
	if err != nil {
		savingError = true
		log.Printf("Unable to get sheet.NewGoogleSpreadsheet: %v", err)
	}
	/*
		@INFO this code allows to save data to separate tab with
		name of trip with following format: 13.08.2022Ника, 14.08.2022Шанс.
		It checks if tab exists, it save it, otherwise it creates new tab.

		date := newTripToShelter.Date

		date = date[strings.Index(date, " ")+1 : strings.Index(date, " ")+10]
		sheetName := date + newTripToShelter.Shelter.ShortTitle

		if !savingError {
			err := googleSpreadsheet.PrepareSheetForSavingData(sheetName)
			if err != nil {
				savingError = true
				log.Printf("Unable to create sheet or add headers: %v", err)
			}
		}
	*/

	sheetName := "Trips"

	if !savingError {
		resp, err := googleSpreadsheet.SaveTripToShelter(sheetName, newTripToShelter)

		if err != nil {
			savingError = true
			log.Printf("Unable to write data to sheet: %v", err)
		}
		if resp.ServerResponse.HTTPStatusCode != 200 {
			savingError = true
			log.Printf("Response status code is not 200: %+v", resp)
		}
	}

	if !savingError {
		// because trip was saved we need to remove it from cache.
		app.removeTripFromCache(newTripToShelter.ID, chatId)
		return true
	} else {
		return false
	}
}
