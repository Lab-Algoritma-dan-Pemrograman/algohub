package database

import (
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
)

const connectTimeout = "15"

type DBs struct {
	Praktikum *gorm.DB
	Siakad    *gorm.DB
	Elearning *gorm.DB
}

func Open(cfg config.DBConfig, debug bool) (*DBs, error) {
	return openUnique(cfg, func(dsn string) (*gorm.DB, error) {
		return connect(dsn, debug)
	})
}

// DSN yang sama dipakai bersama satu pool supaya jumlah koneksi ke pooler tidak berlipat.
func openUnique(cfg config.DBConfig, open func(dsn string) (*gorm.DB, error)) (*DBs, error) {
	pools := map[string]*gorm.DB{}
	get := func(dsn string) (*gorm.DB, error) {
		if dsn == "" {
			return nil, nil
		}
		if db, ok := pools[dsn]; ok {
			return db, nil
		}
		db, err := open(dsn)
		if err != nil {
			return nil, err
		}
		pools[dsn] = db
		return db, nil
	}

	var (
		dbs DBs
		err error
	)
	if dbs.Praktikum, err = get(cfg.Praktikum); err != nil {
		return nil, err
	}
	if dbs.Siakad, err = get(cfg.Siakad); err != nil {
		return nil, err
	}
	if dbs.Elearning, err = get(cfg.Elearning); err != nil {
		return nil, err
	}
	return &dbs, nil
}

func connect(dsn string, debug bool) (*gorm.DB, error) {
	logLevel := logger.Warn
	if debug {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: withConnectTimeout(dsn),
		// PreferSimpleProtocol mematikan prepared statement — WAJIB untuk
		// Supabase Transaction Pooler (pgbouncer mode transaction) yang dipakai
		// di serverless. Tanpa ini, query akan error "prepared statement exists".
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// Pool kecil: di serverless koneksi mudah menumpuk lewat pooler.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(time.Minute)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

// connect_timeout supaya cold-start serverless gagal cepat (bukan menggantung).
func withConnectTimeout(dsn string) string {
	if strings.Contains(dsn, "connect_timeout") {
		return dsn
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return dsn
		}
		q := u.Query()
		q.Set("connect_timeout", connectTimeout)
		u.RawQuery = q.Encode()
		return u.String()
	}
	return dsn + " connect_timeout=" + connectTimeout
}
