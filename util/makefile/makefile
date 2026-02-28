sqlc:
	cd ../sqlc && sqlc generate
test_service:
	go test ./internal/service/... -v
race_service:
	go test -race ./internal/service/...