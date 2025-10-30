package monitoring

import (
	"database/sql"
	"time"

	"carwash_backend/internal/logger"
)

// ConnectionMonitor мониторит состояние пула соединений
type ConnectionMonitor struct {
	db *sql.DB
}

// NewConnectionMonitor создает новый монитор соединений
func NewConnectionMonitor(db *sql.DB) *ConnectionMonitor {
	return &ConnectionMonitor{db: db}
}

// StartMonitoring запускает мониторинг соединений
func (cm *ConnectionMonitor) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			cm.logConnectionStats()
		}
	}()
}

// logConnectionStats логирует статистику соединений
func (cm *ConnectionMonitor) logConnectionStats() {
	stats := cm.db.Stats()
	
	logger.Info("Database connection pool stats", map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	})
}
