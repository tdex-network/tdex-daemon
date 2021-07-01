package interfaces

// Service interface defines the methods that every kind of interface, whether
// gRPC, REST, or whatever must be compliant with.
type Service interface {
	Start() error
	Stop()
}
