package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	botToken := os.Getenv("BOT_TOKEN")
	weatherKey := os.Getenv("WEATHER_KEY")

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Бот %s запущен", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			text := update.Message.Text

			switch text {
			case "/start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "👋 Привет! Хочешь узнать погоду?")
				msg.ReplyMarkup = yesNoKeyboard()
				bot.Send(msg)

			case "Да":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🌍 Введите название города:")
				bot.Send(msg)

			case "Нет":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хорошо! Обращайся, если передумаешь 🙂")
				bot.Send(msg)

			default:
				// Обработка города
				city := strings.TrimSpace(text)
				result, err := getWeather(city, weatherKey)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❗ "+err.Error())
					bot.Send(msg)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
				bot.Send(msg)
			}
		}
	}
}

func yesNoKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Да"),
			tgbotapi.NewKeyboardButton("Нет"),
		),
	)
}

func getWeather(city, apiKey string) (string, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=ru",
		city, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к API: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("город не найден или ошибка API: %s", string(body))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("ошибка разбора ответа: %v", err)
	}

	main, ok1 := data["main"].(map[string]interface{})
	weatherArr, ok2 := data["weather"].([]interface{})
	if !ok1 || !ok2 || len(weatherArr) == 0 {
		return "", fmt.Errorf("ошибка получения данных о погоде")
	}

	weather := weatherArr[0].(map[string]interface{})
	temp := main["temp"].(float64)
	description := weather["description"].(string)

	return fmt.Sprintf("🌤 В городе %s сейчас %.1f°C, %s.", city, temp, description), nil
}
