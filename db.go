package yiigo

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	entsql "github.com/facebook/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "postgres"
	SQLite   DBDriver = "sqlite3"
)

var (
	defaultDB *sqlx.DB
	dbmap     sync.Map

	defaultEntDriver *entsql.Driver
	entmap           sync.Map
)

type dbConfig struct {
	Driver          string `toml:"driver"`
	DSN             string `toml:"dsn"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxIdleTime int    `toml:"conn_max_idle_time"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

func dbDial(cfg *dbConfig) (*sql.DB, error) {
	if !InStrings(cfg.Driver, string(MySQL), string(Postgres), string(SQLite)) {
		return nil, fmt.Errorf("yiigo: unknown db driver %s, expects mysql, postgres, sqlite3", cfg.Driver)
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()

		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

func initDB() {
	tree, ok := env.get("db").(*toml.Tree)

	if !ok {
		return
	}

	keys := tree.Keys()

	if len(keys) == 0 {
		return
	}

	for _, name := range keys {
		node, ok := tree.Get(name).(*toml.Tree)

		if !ok {
			continue
		}

		cfg := new(dbConfig)

		if err := node.Unmarshal(cfg); err != nil {
			logger.Panic("yiigo: db init error", zap.String("name", name), zap.Error(err))
		}

		db, err := dbDial(cfg)

		if err != nil {
			logger.Panic("yiigo: db init error", zap.String("name", name), zap.Error(err))
		}

		sqlxDB := sqlx.NewDb(db, cfg.Driver)
		entDriver := entsql.OpenDB(cfg.Driver, db)

		if name == AsDefault {
			defaultDB = sqlxDB
			defaultEntDriver = entDriver
		}

		dbmap.Store(name, sqlxDB)
		entmap.Store(name, entDriver)

		logger.Info(fmt.Sprintf("yiigo: db.%s is OK.", name))
	}
}

// DB returns a db.
func DB(name ...string) *sqlx.DB {
	if len(name) == 0 {
		if defaultDB == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", AsDefault))
		}

		return defaultDB
	}

	v, ok := dbmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*sqlx.DB)
}

// EntDriver returns an ent dialect.Driver.
func EntDriver(name ...string) *entsql.Driver {
	if len(name) == 0 || name[0] == AsDefault {
		if defaultEntDriver == nil {
			logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", AsDefault))
		}

		return defaultEntDriver
	}

	v, ok := entmap.Load(name[0])

	if !ok {
		logger.Panic(fmt.Sprintf("yiigo: unknown db.%s (forgotten configure?)", name[0]))
	}

	return v.(*entsql.Driver)
}
