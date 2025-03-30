//lint:file-ignore ST1006 generated
package classptrs

import (
	"fmt"
	"math"
)

type IShape interface {
	CalculateArea() float64
}

type Shape struct {
	face   IShape
	color_ string
}

var _ IShape = (*Shape)(nil)

func NewShape() *Shape {
	this := &Shape{}
	this.face = this
	return this
}

func SuperShape(face IShape) Shape {
	return Shape{face: face}
}

func ShapeConstructor(this *Shape, color_ string) *Shape {
	this.color_ = color_
	return this
}

func (this *Shape) CalculateArea() float64 {
	return 0.0
}

type ICircle interface {
	IShape
}

type Circle struct {
	Shape
	face    ICircle
	radius_ float64
}

var _ ICircle = (*Circle)(nil)

func NewCircle() *Circle {
	this := &Circle{}
	this.Shape = SuperShape(this)
	this.face = this
	return this
}

func SuperCircle(face ICircle) Circle {
	return Circle{face: face}
}

func CircleConstructor(this *Circle, color_ string, radius_ float64) *Circle {
	ShapeConstructor(&this.Shape, color_)
	this.radius_ = radius_
	return this
}

func (this *Circle) CalculateArea() float64 {
	return math.Pi * this.radius_ * this.radius_
}

type IRectangle interface {
	IShape
}

type Rectangle struct {
	Shape
	face    IRectangle
	width_  float64
	height_ float64
}

var _ IRectangle = (*Rectangle)(nil)

func SuperRectangle(face IRectangle) Rectangle {
	return Rectangle{face: face}
}

func NewRectangle() *Rectangle {
	this := &Rectangle{}
	this.Shape = SuperShape(this)
	this.face = this
	return this
}

func RectangleConstructor(this *Rectangle, color_ string, width_, height_ float64) *Rectangle {
	ShapeConstructor(&this.Shape, color_)
	this.width_ = width_
	this.height_ = height_
	return this
}

func (this *Rectangle) CalculateArea() float64 {
	return this.width_ * this.height_
}

func ExampleShape() {
	rect := RectangleConstructor(NewRectangle(), "red", 1.0, 1.0)
	circle := CircleConstructor(NewCircle(), "blue", 1.0)

	describeShape(&rect.Shape)
	describeShape(&circle.Shape)

	// Output:
	// 1
	// red
	// 3.141592653589793
	// blue
}

func describeShape(shape *Shape) {
	fmt.Println(shape.face.CalculateArea())
	fmt.Println(shape.color_)
}
