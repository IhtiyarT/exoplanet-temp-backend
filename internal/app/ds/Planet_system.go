package ds

import (
	"time"
)

type Planet_system struct {
	PlanetSystemID uint      `json:"planet_system_id" gorm:"primaryKey;autoIncrement"`
	DateCreated    time.Time `json:"date_created" gorm:"not null"`
	Status         string    `json:"status" gorm:"type:varchar(50);not null"`
	UserID         uint      `json:"user_id" gorm:"foreignKey:UserID;not null"`
	User           Users     `json:"user" gorm:"foreignKey:UserID;references:UserID"`

	ModerID        uint      `json:"moder_id"`
	Moder          Users     `json:"moder" gorm:"foreignKey:ModerID;references:UserID"`
	DateFormed     time.Time `json:"date_formed"`
	DateEnded      time.Time `json:"date_ended"`
	StarType       string    `json:"star_type" gorm:"type:varchar(255)"`
	StarName       string    `json:"star_name" gorm:"type:varchar(255)"`
	StarLuminosity float64   `json:"star_luminosity"`

	Planets []Planets `json:"planets" gorm:"many2many:temperature_requests;joinForeignKey:PlanetSystemID;joinReferences:PlanetID"`
}
