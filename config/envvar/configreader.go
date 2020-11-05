package envvar

// ConfigReader lists the methods exposed by ConfigReaderImpl.
type ConfigReader interface {
	// Get returns an interface{}.
	// For a specific value use one of the Get____ methods.
	Get(key string) (interface{}, error)

	// GetString retrieves the associated key value as a string.
	GetString(key string) (string, error)

	// Unmarshal deserializes the loaded cofig into a struct.
	Unmarshal(config interface{}) error
}

// NilValueError is raised when the key value is nil.
type NilValueError struct {
	message string
}

func (e NilValueError) Error() string {
	return e.message
}

// ValueConversionError is raised when the key value is incompatible with the invoked method.
type ValueConversionError struct {
	message string
}

func (e ValueConversionError) Error() string {
	return e.message
}
