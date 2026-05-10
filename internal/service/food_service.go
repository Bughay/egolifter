package service

import (
	"github.com/Bughay/egolifter/sql/sqlc"
)

type FoodService struct {
	queries *sqlc.Queries
}
