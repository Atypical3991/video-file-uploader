package main

import (
	"city_os/cmd/app/configs"
	logger "city_os/src/common"
	"city_os/src/controllers"
	"city_os/src/dbconnectors"
	"city_os/src/handlers"
	"city_os/src/middlewares"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

func main() {

	// Loading application configs
	configs.LoadConfig()

	// Init global logger instance with set of common rules
	logger.InitLogger()

	logger.Logger.Info("App  config loaded successfully!!")

	logger.Logger.Info("Logger initiated!!")

	// MongoDBClient contains  db connection object and connection settings
	mongoClient := dbconnectors.MongoDBClient{}
	mongoClient.InitConnection(
		&dbconnectors.MongoDBSettings{
			URI:                      configs.Config.DB.URI,
			PoolSize:                 configs.Config.DB.PoolSize,
			VideoCatalogueDB:         configs.Config.DB.DBs.VideoCatalogueDB,
			VideoFilesCollection:     configs.Config.DB.Collections.VideoFilesColl,
			VideoCatalogueCollection: configs.Config.DB.Collections.VideoCatalogueColl,
		})

	logger.Logger.Info("Mongo Client connected....")

	// Defer function to handle panics and closing of open connections
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error(fmt.Sprintf("Panic occurred!!Error: %v", err))
			// Closing all open connections to MongoDB, as the server is shutting down.
			if conn := mongoClient.GetConnection(); conn != nil {
				if err = conn.(*mongo.Client).Disconnect(context.Background()); err != nil {
					logger.Logger.Error(fmt.Sprintf("Error occurred while closing MongoDB connections!! Error: %v", err))
				}
			}
		}
	}()

	// VideoCatalogueDBWrapper, an abstraction Catalogue DB level methods/function,
	// so that we can replace DB in future with ease and minimum code changes if needed.
	videoCatalogueDBWrapper := dbconnectors.VideoCatalogueDBWrapper{}

	// Initialising Mongo DB level connection object for Catalogue DB
	videoCatalogueDBWrapper.InitDatabase(&mongoClient)

	// VideoFilesDBWrapper, an abstraction Files DB level methods/function,
	// so that we can replace DB in future with ease and minimum code changes if needed.
	videoFilesDBWrapper := dbconnectors.VideoFilesDBWrapper{}

	// Initialising Mongo DB level connection object for Catalogue DB for files.
	videoFilesDBWrapper.InitDatabase(&mongoClient)

	// VideoCatalogueManager, an abstraction for Video Catalogue related Methods/Functions,
	// which will also contains business logics.
	videoCatalogueManagerObj := controllers.VideoCatalogueManager{
		&videoCatalogueDBWrapper,
		&videoFilesDBWrapper,
	}

	// Handler, router handler object, which contains all the common Object instances required to server
	// response for a given request, such as db connections, app config etc
	handler := handlers.Handler{VideoCatalogueManager: &videoCatalogueManagerObj}

	logger.Logger.Info("Router Handler initiated....")

	router := gin.Default()

	// applying CORS rules here
	router.Use(middlewares.CORSMiddleware())

	v1 := router.Group("/v1")
	{
		v1.GET("/health", handler.HealthCheck)
		v1.GET("/files/:fileid", handler.GetFileByIdHandler)
		v1.GET("/files/locate/:fileid", handler.LocateFileByIdHandler)
		v1.DELETE("/files/:fileid", handler.DeleteFileByIdHandler)
		v1.POST("/files", handler.PostSingleFileHandler)
		v1.GET("/files", handler.GetFilesListHandler)
	}

	logger.Logger.Info("Server Starting up.....")
	if err := router.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		logger.Logger.Error(fmt.Printf("Server Startup failed.... Error:%s", err.Error()))
	} else {
		logger.Logger.Info("Server Started successfully.....")
	}
}
