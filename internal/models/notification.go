package models

type Notification struct {
	UserID    string  `json:"user_id"`
	Channel   string  `json:"channel"`
	To        string  `json:"to"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
}