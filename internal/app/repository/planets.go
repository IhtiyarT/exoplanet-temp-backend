package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
)

func (r *Repository) GetPlanets() ([]ds.Planets, error) {
	var planets []ds.Planets
	err := r.db.Where("is_delete = false").Find(&planets).Error
	if err != nil {
		return nil, err
	}
	return planets, nil
}

func (r *Repository) GetPlanet(id uint) (ds.Planets, error) {
	planet := ds.Planets{}
	err := r.db.Where("planet_id = ?", id).First(&planet).Error
	if err != nil {
		return ds.Planets{}, err
	}
	return planet, nil
}

func (r *Repository) GetPlanetsByTitle(title string) ([]ds.Planets, error) {
	var planets []ds.Planets
	err := r.db.
		Where("planet_title ILIKE ? AND is_delete = false", "%"+title+"%").
		Find(&planets).Error
	if err != nil {
		return nil, err
	}
	return planets, nil
}

func (r *Repository) GetCountBySystemID(systemID uint) (int64, error) {
	var count int64
	err := r.db.
		Model(&ds.Temperature_request{}).
		Where("planet_system_id = ?", systemID).
		Count(&count).Error

	return count, err
}

func (r *Repository) CreatePlanet(planet ds.Planets) (int, error) {
	if err := r.db.Create(&planet).Error; err != nil {
		return -1, err;
	}
	
	return int(planet.PlanetID), nil
}

func (r *Repository) UpdatePlanet(id uint, input interface{}) (error) {
	var planet ds.Planets

	if err := r.db.First(&planet, id).Error; err != nil {
		return err
	}

	if err := r.db.Model(&planet).Updates(input).Error; err != nil {
        return err
    }

	return nil
}

func (r *Repository) DeletePlanet(planet_id uint) error {
	if err := r.db.Model(&ds.Planets{}).
		Where("planet_id = ?", planet_id).
		Update("planet_image", "").Error; err != nil {
		return err
	}

	if err := r.db.Model(&ds.Planets{}).
		Where("planet_id = ?", planet_id).
		Update("is_delete", true).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdatePlanetImage(planet_id string, new_image string) error {
	planet := ds.Planets{}
	if result := r.db.First(&planet, planet_id); result.Error != nil {
		return result.Error
	}
	planet.PlanetImage = new_image
	result := r.db.Save(planet)
	return result.Error
}
