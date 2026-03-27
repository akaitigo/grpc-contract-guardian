// Package analyzer provides proto file parsing and dependency extraction.
// It reads .proto files and builds a structured representation of services,
// messages, and their field-level dependencies.
package analyzer

// ProtoFile represents a parsed .proto file with its services and messages.
type ProtoFile struct {
	// Path is the file path of the .proto file.
	Path string
	// Package is the proto package name (e.g., "myservice.v1").
	Package string
	// Services defined in this file.
	Services []Service
	// Messages defined in this file.
	Messages []Message
}

// Service represents a gRPC service definition.
type Service struct {
	// Name is the service name.
	Name string
	// Methods are the RPC methods defined in this service.
	Methods []Method
}

// Method represents an RPC method in a service.
type Method struct {
	// Name is the method name.
	Name string
	// InputType is the fully qualified input message type.
	InputType string
	// OutputType is the fully qualified output message type.
	OutputType string
}

// Message represents a protobuf message definition.
type Message struct {
	// Name is the message name.
	Name string
	// Fields are the fields defined in this message.
	Fields []Field
}

// Field represents a field in a protobuf message.
type Field struct {
	// Name is the field name.
	Name string
	// Number is the field number.
	Number int32
	// Type is the field type (e.g., "string", "int32", or a message type reference).
	Type string
}

// Analyze parses a .proto file and returns its structured representation.
// Returns an error if the file cannot be read or parsed.
func Analyze(path string) (*ProtoFile, error) {
	// Placeholder: will be implemented with protocompile in Issue #2
	return &ProtoFile{
		Path: path,
	}, nil
}
