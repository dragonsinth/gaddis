package classvals_test

import (
	"fmt"
	"math"
)

type IShape interface {
	CalculateArea() float64
}

type Shape struct {
	color string
}

var _ IShape = (*Shape)(nil)

func (s *Shape) CalculateArea() float64 {
	return 0.0
}

type ShapeVal struct {
	IFace IShape
	IBody *Shape
}

type ICircle interface {
	IShape
}

type Circle struct {
	Shape
	radius float64
}

var _ ICircle = (*Circle)(nil)

func (c *Circle) CalculateArea() float64 {
	return math.Pi * c.radius * c.radius
}

type CircleVal struct {
	IFace ICircle
	IBody *Circle
}

func newCircle(v Circle) CircleVal {
	return CircleVal{IFace: &v, IBody: &v}
}

type IRectangle interface {
	IShape
}

type Rectangle struct {
	Shape
	width, height float64
}

var _ IRectangle = (*Rectangle)(nil)

func (r *Rectangle) CalculateArea() float64 {
	return r.width * r.height
}

type RectangleVal struct {
	IFace IRectangle
	IBody *Rectangle
}

func newRectangle(v Rectangle) RectangleVal {
	return RectangleVal{IFace: &v, IBody: &v}
}

func ExampleShape() {
	rect := newRectangle(Rectangle{
		Shape:  Shape{color: "red"},
		width:  1.0,
		height: 1.0,
	})

	circle := newCircle(Circle{
		Shape:  Shape{color: "blue"},
		radius: 1.0,
	})

	describeShape(ShapeVal{IFace: rect.IFace, IBody: &rect.IBody.Shape})
	describeShape(ShapeVal{IFace: circle.IFace, IBody: &circle.IBody.Shape})

	// Output:
	// 1
	// red
	// 3.141592653589793
	// blue
}

func describeShape(value ShapeVal) {
	fmt.Println(value.IFace.CalculateArea())
	fmt.Println(value.IBody.color)
}
