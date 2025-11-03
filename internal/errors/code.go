package errors

// ErrorCode represents a unique identifier for different error types.
type ErrorCode string

const (
	// ABI Domain Error Codes.
	ErrCodeInvalidABIFormat    ErrorCode = "INVALID_ABI_FORMAT"
	ErrCodeABIConversionFailed ErrorCode = "ABI_CONVERSION_FAILED"
	ErrCodeABIMarshalFailed    ErrorCode = "ABI_MARSHAL_FAILED"
	ErrCodeABIParseFailed      ErrorCode = "ABI_PARSE_FAILED"
	ErrCodeABIPackFailed       ErrorCode = "ABI_PACK_FAILED"
	ErrCodeABIUnpackFailed     ErrorCode = "ABI_UNPACK_FAILED"
	ErrCodeMethodNotFound      ErrorCode = "METHOD_NOT_FOUND"

	// Signer Domain Error Codes.
	ErrCodeInvalidPrivateKey      ErrorCode = "INVALID_PRIVATE_KEY"
	ErrCodeSigningFailed          ErrorCode = "SIGNING_FAILED"
	ErrCodeTransactionSignFailed  ErrorCode = "TRANSACTION_SIGN_FAILED"
	ErrCodeInvalidSignature       ErrorCode = "INVALID_SIGNATURE"
	ErrCodeInvalidSignatureLength ErrorCode = "INVALID_SIGNATURE_LENGTH"
	ErrCodeSignatureDecode        ErrorCode = "SIGNATURE_DECODE_FAILED"
	ErrCodePublicKeyRecovery      ErrorCode = "PUBLIC_KEY_RECOVERY_FAILED"

	// Transport Domain Error Codes.
	ErrCodeEndpointRequired      ErrorCode = "ENDPOINT_REQUIRED"
	ErrCodeConnectionFailed      ErrorCode = "CONNECTION_FAILED"
	ErrCodeRPCCallFailed         ErrorCode = "RPC_CALL_FAILED"
	ErrCodeTransactionTimeout    ErrorCode = "TRANSACTION_TIMEOUT"
	ErrCodeInvalidChainID        ErrorCode = "INVALID_CHAIN_ID"
	ErrCodeTransactionSendFailed ErrorCode = "TRANSACTION_SEND_FAILED"
	ErrCodeGasEstimateFailed     ErrorCode = "GAS_ESTIMATE_FAILED"
	ErrCodeBalanceQueryFailed    ErrorCode = "BALANCE_QUERY_FAILED"
	ErrCodeNonceQueryFailed      ErrorCode = "NONCE_QUERY_FAILED"
	ErrCodeReceiptQueryFailed    ErrorCode = "RECEIPT_QUERY_FAILED"
	ErrCodeChainIDQueryFailed    ErrorCode = "CHAIN_ID_QUERY_FAILED"

	// Contract Domain Error Codes.
	ErrCodeContractCodeRequired  ErrorCode = "CONTRACT_CODE_REQUIRED"
	ErrCodeContractCompileFailed ErrorCode = "CONTRACT_COMPILE_FAILED"

	// Database Domain Error Codes.
	ErrCodeRecordNotFound          ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDatabaseOperationFailed ErrorCode = "DATABASE_OPERATION_FAILED"
	ErrCodeInvalidPageNumber       ErrorCode = "INVALID_PAGE_NUMBER"
	ErrCodeInvalidPageSize         ErrorCode = "INVALID_PAGE_SIZE"

	// Storage Domain Error Codes.
	ErrCodeInvalidStorageClientType    ErrorCode = "INVALID_STORAGE_CLIENT_TYPE"
	ErrCodeStorageClientNotInitialized ErrorCode = "STORAGE_CLIENT_NOT_INITIALIZED"

	// General Error Codes.
	ErrCodeUnknown ErrorCode = "UNKNOWN"
)

// String returns the string representation of the error code.
func (e ErrorCode) String() string {
	return string(e)
}
