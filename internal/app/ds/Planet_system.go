package ds

import (
	"time"
)

type Planet_system struct {
	PlanetSystemID uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	DateCreated    time.Time `json:"date_created" gorm:"not null"`
	Status         string    `json:"status" gorm:"type:varchar(50);not null"`
	UserID         uint      `json:"user_id" gorm:"foreignKey:UserID;not null"`

	ModerID         uint      `json:"moder_id"`
	DateFormed      time.Time `json:"date_formed"`
	DateEnded       time.Time `json:"date_ended"`
	StarType        string    `json:"star_type" gorm:"type:varchar(255)"`
	StarName        string    `json:"star_name" gorm:"type:varchar(255)"`
	StarTemperature uint      `json:"star_temperature"`
}
