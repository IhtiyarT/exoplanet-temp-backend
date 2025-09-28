package main

import (
	"fmt"

	"LABS-BMSTU-BACKEND/internal/app/config"
	"LABS-BMSTU-BACKEND/internal/app/dsn"
	"LABS-BMSTU-BACKEND/internal/app/handler"
	"LABS-BMSTU-BACKEND/internal/app/myminio"
	"LABS-BMSTU-BACKEND/internal/app/repository"
	"LABS-BMSTU-BACKEND/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)	

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