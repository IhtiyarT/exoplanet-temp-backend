package repository

import (
	"LABS-BMSTU-BACKEND/internal/app/ds"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func (r *Repository) CreateUser(user ds.Users) error {
	return r.db.Create(&user).Error
}

func (r *Repository) GetUserByLogin(login string) (ds.Users, error) {
	var user ds.Users

	err := r.db.Where("login = ?", login).Find(&user).Error
	if err != nil {
		return ds.Users{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByID(id uint) (ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, id).Error; err != nil {
		return ds.Users{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(id uint, new_login, new_password string) error {
    updates := map[string]interface{}{}
    
    if new_login != "" {
        updates["login"] = new_login
    }
    
    if new_password != "" {
        updates["password"] = new_password
    }
    
    if len(updates) == 0 {
        return nil
    }
    
    return r.db.Model(&ds.Users{}).Where("user_id = ?", id).Updates(updates).Error
}

func (r *Repository) LoginUser(login, password string) (ds.Users, error) {
    var user ds.Users
    err := r.db.Where("login = ? AND password = ?", login, password).First(&user).Error
    if err != nil {
        return ds.Users{}, err
    }
    return user, nil
}

func (r *Repository) SaveJWTToken(id uint, token string) error {
	expiration := 1 * time.Hour

	user_id_str := strconv.FormatUint(uint64(id), 10)

	err := r.rd.Set(user_id_str, token, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) AddToBlacklist(token string, expiration time.Duration) error {
    return r.rd.Set(token, "blacklisted", expiration).Err()
}

func (r *Repository) IsInBlacklist(token string) (bool, error) {
    result, err := r.rd.Get(token).Result()
    if err == redis.Nil {
        return false, nil
    } else if err != nil {
        return false, err
    }
    return result == "blacklisted", nil
}

func (r *Repository) DeleteTokenByUserID(user_id_str string){
	r.rd.Del(user_id_str).Err()
}
