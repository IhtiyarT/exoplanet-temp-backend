package main

import (
	"fmt"

	"LABS-BMSTU-BACKEND/internal/app/config"
	"LABS-BMSTU-BACKEND/internal/app/dsn"
	"LABS-BMSTU-BACKEND/internal/app/handler"
	"LABS-BMSTU-BACKEND/internal/app/myminio"
	"LABS-BMSTU-BACKEND/internal/app/repository"
	"LABS-BMSTU-BACKEND/internal/pkg"

	_ "LABS-BMSTU-BACKEND/docs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)	

// @title BMSTU LAB
// @version 1.0
// @description BMSTU dia lab

// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru

// @license.name AS IS (NO WARRANTY)

// @host localhost:8082
// @schemes http
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите: "Bearer {token}" (без кавычек)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	logger := logrus.New()
	minioClient := myminio.NewMinioClient(logger)

	hand := handler.NewHandler(rep, minioClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}