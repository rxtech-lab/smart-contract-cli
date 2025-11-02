package tools

//go:generate go run ./routergen -dir ../app -module-root ..
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/wallet/service.go -destination=../internal/contract/evm/wallet/mock_service.go -package=wallet
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/storage/sql/storage.go -destination=../internal/contract/evm/storage/sql/mock_storage.go -package=sql
//go:generate go run go.uber.org/mock/mockgen -source=../internal/storage/secure.go -destination=../internal/storage/mock_secure.go -package=storage
