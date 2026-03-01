package main

import (
	"os"
	"time"

	"blog_backend/config"
	"blog_backend/routes"

	"log/slog"

	"gorm.io/gorm"
)

func main() {
	// Logger initialisieren
	logger := slog.New(slog.NewTextHandler(
		// Writer
		os.Stdout,
		// Optionen (nil = Standard)
		nil,
	))

	slog.SetDefault(logger)

	slog.Info("Starte Backend...")

	// DB initialisieren mit Retry
	var db *gorm.DB // passe den Typ an, z.B. *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = config.ConnectDB()
		if err == nil {
			break
		}
		slog.Warn("DB noch nicht ready, warte 2s...", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		slog.Error("DB-Verbindung fehlgeschlagen", "error", err)
		os.Exit(1)
	}

	// Router starten
	r := routes.SetupRouter(db)

	slog.Info("Backend läuft", "port", 8080)
	if err := r.Run(":8080"); err != nil {
		slog.Error("Server konnte nicht gestartet werden", "error", err)
		os.Exit(1)
	}
}
