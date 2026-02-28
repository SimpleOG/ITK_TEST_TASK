package mocks

//go:generate mockgen -source=../../internal/repository/repository.go -destination=mock_repository.go -package=mocks
//go:generate mockgen -source=../../internal/repository/querier.go -destination=mock_querier.go -package=mocks
