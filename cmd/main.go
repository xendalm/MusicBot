package main

import (
	"bobot/internal/app/bot"
	"bobot/internal/app/db"
	"context"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error while reading env file: %v", err)
	}

	database, err := db.NewConnection(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
		return
	}
	defer database.GetPool(ctx).Close()

	db.MigrationUp()

	b := bot.NewBot(database)
	b.Run()
}
