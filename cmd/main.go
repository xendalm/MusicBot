package main

import (
	"bobot/internal/pkg/bot"
	"bobot/internal/pkg/db"
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

	db, err := db.NewConnection(ctx)
	if err != nil {
		log.Fatalf("unable to connect to the database: %v", err)
		return
	}

	bot := bot.NewBot(db)

	bot.Run()
}
