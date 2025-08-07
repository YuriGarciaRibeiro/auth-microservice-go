package model

type User struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	Password string
	Verified bool
}
