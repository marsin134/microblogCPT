package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type tablesRepository struct {
	db *sqlx.DB
}

func NewTablesRepository(db *sqlx.DB) TablesRepository {
	return &tablesRepository{db: db}
}

func (r *tablesRepository) CountTablesDB() (int, error) {
	var count int

	err := r.db.Get(&count, `
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = 'public'
		`)

	if err != nil {
		return 0, fmt.Errorf("ошибка при подсчёте таблиц базы данных: %w", err)
	}

	return count, nil
}
