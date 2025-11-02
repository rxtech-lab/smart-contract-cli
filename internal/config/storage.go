package config

const (
	// SecureStorageKeySqlitePathKey is the key for the sqlite path in secure storage.
	SecureStorageKeySqlitePathKey = "storage_client_sqlite_path"
	// SecureStorageKeyPostgresURLKey is the key for the postgres url in secure storage.
	SecureStorageKeyPostgresURLKey = "storage_client_postgres_url"
	// SecureStoragePasswordKey is the key for the hashed password in secure storage.
	SecureStoragePasswordKey = "secure_storage_password"
	// SecureStorageClientTypeKey is the key for the storage client type in secure storage.
	// can be sqlite or postgres.
	SecureStorageClientTypeKey = "storage_client_type"
)
