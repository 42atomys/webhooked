package core

type Storage interface {
	// Get the name of the storage
	// Will be unique across all storages
	Name() string

	// TODO: Add methods to define the storage
}
