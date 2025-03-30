package classvals_test

import (
	"fmt"
	"math"
)

type IShape interface {
	CalculateArea_() float64
}

type Shape struct {
	color string
}

var _ IShape = (*Shape)(nil)

func (s *Shape) CalculateArea_() float64 {
	return 0.0
}

type Shape_ struct {
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

func (c *Circle) CalculateArea_() float64 {
	return math.Pi * c.radius * c.radius
}

type Circle_ struct {
	IFace ICircle
	IBody *Circle
}

func newCircle(v Circle) Circle_ {
	return Circle_{IFace: &v, IBody: &v}
}

type IRectangle interface {
	IShape
}

type Rectangle struct {
	Shape
	width, height float64
}

var _ IRectangle = (*Rectangle)(nil)

func (r *Rectangle) CalculateArea_() float64 {
	return r.width * r.height
}

type Rectangle_ struct {
	IFace IRectangle
	IBody *Rectangle
}

func newRectangle(v Rectangle) Rectangle_ {
	return Rectangle_{IFace: &v, IBody: &v}
}

type ISquare interface {
	IRectangle
}

type Square struct {
	Rectangle
}

var _ ISquare = (*Square)(nil)

type Square_ struct {
	IFace ISquare
	IBody *Square
}

func newSquare(v Square) Square_ {
	return Square_{IFace: &v, IBody: &v}
}

func Example() {
	rect := newRectangle(Rectangle{
		Shape:  Shape{color: "red"},
		width:  1.0,
		height: 1.0,
	})

	circle := newCircle(Circle{
		Shape:  Shape{color: "blue"},
		radius: 1.0,
	})

	square := newSquare(Square{
		Rectangle: Rectangle{
			Shape:  Shape{color: "green"},
			width:  4.0,
			height: 4.0,
		},
	})

	describeShape(Shape_{IFace: rect.IFace, IBody: &rect.IBody.Shape})
	describeShape(Shape_{IFace: circle.IFace, IBody: &circle.IBody.Shape})
	describeShape(Shape_{IFace: square.IFace, IBody: &square.IBody.Shape})

	// Output:
	// 1
	// red
	// 3.141592653589793
	// blue
	// 16
	// green
}

func describeShape(value Shape_) {
	fmt.Println(value.IFace.CalculateArea_())
	fmt.Println(value.IBody.color)
}
