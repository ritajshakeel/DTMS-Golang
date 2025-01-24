package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
