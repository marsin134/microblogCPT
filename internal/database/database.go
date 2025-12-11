package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"microblogCPT/internal/config"
	"os"
	"time"
)

type MethodsDB interface {
	CloseDB() error
	RunMigrations(migrationFilePath string) error
	HealthCheck() error
	GetDB() *DB
}

type DB struct {
	*sqlx.DB
}

func ConnectDB(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.DbHOST,
		cfg.DB.DbPORT,
		cfg.DB.DbUSER,
		cfg.DB.DbPASSWORD,
		cfg.DB.DbNAME,
		cfg.DB.DbSSLMODE,
	)

	fmt.Printf("Подключаемся к БД: host=%s, dbname=%s\n", cfg.DB.DbHOST, cfg.DB.DbNAME)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке подключения к БД: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	dbStruct := DB{db}

	err = MethodsDB.RunMigrations(&dbStruct, "migrations/001_create_tables.sql")
	if err != nil {
		log.Printf("Внимание: ошибка при применении миграций: %v", err)
	}

	err = MethodsDB.HealthCheck(&dbStruct)
	if err != nil {
		log.Fatalf("Проверка БД не пройдена: %v", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL")
	return &dbStruct, nil
}

func (db *DB) CloseDB() error {
	return db.DB.Close()
}

func (db *DB) RunMigrations(migrationFilePath string) error {
	if _, err := os.Stat(migrationFilePath); os.IsNotExist(err) {
		return fmt.Errorf("файл миграций не найден: %s", migrationFilePath)
	}

	migrationSQL, err := os.ReadFile(migrationFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при чтении файла миграций: %w", err)
	}

	fmt.Printf("Применяем миграции из файла: %s\n", migrationFilePath)

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	fmt.Println("Миграции успешно применены")
	return nil
}

func (db *DB) HealthCheck() error {
	if db == nil {
		return fmt.Errorf("подключение к БД не инициализировано")
	}

	return db.Ping()
}

func (db *DB) GetDB() *DB {
	return db
}

// psql -h localhost -U postgres
// psql -h localhost -U postgres -d microblog
// psql -h localhost -U postgres -d microblog -c "\dt"
// psql -h localhost -U postgres -d microblog -f migrations/001_create_tables.sql
// \c microblog
// SELECT * FROM users;
// SELECT * FROM posts;
// SELECT * FROM images;
// DROP DATABASE IF EXISTS microblog;
// CREATE DATABASE microblog;
