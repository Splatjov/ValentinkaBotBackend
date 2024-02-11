package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {

	frontendUrl := os.Getenv("FRONTEND_URL")
	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken)

	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", "db.db")

	if err != nil {
		log.Fatal(err)
	}

	var f = urlFunctions{
		db:       db,
		bot:      bot,
		botToken: botToken,
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "*"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tg-Token", "ngrok-skip-browser-warning", "referer", "user-agent", "userID", "valentineText", "senderID", "receiverID", "valentineType", "*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/get_user_photo", f.getUserPhoto)
	r.Get("/get_user_info", f.getUserInfo)
	r.Get("/check", f.check)
	r.Get("/ping", f.ping)
	r.Post("/send_valentine", f.sendValentine)
	r.Post("/delete_valentine", f.deleteValentine)
	r.Get("/get_my_valentine", f.getMyValentine)
	r.Get("/get_valentine_info", f.getValentineInfo)

	go http.ListenAndServe(":3000", r)

	updates, _ := bot.UpdatesViaLongPolling(nil)

	defer bot.StopLongPolling()

	for update := range updates {
		log.Println(update.Message)
		if update.Message == nil {
			continue
		}
		user := telego.User{
			ID:        update.Message.Chat.ID,
			FirstName: update.Message.Chat.FirstName,
			LastName:  update.Message.Chat.LastName,
			Username:  update.Message.Chat.Username,
		}
		if user.Username != "" {
			f.updateUser(user)
		}
		if update.Message.Text == "/start" || update.Message.Text == "/restart" {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: update.Message.Chat.ID},
				Text: "Привет! Это вторая версия ВалентинкаБотика, открой MiniApp по ссылке https://t.me/valentinka_kalbot/miniapp, или нажми на кнопку под этим сообщением.\n" +
					"Что бы отправить валентинку пользователю, тебе нужно отправить его через главную кнопку рядом с вводом сообщения.\n",
				ReplyMarkup: &telego.InlineKeyboardMarkup{

					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{
							telego.InlineKeyboardButton{Text: "Кнопка для открытия бота:)", WebApp: &telego.WebAppInfo{
								URL: frontendUrl,
							}},
						},
					},
				},
			})
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: update.Message.Chat.ID},
				Text: "Теперь по правилам: нельзя отправлять более 20 обычных валентинок и более 5 'Be mine'\n" +
					"В этой версии ВалентинкаБотика ты можешь отправлять валентинки пользователям без юзернейма и тем, кто не зарегистрировался в боте, они смогут посмотреть свои валентинки и после 14 февраля, но советую поделиться ботом, что бы твои знакомые могли отправить валентинку тебе:)\n" +
					"Если хочешь сохранить анонимность, попроси @kalexina (создатель бота), и он напишет человеку, которому ты хочешь с ссылкой на ботика!\n",
				ReplyMarkup: tu.Keyboard(tu.KeyboardRow(
					tu.KeyboardButton("Отправь мне человека, которому ты хочешь отправить валентинку").WithRequestUser(&telego.KeyboardButtonRequestUser{RequestID: 1}),
				)).WithResizeKeyboard(),
			})
		} else if update.Message.UserShared == nil {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: update.Message.Chat.ID},
				Text:   "Я развился, и теперь не читаю текст:)\nПопробуй отправить валентинку кому-нибудь с новым красивым интерфейсом!\n",
				ReplyMarkup: &telego.InlineKeyboardMarkup{

					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{
							telego.InlineKeyboardButton{Text: "Кнопка для открытия бота:)", WebApp: &telego.WebAppInfo{
								URL: frontendUrl,
							}},
						},
					},
				},
			})
		}
		if update.Message.UserShared != nil {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{
					ID: update.Message.Chat.ID,
				},
				Text: "Отлично! Теперь нажми на кнопку и отправь валентинку этому человеку!",
				ReplyMarkup: &telego.InlineKeyboardMarkup{

					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{
							telego.InlineKeyboardButton{Text: fmt.Sprintf("Нажми на кнопку, что бы отправить %v валентинку", update.Message.UserShared.UserID), WebApp: &telego.WebAppInfo{
								URL: frontendUrl + "/send_valentine?userID=" + strconv.FormatInt(update.Message.UserShared.UserID, 10),
							}},
						},
					},
				},
			})
		}
	}
}
