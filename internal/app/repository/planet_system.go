package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"errors"
	"fmt"
	"math"
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

func (r *Repository) GetDraftPlanetSystemID() (uint, error) {
	var system_id uint

	err := r.db.
		Model(&ds.Planet_system{}).
		Select("planet_system_id").
		Where("status = ?", "Черновик").
		First(&system_id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return system_id, nil
}

func (r *Repository) CreateNewDraftPlanetSystem(userID uint) (uint, error) {
	new_system := ds.Planet_system{
		DateCreated:    time.Now(),
		Status:         "Черновик",
		UserID:         userID,
		StarName:       "Солнце",
		StarType:       "Желтый карлик",
		StarLuminosity: 1,
	}

	result := r.db.Create(&new_system)
	if result.Error != nil {
		return 0, result.Error
	}

	return new_system.PlanetSystemID, nil
}

func (r *Repository) GetPlanetSystemsList(system_status string, startDate time.Time, endDate time.Time) ([]ds.Planet_system, error) {
	var planetSystems []ds.Planet_system

	query := r.db.Model(&ds.Planet_system{}).
		Preload("User").
		Preload("Moder").
		Where("status NOT IN ?", []string{"Удалена", "Черновик"})

	if system_status != "" {
		query = query.Where("status = ?", system_status)
	}

	query = query.Where("date_formed BETWEEN ? AND ?", startDate, endDate)

	result := query.Find(&planetSystems)
	return planetSystems, result.Error
}

func (r *Repository) GetPlanetSystemAndPlanetsByID(system_id uint) (*ds.Planet_system, error) {
	var system ds.Planet_system

	result := r.db.
		Preload("User").
		Preload("Moder").
		Preload("Planets", "is_delete = false").
		First(&system, system_id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &system, nil
}

func (r *Repository) UpdatePlanetSystem(system_id uint, input interface{}) (error) {
	var system ds.Planet_system

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	if err := r.db.Model(&system).Updates(input).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetPlanetSystemFormed(system_id uint, user_id uint) (error) {
	var system ds.Planet_system

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	if system.Status != "Черновик" || system.DateCreated.IsZero() {
		return fmt.Errorf("статус должен быть - черновик, дата создания должна быть указа")
	}

	if system.UserID != user_id {
		return fmt.Errorf("только создатель может формировать заявку")
	}

	now := time.Now()
	if err := r.db.Model(&system).Updates(map[string]interface{}{
		"date_formed": now,
		"status":      "Сформирована",
	}).Error; err != nil {
		return err
	}

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetPlanetSystemModerStatus(system_id, moder_id uint, status string) error {
	var system ds.Planet_system

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	if system.Status != "Сформирована" {
		return fmt.Errorf("для завершения/отклонения статус заявки должен быть 'Сформирована'")
	}

	updates := map[string]interface{}{
		"moder_id":   moder_id,
		"date_ended": time.Now().UTC(),
		"status":     status,
	}

	if err := r.db.Model(&system).Updates(updates).Error; err != nil {
		return err
	}

	if status == "Завершена" {
		var requests []ds.Temperature_request
		if err := r.db.Where("planet_system_id = ?", system_id).Find(&requests).Error; err != nil {
			return err
		}

		starLum := float64(system.StarLuminosity)

		for _, req := range requests {
			var planet ds.Planets
			if err := r.db.First(&planet, req.PlanetID).Error; err != nil {
				return err
			}

			albedo := planet.Albedo
			dist := float64(req.PlanetDistance)
			if dist == 0 {
				return fmt.Errorf("дистанция не может быть равной нулю")
			}
			temp := 1.077e5 * math.Sqrt(math.Sqrt(starLum*((1-albedo)/(dist*dist))))

			if err := r.db.Model(&ds.Temperature_request{}).
				Where("planet_id = ? AND planet_system_id = ?", req.PlanetID, system_id).
				Update("planet_temperature", uint(temp)).Error; err != nil {
				return err
			}
		}
	}

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	return nil
}
