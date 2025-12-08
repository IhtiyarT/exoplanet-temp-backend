package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
	"LABS-BMSTU-BACKEND/internal/app/dto"
)

func (r *Repository) AddPlanetToSystem(planetID, systemID uint) error {
	temp_request := &ds.Temperature_request{
		PlanetID:       planetID,
		PlanetSystemID: systemID,
	}

	return r.db.Create(temp_request).Error
}

func (r *Repository) DeletePlanetSystem(id uint) error {
	result := r.db.Model(&ds.Planet_system{}).
		Where("planet_system_id = ?", id).
		Update("status", "Удалена")
	return result.Error
}

func (r *Repository) GetPlanetSystemByID(systemID uint) (ds.Planet_system, error) {
	var system ds.Planet_system
	err := r.db.First(&system, systemID).Error
	if err != nil {
		return ds.Planet_system{}, err
	}
	return system, nil
}

func (r *Repository) GetDraftPlanetSystemID(userID uint) (uint, error) {
	var system_id uint

	err := r.db.
		Model(&ds.Planet_system{}).
		Select("planet_system_id").
		Where("status = ? AND user_id = ?", "Черновик", userID).
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

func (r *Repository) GetPlanetSystemsByUserID(userID uint, system_status string, startDate time.Time, endDate time.Time) ([]ds.Planet_system, error) {
	var planetSystems []ds.Planet_system

	query := r.db.Model(&ds.Planet_system{}).
		Preload("User").
		Preload("Moder").
		Where("user_id = ?", userID).
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

func (r *Repository) UpdatePlanetSystem(system_id uint, input interface{}) error {
	var system ds.Planet_system

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	if err := r.db.Model(&system).Updates(input).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetPlanetSystemFormed(system_id, user_id uint) error {
	var system ds.Planet_system

	if err := r.db.First(&system, system_id).Error; err != nil {
		return err
	}

	if system.UserID != user_id {
		return fmt.Errorf("нельзя формировать чужую заявку")
	}

	if system.Status != "Черновик" || system.DateCreated.IsZero() {
		return fmt.Errorf("статус должен быть 'Черновик' и дата создания указана")
	}

	now := time.Now()
	if err := r.db.Model(&system).Updates(map[string]interface{}{
		"date_formed": now,
		"status":      "Сформирована",
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetPlanetSystemModerStatus(system_id, moder_id uint, status string) error {
    var system ds.Planet_system

    if err := r.db.Preload("Planets").First(&system, system_id).Error; err != nil {
        return err
    }

    if system.Status != "Сформирована" {
        return fmt.Errorf("для завершения/отклонения статус заявки должен быть 'Сформирована'")
    }

    if status == "Завершена" {
        if system.StarLuminosity == 0 {
            if err := r.db.Model(&system).Updates(map[string]interface{}{
                "moder_id":   moder_id,
                "date_ended": time.Now().UTC(),
                "status":     "Отклонена",
            }).Error; err != nil {
                return err
            }
            return nil
        }

        var requests []ds.Temperature_request
        if err := r.db.Where("planet_system_id = ?", system_id).Find(&requests).Error; err != nil {
            return err
        }

        for _, req := range requests {
            if req.PlanetDistance == 0 {
                if err := r.db.Model(&system).Updates(map[string]interface{}{
                    "moder_id":   moder_id,
                    "date_ended": time.Now().UTC(),
                    "status":     "Отклонена",
                }).Error; err != nil {
                    return err
                }
                return nil
            }
        }

        type PlanetData struct {
            PlanetID       uint    `json:"planet_id"`
            Albedo         float64 `json:"albedo"`
            PlanetDistance uint    `json:"planet_distance"`
        }

        planetsData := make([]PlanetData, 0)
        for _, req := range requests {
            for _, planet := range system.Planets {
                if planet.PlanetID == req.PlanetID {
                    planetsData = append(planetsData, PlanetData{
                        PlanetID:       req.PlanetID,
                        Albedo:         planet.Albedo,
                        PlanetDistance: req.PlanetDistance,
                    })
                    break
                }
            }
        }

        payload := map[string]interface{}{
            "system_id":      system_id,
            "planets":        planetsData,
            "star_luminosity": system.StarLuminosity,
        }

        payloadBytes, err := json.Marshal(payload)
        if err != nil {
            return err
        }

        asyncURL := "http://localhost:8000/api/calculate/"
        resp, err := http.Post(asyncURL, "application/json", bytes.NewBuffer(payloadBytes))
        if err != nil {
            return fmt.Errorf("ошибка при вызове асинхронного сервиса: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("асинхронный сервис вернул %d", resp.StatusCode)
        }
    }

    updates := map[string]interface{}{
        "moder_id":   moder_id,
        "date_ended": time.Now().UTC(),
        "status":     status,
    }

    if err := r.db.Model(&system).Updates(updates).Error; err != nil {
        return err
    }

    if err := r.db.First(&system, system_id).Error; err != nil {
        return err
    }

    return nil
}

func (r *Repository) GetPlanetSystemsForList(statusFilter string, startDate, endDate time.Time) ([]dto.SystemListItem, error) {
	return r.getPlanetSystemsForListQuery(statusFilter, startDate, endDate, 0)
}

func (r *Repository) GetPlanetSystemsForListByUser(userID uint, statusFilter string, startDate, endDate time.Time) ([]dto.SystemListItem, error) {
	return r.getPlanetSystemsForListQuery(statusFilter, startDate, endDate, userID)
}

func (r *Repository) getPlanetSystemsForListQuery(statusFilter string, startDate, endDate time.Time, userID uint) ([]dto.SystemListItem, error) {
	var results []struct {
		ID             uint
		DateCreated    time.Time
		Status         string
		UserLogin      string
		ModerID        *uint
		StarName       string
		StarType       string
		StarLuminosity float64
		PlanetCount    int64
	}

	query := r.db.Model(&ds.Planet_system{}).
		Select(`
			planet_systems.planet_system_id AS id,
			planet_systems.date_created,
			planet_systems.status,
			users.login AS user_login,
			planet_systems.moder_id,
			planet_systems.star_name,
			planet_systems.star_type,
			planet_systems.star_luminosity,
			COUNT(temperature_requests.planet_id) AS planet_count
		`).
		Joins("LEFT JOIN users ON users.user_id = planet_systems.user_id").
		Joins("LEFT JOIN temperature_requests ON temperature_requests.planet_system_id = planet_systems.planet_system_id").
		Where("planet_systems.status NOT IN ?", []string{"Черновик", "Удалена"}).
		Where("planet_systems.date_formed BETWEEN ? AND ?", startDate, endDate)

	if statusFilter != "" {
		query = query.Where("planet_systems.status = ?", statusFilter)
	}
	if userID > 0 {
		query = query.Where("planet_systems.user_id = ?", userID)
	}

	query = query.Group("planet_systems.planet_system_id, users.login")

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	items := make([]dto.SystemListItem, len(results))
	for i, res := range results {
		planetTempCount := 0
		if res.Status == "Завершена" {
			planetTempCount = int(res.PlanetCount)
		}
		if res.Status == "Завершена" {
			planetTempCount = int(res.PlanetCount)
		}

		items[i] = dto.SystemListItem{
			ID:              res.ID,
			DateCreated:     res.DateCreated.Format(time.RFC3339),
			Status:          res.Status,
			UserLogin:       res.UserLogin,
			ModerID:         res.ModerID,
			StarName:        res.StarName,
			StarType:        res.StarType,
			StarLuminosity:  res.StarLuminosity,
			PlanetCount:     int(res.PlanetCount),
			PlanetTempCount: planetTempCount,
		}
	}

	return items, nil
}

func (r *Repository) UpdateTemperatureRequest(systemID, planetID, temp uint) error {
	return r.db.Model(&ds.Temperature_request{}).
		Where("planet_system_id = ? AND planet_id = ?", systemID, planetID).
		Update("planet_temperature", temp).Error
}
