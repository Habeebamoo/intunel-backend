package models

type OAuthUser struct {
	Name        string
	Email       string
	Avatar      string
}

type AuthResponse struct {
	Token  string  `json:"token"`
	User   User    `json:"user"`
}