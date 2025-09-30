package utils

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/rs/xid"
)

// Генератор имени для фото
func GenerateUniqueName(imageName *string) error {
	parts := strings.Split(*imageName, ".")
	if len(parts) > 1 {
		fileExt := parts[len(parts)-1]
		uniqueID := xid.New()
		*imageName = fmt.Sprintf("%s.%s", uniqueID.String(), fileExt)
		return nil
	}
	return fmt.Errorf("uncorrect file name. not fount image extension")
}

func ParseDateString(dateString string) (time.Time, error) {
	format := "2006-01-02" // Пример: "2023-11-30"

	// Парсинг строки даты
	parsedTime, err := time.Parse(format, dateString)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// Contains Функция для проверки наличия элемента в срезе
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// достает имя из URL
func ExtractObjectName(url string) string {
	parts := strings.Split(url, "/")
	return path.Base(parts[len(parts)-1])
}
