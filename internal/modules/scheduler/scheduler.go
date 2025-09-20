package scheduler

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sonyarianto/gobete/internal/systems/db"
)

func StartCleanupUserSessionScheduler() {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		db.DB.Exec("DELETE FROM user_sessions WHERE expires_at < ?", time.Now())
	})
	c.Start()
}
