package main

import (
	"fmt"

	"dalista/dalista-be/api/server"

	_ "dalista/dalista-be/docs" // swagger docs (generate with: swag init)
)

// @title           Dalista API
// @version         1.0
// @description     RESTful API for the Dalista shopping-list / shop-catalog sharing application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Dalista API Support

// @license.name  Proprietary

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer " followed by your access token.

func main() {
	fmt.Println("Starting Dalista API!")
	if err := server.Initialize(); err != nil {
		fmt.Printf("Failed to initialize server: %v\n", err)
		return
	}

	// Run the server
	server.Run()
}
