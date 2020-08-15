package main

import (
	"context"
	"flag"
	"log"

	"github.com/nicholas-p1/buna"
	"go.uber.org/zap"
)

func main() {
	var bunaDBFilePath = flag.String("db", "bunaDB.db", "SQLite BunaDB file path")
	flag.Parse()

	ctx := context.Background()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("buna: failed to create zap logger: %v\n", err)
	}
	defer logger.Sync() // nolint:errcheck

	bunaDB, err := buna.OpenSQLiteDB(ctx, logger, *bunaDBFilePath)
	if err != nil {
		logger.Fatal("buna: failed to open SQLite buna database")
	}
	defer bunaDB.Close()
	logger.Info("buna: connected to SQLite buna database")
}
