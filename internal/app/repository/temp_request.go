package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"

	"github.com/sirupsen/logrus"
)

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

func (r *Repository) DeletePlanetFromSystem(system_id, planet_id uint) error {
    if err := r.db.
        Where("planet_system_id = ? AND planet_id = ?", system_id, planet_id).
        Delete(&ds.Temperature_request{}).Error; err != nil {
        return err
    }

    return nil
}

func (r *Repository) UpdatePlanetDistance(system_id, planet_id uint, newDistance uint) error {
	var req ds.Temperature_request
	if err := r.db.
		Where("planet_system_id = ? AND planet_id = ?", system_id, planet_id).
		First(&req).Error; err != nil {
		return err
	}

	if err := r.db.Model(&req).
		Update("planet_distance", newDistance).Error; err != nil {
		return err
	}

	return nil
}

