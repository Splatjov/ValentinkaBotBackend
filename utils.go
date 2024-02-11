package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"log"
	"net/http"
	"time"
)

func (f urlFunctions) findUser(userid int64) bool {
	err := f.db.QueryRow("SELECT userid FROM users WHERE userID = ?", userid).Scan(&userid)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return true
}

func (f urlFunctions) dbCount(rows *sql.Rows) int {
	count := 0
	for rows.Next() {
		count += 1
	}
	return count
}

func (f urlFunctions) updateUser(chat telego.User) {
	log.Println("updateUser")
	if !f.findUser(chat.ID) {
		_, err := f.db.Exec("INSERT INTO users (userID, username, name) VALUES (?, ?, ?)", chat.ID, chat.Username, chat.FirstName+" "+chat.LastName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Username added successfully.")
	}
}

func (f urlFunctions) validateAndParse(initData string, w http.ResponseWriter) (bool, initdata.InitData) {
	expIn := 24 * time.Hour
	err := initdata.Validate(initData, f.botToken, expIn)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		fmt.Println(err)
		return false, initdata.InitData{}
	}
	data, err := initdata.Parse(initData)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Println(err)
		return false, initdata.InitData{}
	}
	f.updateUser(telego.User{
		ID:        data.User.ID,
		FirstName: data.User.FirstName,
		LastName:  data.User.LastName,
		Username:  data.User.Username,
	})
	return true, data

}
