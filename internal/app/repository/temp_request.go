package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"

	"github.com/sirupsen/logrus"
)

type PlanetWithSystemData struct {
	PlanetID          uint    `json:"planet_id"`
	PlanetTitle       string  `json:"planet_title"`
	PlanetImage       string  `json:"planet_image"`
	PlanetDescription string  `json:"planet_description"`
	Albedo            float64 `json:"albedo"`
	PlanetSystemID    uint    `json:"planet_system_id"`
	PlanetTemperature uint    `json:"planet_temperature"`
	PlanetDistance    uint    `json:"planet_distance"`
	StarName          string  `json:"star_name"`
	StarLuminoсity    float64 `json:"star_luminoсity"`
	StarType          string  `json:"star_type"`
}

func (r *Repository) GetPlanetsWithSystemData(system_id uint) ([]PlanetWithSystemData, error) {
	var results []PlanetWithSystemData

	err := r.db.
		Table("planets p").
		Select("p.*, tr.planet_system_id, tr.planet_temperature, tr.planet_distance, ps.star_name, ps.star_luminosity, ps.star_type").
		Joins("INNER JOIN temperature_requests tr ON p.planet_id = tr.planet_id").
		Joins("INNER JOIN planet_systems ps ON tr.planet_system_id = ps.planet_system_id").
		Where("p.is_delete = ? AND tr.planet_system_id = ?", false, system_id).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) CheckIsDelete(systemID uint) bool {
	var count int64

	err := r.db.
		Model(&ds.Planet_system{}).
		Where("planet_system_id = ? AND status = ?", systemID, "Черновик").
		Count(&count).Error

	if err != nil {
		logrus.Errorf("Ошибка проверки системы %d: %v", systemID, err)
		return false
	}

	return count > 0
}
