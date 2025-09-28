package repository

import (
	"fmt"

	"LABS-BMSTU-BACKEND/internal/app/ds"
)

func (r *Repository) GetPlanets() ([]ds.Planets, error) {
	var planets []ds.Planets
	err := r.db.Find(&planets).Error
	if err != nil {
		return nil, err
	}
	if len(planets) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return planets, nil
}

func (r *Repository) GetPlanet(id int) (ds.Planets, error) {
	planet := ds.Planets{}
	err := r.db.Where("planet_id = ?", id).First(&planet).Error
	if err != nil {
		return ds.Planets{}, err
	}
	return planet, nil
}

func (r *Repository) GetPlanetsByTitle(title string) ([]ds.Planets, error) {
	var planets []ds.Planets
	err := r.db.Where("planet_title ILIKE ?", "%"+title+"%").Find(&planets).Error
	if err != nil {
		return nil, err
	}
	return planets, nil
}

func (r *Repository) GetCountBySystemID(system_id uint) (int64, error) {
    var count int64
    err := r.db.
        Model(&ds.Temperature_request{}).
        Where("planet_system_id = ?", system_id).
        Count(&count).Error
    
    return count, err
}

func (r *Repository) CreatePlanet(planet ds.Planets) (int, error) {
	if err := r.db.Create(&planet).Error; err != nil {
		return -1, err;
	}
	
	return int(planet.PlanetID), nil
}

func (r *Repository) UpdatePlanet(id uint, input interface{}) (ds.Planets, error) {
	var planet ds.Planets

	if err := r.db.First(&planet, id).Error; err != nil {
		return ds.Planets{}, err
	}

	if err := r.db.Model(&planet).Updates(input).Error; err != nil {
        return ds.Planets{}, err
    }

	return planet, nil
}

func (r *Repository) DeletePlanet(planet_id uint) error {
	return r.db.Model(&ds.Planets{}).
		Where("planet_id = ?", planet_id).
		Update("is_delete", true).Error
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
