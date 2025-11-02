package tools

// This file contains go:generate directives for code generation.
// Run `make generate` to execute all generators.

// Generate routes from app folder structure (Next.js-style file-based routing)
//go:generate go run ./routergen -dir ../app -module-root ..

// Generate mocks for testing using mockgen
// Pattern: mockgen -source=<interface_file> -destination=<mock_file> -package=<package_name>
//
// IMPORTANT: Always add new mockgen directives here when creating new interfaces.
// This ensures mocks are regenerated when running `make generate`.
//
// Example for adding a new mock:
// //go:generate go run go.uber.org/mock/mockgen -source=../internal/path/to/interface.go -destination=../internal/path/to/mock_interface.go -package=packagename

// Core service mocks
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/wallet/service.go -destination=../internal/contract/evm/wallet/mock_service.go -package=wallet

// Storage mocks
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/storage/sql/storage.go -destination=../internal/contract/evm/storage/sql/mock_storage.go -package=sql
//go:generate go run go.uber.org/mock/mockgen -source=../internal/storage/secure.go -destination=../internal/storage/mock_secure.go -package=storage

// View layer mocks
//go:generate go run go.uber.org/mock/mockgen -source=../internal/view/types.go -destination=../internal/view/mock_router.go -package=view
