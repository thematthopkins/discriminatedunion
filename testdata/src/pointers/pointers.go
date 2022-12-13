package pointers

type Number interface {
    int32 | int64 | float32 | float64
}

type Shape[T Number] interface{ // want Shape: "\\*pointers.Circle, \\*pointers.Square, \\*pointers.Rectangle"
    IsShape(T)
}

type Circle[T Number] struct{
    Radius T
}

type Square[T Number] struct{
    Width T
}

type Rectangle[T Number] struct{
    Width T
    Height T
}

func (*Circle[T])IsShape(T){}
func (*Square[T])IsShape(T){}
func (*Rectangle[T])IsShape(T){}

func Width[T Number](shape Shape[T]) T{
    switch shape := shape.(type) {
    case *Circle[T]:
        return shape.Radius * 2
    case *Square[T]:
        return shape.Width
    case *Rectangle[T]:
        return shape.Width
    }
    panic("unhandled")
}

func WidthMissingCases[T Number](shape Shape[T]) T{
    switch shape := shape.(type) {
    case *Circle[T]:
        return shape.Radius * 2
    } // want "^missing cases for discriminated union types: \\*pointers.Square, \\*pointers.Rectangle$"
    panic("unhandled")
}
