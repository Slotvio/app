package postgres

import "fmt"

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) (*MasterRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("user repository: db is nil")
	}
	return &MasterRepo{db: db}, nil
}


