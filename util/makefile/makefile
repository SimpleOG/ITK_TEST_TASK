sqlc:
	cd ../sqlc && sqlc generate
test_service:
	go test ./internal/service/... -v
test_race_service:
	go test -race ./internal/service/...
test_api:
	go test ./internal/api/... -v
test_race_api:
	go test -race ./internal/api/... -v
test_race_locker:
	go test ./internal/service/wallet/... -run TestWalletLocker -race -v