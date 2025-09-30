package repository

import (
	// "fmt"
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) AddPlanetToSystem(planetID, systemID uint) error {
	temp_request := &ds.Temperature_request{
		PlanetID:       planetID,
		PlanetSystemID: systemID,
	}
	
	return r.db.Create(temp_request).Error
}

func (r *Repository) DeletePlanetSystem(id uint) {
	query := "UPDATE planet_systems SET status = 'Удалена' WHERE planet_system_id = $1"
	r.db.Exec(query, id)
}

func (r *Repository) GetPlanetSystemByID(systemID uint) (ds.Planet_system, error) {
    var system ds.Planet_system
    err := r.db.First(&system, systemID).Error
    if err != nil {
        return ds.Planet_system{}, err
    }
    return system, nil
}

type tempSystemID struct {
    PlanetSystemID uint
}

func (r *Repository) GetDraftPlanetSystemID() (uint, error) {
    var system tempSystemID

    err := r.db.
        Model(&ds.Planet_system{}).
        Select("planet_system_id").
        Where("status = ?", "Черновик").
        First(&system).Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return 0, nil
    }
    if err != nil {
        return 0, err
    }

    return system.PlanetSystemID, nil
}

func (r *Repository) CreateNewDraftPlanetSystem(userID uint) (uint, error) {
    new_system := ds.Planet_system{
        DateCreated:      time.Now(),
        Status:           "Черновик",
        UserID:           userID,
        StarName:         "Класс А (Белые звезды)",
        StarLuminosity:  1.1,
    }
    
    result := r.db.Create(&new_system)
    if result.Error != nil {
        return 0, result.Error
    }
    
    return new_system.PlanetSystemID, nil
}
