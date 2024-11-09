package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func main() {
	// Загружаем текстовые сообщения
	messages, err := loadMessages("messages.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки сообщений: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизация прошла успешно как %s", bot.Self.UserName)

	// Инициализация базы данных
	db, err := InitDB("./user_actions.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			welcomeMsg := fmt.Sprintf(messages.WelcomeMessage)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMsg)

			// Организация кнопок в два ряда
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Получить OpenVPN config", "OpenVPN"),
					tgbotapi.NewInlineKeyboardButtonData("Получить TelegramProxy", "TelegramProxy"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Получить Outline config", "Outline"),
					tgbotapi.NewInlineKeyboardButtonData("Персональный HttpProxy", "HttpProxy"),
				),
			)

			_, err := bot.Send(msg)
			if err != nil {
				return
			}

			// Логируем действие пользователя
			LogUserAction(update.Message.From.ID, update.Message.From.UserName, "Started /start command")
		}

		if update.CallbackQuery != nil {
			var responseText string
			action := ""
			switch update.CallbackQuery.Data {
			case "OpenVPN":
				responseText = messages.ButtonOpenVPN
				action = "Requested OpenVPN config"
			case "Outline":
				responseText = messages.ButtonOutline
				action = "Requested Outline config"
			case "TelegramProxy":
				responseText = messages.ButtonTelegramProxy
				action = "Requested TelegramProxy config"
			case "HttpProxy":
				responseText = messages.ButtonHttpProxy
				action = "Requested HttpProxy config"
			default:
				responseText = messages.UnknownButton
				action = "Unknown button clicked"
			}

			// Логируем действие пользователя
			LogUserAction(update.CallbackQuery.From.ID, update.CallbackQuery.From.UserName, action)

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, responseText)
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
			}

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, responseText)
			_, err := bot.Send(msg)
			if err != nil {
				return
			}
		}
	}
}
