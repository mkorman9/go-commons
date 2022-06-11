package postgres

import (
	"errors"
	"github.com/gookit/config/v2"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DialPostgres() (*gorm.DB, func(), error) {
	dsn := config.String("postgres.dsn")
	verbose := config.Bool("postgres.verbose")
	poolMaxOpen := config.Int("postgres.pool.maxopen")
	poolMaxIdle := config.Int("postgres.pool.maxidle")
	poolMaxLifeTimeValue := config.Int("postgres.pool.maxlifetime")
	poolMaxIdleTimeValue := config.Int("postgres.pool.maxidletime")

	if dsn == "" {
		return nil, func() {}, errors.New("postgres.dsn cannot be empty")
	}

	if poolMaxOpen == 0 {
		poolMaxOpen = 10
	}

	if poolMaxIdle == 0 {
		poolMaxIdle = 5
	}

	poolMaxLifeTime := time.Hour
	if poolMaxLifeTimeValue > 0 {
		poolMaxLifeTime = time.Duration(poolMaxLifeTimeValue) * time.Millisecond
	}

	poolMaxIdleTime := 30 * time.Minute
	if poolMaxIdleTimeValue > 0 {
		poolMaxIdleTime = time.Duration(poolMaxIdleTimeValue) * time.Millisecond
	}

	gormConfig := &gorm.Config{
		Logger:      &gormLogger{verbose: verbose},
		NowFunc:     func() time.Time { return time.Now().UTC() },
		QueryFields: true,
	}

	log.Debug().Msg("Establishing Postgres connection")

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err == nil {
		sqlDB, err := db.DB()
		if err != nil {
			return db, func() {}, err
		}

		sqlDB.SetMaxOpenConns(poolMaxOpen)
		sqlDB.SetMaxIdleConns(poolMaxIdle)
		sqlDB.SetConnMaxLifetime(poolMaxLifeTime)
		sqlDB.SetConnMaxIdleTime(poolMaxIdleTime)

		log.Info().Msg("Successfully connected to Postgres")
	}

	closeDBFunc := func() {
		closeDB(db)
	}

	return db, closeDBFunc, err
}

func closeDB(db *gorm.DB) {
	log.Debug().Msg("Closing Postgres connection")

	sqlDB, err := db.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire *sql.DB reference when closing Postgres connection")
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error when closing Postgres connection")
	} else {
		log.Info().Msg("Postgres connection closed successfully")
	}
}
