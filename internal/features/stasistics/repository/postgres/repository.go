package postgres_statistics_repository

import core_postgres_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool"

type StatisticsRepository struct {
	pool core_postgres_pool.Pool
}


func NewStatisticsRepository(
	pool core_postgres_pool.Pool,
) *StatisticsRepository {
	return &StatisticsRepository{
		pool: pool,
	}
}