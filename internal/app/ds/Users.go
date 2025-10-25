package ds

import (
	"LABS-BMSTU-BACKEND/internal/app/role"
)

type Users struct {
	UserID   uint      `json:"user_id" gorm:"primaryKey;autoIncrement"`
	Login    string    `json:"login" gorm:"type:varchar(100);unique"`
	Password string    `json:"password" gorm:"type:varchar(100)"`
	Role     role.Role `json:"role" sql:"type:string;"`

	PlanetSystems []Planet_system `json:"planet_systems" gorm:"foreignKey:UserID"`
}	
