package main

import (
	"errors"
	"io/ioutil"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v2"
)

const (
	productionEnv  = "production"
	testEnv        = "test"
	developmentEnv = "development"
)

type EnvironmentConfig map[string]*Config

type Config struct {
	APIToken string `yaml:"api_token"`
	Timeout  int    `yaml:"timeout"`
}

func main() {
	env := developmentEnv
	config, err := getConfig(env)
	if err != nil {
		log.Panic(err)
	}
	bot, err := tgbotapi.NewBotAPI(config.APIToken)
	if err != nil {
		log.Panic(err)
	}

	if env == developmentEnv {
		bot.Debug = true
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 20

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}

func getConfig(environment string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile("telegram.yml")
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
