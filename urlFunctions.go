package main

import (
	"database/sql"
	"encoding/json"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"
)

type urlFunctions struct {
	bot      *telego.Bot
	db       *sql.DB
	botToken string
}

func (f urlFunctions) ping(w http.ResponseWriter, r *http.Request) {
	log.Println("ping")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	log.Println(initData)
	ok, initDataParsed := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	log.Println(initDataParsed.User.ID)
	k := telego.KeyboardButtonRequestUser{RequestID: 1}
	var params = telego.SendMessageParams{
		ChatID: telego.ChatID{
			ID: initDataParsed.User.ID,
		},
		Text: "Окей, теперь тебе нужно отправить человека, которому ты хочешь отправить валентинку, важно: если он не откроет бота, то он не увидит эту валентинку, так что делись ботом в истории, каналах и чатах",
		ReplyMarkup: tu.Keyboard(tu.KeyboardRow(
			tu.KeyboardButton("Отправь мне человека, которому ты хочешь отправить валентинку").WithRequestUser(&k),
		)).WithResizeKeyboard(),
	}
	_, err := f.bot.SendMessage(&params)
	if err != nil {
		log.Println(err)
	}
}

func (f urlFunctions) check(w http.ResponseWriter, r *http.Request) {
	log.Println("check")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	ok, _ := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (f urlFunctions) getUserInfo(w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	userid := headers.Get("userID")
	initData := headers.Get("X-Tg-Token")
	if userid == "null" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("userID is required"))
		return
	}
	userID, err := strconv.ParseInt(userid, 10, 64)
	ok, _ := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	log.Println("getUserInfo")
	w.Header().Set("Content-Type", "application/json")
	var user User
	err = f.db.QueryRow("SELECT  userID, username, name FROM users WHERE userID = (?)", userID).Scan(&user.ID, &user.Username, &user.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	jData, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println(string(jData))
}

func (f urlFunctions) getUserPhoto(w http.ResponseWriter, r *http.Request) {
	log.Println("get_user_photo")
	headers := r.Header
	userid := headers.Get("userID")
	initData := headers.Get("X-Tg-Token")
	if userid == "null" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, err := strconv.ParseInt(userid, 10, 64)
	ok, _ := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	photos, err := f.bot.GetUserProfilePhotos(&telego.GetUserProfilePhotosParams{
		UserID: userID,
		Limit:  1,
	})
	if err != nil || len(photos.Photos) == 0 || len(photos.Photos[0]) == 0 {
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Println(err)
			return
		}
	}
	photoFile, err := f.bot.GetFile(&telego.GetFileParams{
		FileID: photos.Photos[0][0].FileID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	telegramAPIURL := "https://api.telegram.org/file/bot" + f.botToken + "/" + photoFile.FilePath

	req, err := http.NewRequest(r.Method, telegramAPIURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	req.Header = make(http.Header)
	for key, values := range r.Header {
		req.Header[key] = values
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// Copy the response from the Telegram API to the original response writer
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Write(body)
	log.Println(photoFile.FilePath)
	if err != nil {
		log.Println(err)
	}

}

func (f urlFunctions) sendValentine(w http.ResponseWriter, r *http.Request) {
	log.Println("sendValentine")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	ok, dataParsed := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Convert the body bytes to a string
	valentineText := string(body)
	log.Println(valentineText)
	senderID := dataParsed.User.ID
	receiverID, err := strconv.ParseInt(headers.Get("receiverID"), 10, 64)
	log.Println(senderID)
	log.Println(receiverID)
	if err != nil || !f.findUser(senderID) {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println("406 ID is incorrect")
		return
	}
	if utf8.RuneCountInString(valentineText) > 1010 {
		w.WriteHeader(http.StatusConflict)
		log.Println("406")
		return
	}
	s := ""
	err = f.db.QueryRow("SELECT type FROM valentines WHERE senderID = (?) AND receiverID = (?)", senderID, receiverID).Scan(&s)
	if err == nil {
		w.WriteHeader(http.StatusAlreadyReported)
		log.Println("208 User already sent valentine to this user")
		return
	}
	rows, err := f.db.Query("SELECT senderID FROM valentines WHERE senderID = (?) and type = (?)", senderID, "be mine")
	count := f.dbCount(rows)

	valentineType := headers.Get("valentineType") //"default", "be mine"
	if count+1 > 5 && valentineType == "be mine" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("too much be mine")
		return
	}

	rows, err = f.db.Query("SELECT senderID FROM valentines WHERE senderID = (?) and type = (?)", senderID, "default")
	count = f.dbCount(rows)
	if count+1 > 20 && valentineType == "default" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("too much default")
		return
	}

	_, err = f.db.Exec("INSERT INTO valentines (senderID, receiverID, text, type) VALUES (?, ?, ?, ?)", senderID, receiverID, valentineText, valentineType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("500 error while inserting")
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Println("200 ok")
}

func (f urlFunctions) getValentineInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("getValentineInfo")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	ok, dataParsed := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	senderID := dataParsed.User.ID
	valentines := make([]Valentine, 0)
	rows, _ := f.db.Query("SELECT ID, receiverID, text, type FROM valentines WHERE senderID = (?)", senderID)
	countDef := 0
	countBM := 0
	for rows.Next() {
		val := Valentine{}
		err := rows.Scan(&val.ID, &val.Receiver.ID, &val.Text, &val.Type)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		if val.Type == "be mine" {
			countBM += 1
		} else {
			countDef += 1
		}
		err = f.db.QueryRow("SELECT name, username FROM users WHERE userID = (?)", val.Receiver.ID).Scan(&val.Receiver.Name, &val.Receiver.Username)
		log.Println(val)
		valentines = append(valentines, val)
	}
	rows, err := f.db.Query("SELECT senderID FROM valentines WHERE receiverID = (?)", senderID)
	count := f.dbCount(rows)
	jData, err := json.Marshal(ValentineInfo{
		User: User{
			ID:       dataParsed.User.ID,
			Name:     dataParsed.User.FirstName + " " + dataParsed.User.LastName,
			Username: dataParsed.User.Username,
		},
		CountReceived:    count,
		CountSentBeMine:  countBM,
		CountSentDefault: countDef,
		Valentines:       valentines,
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(503)
		return
	}
	w.WriteHeader(200)
	w.Write(jData)
	log.Println(string(jData))
}

func (f urlFunctions) getMyValentine(w http.ResponseWriter, r *http.Request) {
	log.Println("getMyValentine")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	ok, dataParsed := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	today := time.Now()
	valDay := time.Date(2024, 02, 14, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*3600))
	if today.Before(valDay) {
		w.WriteHeader(http.StatusTooEarly)
		log.Println("too early")
		return
	}
	senderID := dataParsed.User.ID
	valentines := make([]Valentine, 0)
	rows, _ := f.db.Query("SELECT ID, senderID, text, type FROM valentines WHERE receiverID = (?)", senderID)
	countDef := 0
	countBM := 0
	for rows.Next() {
		countDef += 1
		val := Valentine{}
		err := rows.Scan(&val.ID, &val.Receiver.ID, &val.Text, &val.Type)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err, " 1")
			return
		}
		if val.Type == "be mine" {
			countBM += 1
			err = f.db.QueryRow("SELECT receiverID FROM valentines WHERE senderID = (?) and type = (?) and receiverID = (?)", senderID, val.Type, val.Receiver.ID).Scan(&val.Receiver.ID)

			if err != nil {
				val.Receiver = User{
					ID:       0,
					Username: "",
					Name:     "",
				}
			}

		} else {
			countDef += 1
		}
		err = f.db.QueryRow("SELECT name, username FROM users WHERE userID = (?)", val.Receiver.ID).Scan(&val.Receiver.Name, &val.Receiver.Username)

		valentines = append(valentines, val)
	}
	rows, err := f.db.Query("SELECT senderID FROM valentines WHERE receiverID = (?)", senderID)
	count := f.dbCount(rows)
	jData, err := json.Marshal(ValentineInfo{
		User: User{
			ID:       dataParsed.User.ID,
			Name:     dataParsed.User.FirstName + " " + dataParsed.User.LastName,
			Username: dataParsed.User.Username,
		},
		CountReceived:    count,
		CountSentBeMine:  countBM,
		CountSentDefault: countDef,
		Valentines:       valentines,
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err, " 3")

		return
	}
	w.Write(jData)
	log.Println(string(jData))
}

func (f urlFunctions) deleteValentine(w http.ResponseWriter, r *http.Request) {
	log.Println("deleteValentine")
	headers := r.Header
	log.Println(headers)
	initData := headers.Get("X-Tg-Token")
	ok, dataParsed := f.validateAndParse(initData, w)
	if !ok {
		return
	}
	senderID := dataParsed.User.ID
	ID, err := strconv.ParseInt(headers.Get("valID"), 10, 64)
	log.Println(senderID)
	log.Println(ID)

	if err != nil || !f.findUser(senderID) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	err = f.db.QueryRow("SELECT senderID, ID FROM valentines WHERE senderID = (?) AND ID = (?)", senderID, ID).Scan(&senderID, &ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("No such valentine")
		return
	}

	_, err = f.db.Exec("DELETE FROM valentines WHERE senderID=(?) AND ID = (?)", senderID, ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error while deleting")
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
