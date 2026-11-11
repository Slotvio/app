package service

import "time"

type AuthData struct {
	UserID int64
	Token  string
	ExpAt  time.Time
}
