## discriminatedunion

Check exhaustiveness of switch statements against discriminated unions.

```
go install github.com/thematthopkins/discriminatedunion/cmd/discriminatedunion@latest
```

Go frequently uses interfaces to signify discriminated unions.  One bit that is lacking 
though is it does not have a way of enforcing that switches exhaustively handle
all members of a discriminated union.  This analyzer performs that check, so that
if you add a new member to a discriminated union, you can have confidence that 
all type switches will be forced to account for the new member.

## Example

Given a discriminated union defined by:

```go
package shape

type Circle struct {
	float32 Radius
}

type Square struct {
	float32 Width
}

type Arc struct {
	float32 Radius
	float32 Radians
}

type Shape interface {
    // IsShape takes the form Is<interface>, which tells the discriminatedunion
    // checker that it is a discriminated union.
	IsShape()
}

// The receivers below signify which structs are members of the Shape discriminated union
func (Circle) IsShape() {}
func (Square) IsShape() {}
func (Arc) IsShape()    {}
```

with the code below:

```go
package shapeprinter

import (
    "fmt"
    "shape"
)

func printShape(s shape.Shape) {
	switch s := s {
	case shape.Circle:
        fmt.Printf("Circle Radius %v\n", s.Radius)
	case shape.Square:
        fmt.Printf("Square Width %v\n", s.Width)
	}
}
```

running `discriminatedunion ./...`, or adding it to your `go/analysis/multichecker` will print:

```
shapeprinter.go:14:3: missing cases for discriminated union types: shape.Arc
```
