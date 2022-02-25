package core

type Pusher interface {
	// Get the name of the storage
	// Will be unique across all storages
	Name() string
	// Method call when insert new data in the storage
	Push(value interface{}) error
}
