package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bonggalshn/budget-be/internal/config"
)

func main(){
	cfg := config.Load()
	
	fmt.Printf("Budget API Server\n")
	fmt.Printf("================\n")
	fmt.Printf("Server: %s:%s\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s:%s\n", cfg.DB.Host, cfg.DB.Port)
	fmt.Printf("JWT Expiry: %v\n", cfg.JWT.Expiry)
	
	// TODO: Implement HTTP server
	log.Println("Server not yet implemented - run /speckit.implement to continue")
	os.Exit(0)
}