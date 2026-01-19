package service

import "microblogCPT/internal/repository"

type TablesService interface {
	GetCountTablesBD() (int, error)
}

type tablesService struct {
	tablesRepo repository.TablesRepository
}

func NewTablesService(tablesRepo repository.TablesRepository) TablesService {
	return &tablesService{tablesRepo: tablesRepo}
}

func (t *tablesService) GetCountTablesBD() (int, error) {
	countTables, err := t.tablesRepo.CountTablesDB()
	if err != nil {
		return 0, err
	}

	return countTables, nil
}
