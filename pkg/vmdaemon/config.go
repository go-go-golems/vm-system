package vmdaemon

import "time"

// Config controls daemon host runtime behavior.
type Config struct {
	DBPath          string
	ListenAddr      string
	ReadTimeout     time.Duration
	ReadHeaderTime  time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func DefaultConfig(dbPath string) Config {
	return Config{
		DBPath:          dbPath,
		ListenAddr:      "127.0.0.1:3210",
		ReadTimeout:     15 * time.Second,
		ReadHeaderTime:  5 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}
