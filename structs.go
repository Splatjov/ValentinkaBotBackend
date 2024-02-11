package main

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Valentine struct {
	Receiver User   `json:"receiver"`
	Text     string `json:"text"`
	Type     string `json:"type"`
	ID       int    `json:"id"`
}

type ValentineInfo struct {
	User             User        `json:"user"`
	CountReceived    int         `json:"countReceived"`
	CountSentDefault int         `json:"countSentDefault"`
	CountSentBeMine  int         `json:"countSentBeMine"`
	Valentines       []Valentine `json:"valentines"`
}
