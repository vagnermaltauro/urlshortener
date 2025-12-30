package repository

// IDGenerator defines the interface for generating unique IDs.
// This interface lives in the domain layer and is implemented by adapters (e.g., Snowflake).
type IDGenerator interface {
	// Generate creates a new unique 64-bit ID
	// The ID is guaranteed to be unique across all machines and time
	Generate() (int64, error)

	// Encode converts a 64-bit ID to a Base62 string (7 characters)
	// This creates the short code used in URLs
	Encode(id int64) string

	// Decode converts a Base62 string back to a 64-bit ID
	// Returns an error if the string is invalid
	Decode(code string) (int64, error)
}
