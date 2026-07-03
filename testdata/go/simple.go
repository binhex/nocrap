package simple

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Multiply returns the product of two integers.
func Multiply(a, b int) int {
	return a * b
}

// Greeter is a simple greeter struct.
type Greeter struct {
	name string
}

// NewGreeter creates a new Greeter.
func NewGreeter(name string) *Greeter {
	return &Greeter{name: name}
}

// Greet returns a greeting message.
func (g *Greeter) Greet() string {
	return "Hello, " + g.name
}
