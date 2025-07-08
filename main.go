package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
	"vpn_bot_go/config"
	"vpn_bot_go/outline"
)

var (
	bot             *tgbotapi.BotAPI
	messages        Messages
	defaultKeyboard = getDefaultKeyboard()
)

func main() {
	var err error
	messages, err = loadMessages("messages.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки сообщений: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Авторизация прошла успешно как %s", bot.Self.UserName)

	db, err := InitDB("./user_actions.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(db)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(update, cfg, db)
		}
	}
}

func handleMessage(update tgbotapi.Update, cfg *config.Config, db *sql.DB) {
	chatID := update.Message.Chat.ID
	userID := fmt.Sprintf("%d", update.Message.From.ID)
	userName := update.Message.From.UserName

	switch update.Message.Text {
	case "/start":
		sendMessage(chatID, messages.WelcomeMessage, defaultKeyboard)
		LogUserAction(update.Message.From.ID, update.Message.From.UserName, "Started /start command")
	case "OpenVPN":
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🇳🇱 Нидерланды"),
				tgbotapi.NewKeyboardButton("🇷🇺 Россия"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("⬅️ Назад"),
			),
		)
		sendMessage(chatID, "Выберите страну для OpenVPN:", keyboard)
	case "Outline":
		handleOutline(chatID, userID, userName, cfg.Outline.Server, cfg.Outline.Token)
	case "TelegramProxy":
		sendMessage(chatID, messages.ButtonTelegramProxy, defaultKeyboard)
	case "HttpProxy":
		handleHttpProxy(chatID, cfg.HttpProxy)
	case "🇳🇱 Нидерланды":
		handleOpenvpnNl(chatID, userID, cfg.OpenVPN.Script, cfg.OpenVPN.Configs)
	case "🇷🇺 Россия":
		handleOpenvpnRu(chatID, userID, cfg.RU.Server, cfg.RU.Secret)
	case "⬅️ Назад":
		sendMessage(chatID, "Вы вернулись в главное меню.", defaultKeyboard)
	default:
		sendMessage(chatID, "Неизвестная команда. Используйте кнопки меню.", defaultKeyboard)
	}
}

func handleOpenvpnNl(chatID int64, userID, scriptPath, configsPath string) {
	filePath, err := generateOpenVPNConfig(userID, scriptPath, configsPath)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка генерации OpenVPN: %v", err), defaultKeyboard)
		return
	}

	sendMessage(chatID, `Ваш OpenVPN конфиг сгенерирован. Для того что бы им воспользоваться скачайте одноименную программу OpenVPN

Рекомендуется пользоваться официальным приложением для Windows/Mac/Linux/Iphone/Android: 
ссылка для скачивания - openvpn.net/client/

Для того что бы воспользоваться конфигом, необходимо
1. Нажать на сгенерированный Файл ниже с расширением .ovpn 
2. Выбрать в меню -> открыть с  помощью -> указать программу OpenVPN

Или 
Cохранить полученный файл на устройство и добавить переносом в программу`, defaultKeyboard)

	if err := sendDocument(chatID, filePath); err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка отправки файла: %v", err), defaultKeyboard)
	}
}

func handleHttpProxy(chatID int64, httpProxy []string) {
	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))
	randIndex := randSeed.Intn(len(httpProxy))
	randProxy := strings.Split(httpProxy[randIndex], ":")
	message := "Ваши данные для использования http proxy \nСервер: cetus.shaneque.ru \nПорт: 8443 \nЛогин: " + randProxy[0] + " \nПароль: " + randProxy[1]
	sendMessage(chatID, message, defaultKeyboard)
}

func handleOpenvpnRu(chatID int64, userID, server, secret string) {
	params := url.Values{}
	params.Add("user_id", userID)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/get_config?%s", server, params.Encode()), nil)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка создания запроса: %v", err), defaultKeyboard)
		return
	}
	req.Header.Set("secret", secret)

	resp, err := client.Do(req)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка отправки запроса: %v", err), defaultKeyboard)
		return
	}
	defer resp.Body.Close()

	reader, err := io.ReadAll(resp.Body)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка чтения ответа: %v", err), defaultKeyboard)
		return
	}

	fileName := fmt.Sprintf("RU_%s.ovpn", userID)
	if err := os.WriteFile(fileName, reader, 0644); err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка записи файла: %v", err), defaultKeyboard)
		return
	}

	sendMessage(chatID, `Ваш OpenVPN конфиг сгенерирован. Для того что бы им воспользоваться скачайте одноименную программу OpenVPN

Рекомендуется пользоваться официальным приложением для Windows/Mac/Linux/Iphone/Android: 
ссылка для скачивания - openvpn.net/client/

Для того что бы воспользоваться конфигом, необходимо
1. Нажать на сгенерированный Файл ниже с расширением .ovpn 
2. Выбрать в меню -> открыть с  помощью -> указать программу OpenVPN

Или 
Cохранить полученный файл на устройство и добавить переносом в программу`, defaultKeyboard)

	if err := sendDocument(chatID, fileName); err != nil {
		sendMessage(chatID, fmt.Sprintf("Ошибка отправки файла: %v", err), defaultKeyboard)
	}
}

func handleOutline(chatID int64, userID, userName, OutlineServer, OutlineToken string) {

	link, err := outline.CheckUserExists(userID, userName, OutlineServer, OutlineToken)
	if err != nil {
		log.Printf("Ошибка проверки пользователя Outline: %v", err)
		sendMessage(chatID, "Ошибка при проверке пользователя.", defaultKeyboard)
		return
	}

	if link != "" {
		sendMessage(chatID, `Ваш  ключ доступа сгенерирован. Для того что бы им воспользоваться скачайте программу Outline Client

Рекомендуется пользоваться официальным приложением для Windows/Mac/Linux/Iphone/Android: 
ссылка для скачивания - getoutline.org/ru/get-started/

Для того что бы воспользоваться сгенерированным ключом:
1. Скачиваем приложение
2. Копируем сгенерированный ключ
3. Переходим в приложение, которое предлагает добавить сервер.
4. Нажимаем добавить - пользуемся!`, defaultKeyboard)
		sendMessage(chatID, fmt.Sprintf("```%s```", link), defaultKeyboard)
		return
	}

	link, err = outline.CreateNewUser(userID, userName, OutlineServer, OutlineToken)
	if err != nil {
		log.Printf("Ошибка создания пользователя Outline: %v", err)
		sendMessage(chatID, "Ошибка при создании пользователя.", defaultKeyboard)
		return
	}

	if link != "" {
		sendMessage(chatID, `Ваш  ключ доступа сгенерирован. Для того что бы им воспользоваться скачайте программу Outline Client

Рекомендуется пользоваться официальным приложением для Windows/Mac/Linux/Iphone/Android: 
ссылка для скачивания - getoutline.org/ru/get-started/

Для того что бы воспользоваться сгенерированным ключом:
1. Скачиваем приложение
2. Копируем сгенерированный ключ
3. Переходим в приложение, которое предлагает добавить сервер.
4. Нажимаем добавить - пользуемся!`, defaultKeyboard)
		sendMessage(chatID, fmt.Sprintf("```%s```", link), defaultKeyboard)
		return
	}

	//sendMessage(chatID, "Пользователь успешно создан.", defaultKeyboard)
}

func generateOpenVPNConfig(userID, scriptPath, configsPath string) (string, error) {
	cmd := exec.Command(scriptPath, "1", userID)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.ovpn", configsPath, userID), nil
}

func sendDocument(chatID int64, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	document := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   filePath,
		Reader: file,
	})

	if _, err := bot.Send(document); err != nil {
		return err
	}
	return nil
}

func sendMessage(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if strings.Contains(text, "```") {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}
	msg.DisableWebPagePreview = true
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func getDefaultKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("OpenVPN"),
			tgbotapi.NewKeyboardButton("Outline"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("TelegramProxy"),
			tgbotapi.NewKeyboardButton("HttpProxy"),
		),
	)
}
