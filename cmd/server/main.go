package main

import (
	"log"
	"os"

	"Exlord/otpservice/internal/server"
)

func main() {
	port := getenv("PORT", "8080")
	jwtSecret := getenv("JWT_SECRET", "devsecret")
	log.Printf("Starting server on :%s", port)

	if err := server.Start(":"+port, jwtSecret); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
