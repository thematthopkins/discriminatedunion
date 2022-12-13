package missing_case

type Circle struct {
	Radius float32
}

type Square struct {
	Width float32
}

type Arc struct {
	Radius  float32
	Radians float32
}

type Shape interface { // want Shape: "missing_case.Circle, missing_case.Square, missing_case.Arc"
	IsShape()
}

func (Circle) IsShape() {}
func (Square) IsShape() {}
func (Arc) IsShape()    {}

func PrintShape(s Shape) {
	switch s.(type) {
	case Circle:
	} // want "^missing cases for discriminated union types: missing_case.Square, missing_case.Arc$"

	switch s.(type) {
	case Circle, Square:
	} // want "^missing cases for discriminated union types: missing_case.Arc$"
}
