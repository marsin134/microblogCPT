package service

import "microblogCPT/internal/repository"

type TablesService interface {
	GetCountTablesBD(req repository.TablesRepository) (int, error)
}

type tablesService struct {
	tablesRepo repository.TablesRepository
}

func NewTablesService(tablesRepo repository.TablesRepository) TablesService {
	return &tablesService{tablesRepo: tablesRepo}
}

func (t *tablesService) GetCountTablesBD(req repository.TablesRepository) (int, error) {
	countTables, err := req.CountTablesDB()
	if err != nil {
		return 0, err
	}

	return countTables, nil
}
