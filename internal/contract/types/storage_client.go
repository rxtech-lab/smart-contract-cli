package types

type StorageClient string

const (
	StorageClientSQLite   StorageClient = "sqlite"
	StorageClientPostgres StorageClient = "postgres"
)
