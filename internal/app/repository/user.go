package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
)

func (r *Repository) CreateUser(input map[string]interface{}) error {
	isAdmin := false
	if v, ok := input["is_admin"].(bool); ok {
		isAdmin = v
	}

	user := ds.Users{
		Login:    input["login"].(string),
		Password: input["password"].(string),
		IsAdmin:  isAdmin,
	}

	return r.db.Create(&user).Error
}

func (r *Repository) GetUserByID(id uint) (ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, id).Error; err != nil {
		return ds.Users{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(id uint, input map[string]string) error {
	updates := map[string]interface{}{}
	if login, ok := input["login"]; ok {
		updates["login"] = login
	}
	if password, ok := input["password"]; ok {
		updates["password"] = password
	}

	return r.db.Model(&ds.Users{}).Where("user_id = ?", id).Updates(updates).Error
}

func (r *Repository) AuthenticateUser(login, password string) (ds.Users, error) {
	var user ds.Users
	if err := r.db.Where("login = ? AND password = ?", login, password).First(&user).Error; err != nil {
		return ds.Users{}, err
	}
	return user, nil
}

