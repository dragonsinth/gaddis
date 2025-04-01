//lint:file-ignore ST1006 generated
package classptrs

import (
	"fmt"
	"math"
)

type IShape interface {
	SetColor_(color_ string)
	CalculateArea_() float64
}

type Shape_ struct {
	face   IShape
	color_ string
}

var _ IShape = (*Shape_)(nil)

func NewShape() *Shape_ {
	this := &Shape_{}
	this.face = this
	return this
}

func SuperShape(face IShape) Shape_ {
	return Shape_{face: face}
}

func (this *Shape_) Shape_(color_ string) *Shape_ {
	this.color_ = color_
	return this
}

func (this *Shape_) SetColor_(color_ string) {
	this.color_ = color_
}

func (this *Shape_) CalculateArea_() float64 {
	return 0.0
}

type ICircle interface {
	IShape
}

type Circle_ struct {
	Shape_
	face    ICircle
	radius_ float64
}

var _ ICircle = (*Circle_)(nil)

func NewCircle() *Circle_ {
	this := &Circle_{}
	this.Shape_ = SuperShape(this)
	this.face = this
	return this
}

func SuperCircle(face ICircle) Circle_ {
	return Circle_{
		Shape_: SuperShape(face),
		face:   face,
	}
}

func CircleConstructor(this *Circle_, color_ string, radius_ float64) *Circle_ {
	this.Shape_.Shape_(color_)
	this.radius_ = radius_
	return this
}

func (this *Circle_) CalculateArea_() float64 {
	return math.Pi * this.radius_ * this.radius_
}

type IRectangle interface {
	IShape
}

type Rectangle_ struct {
	Shape_
	face    IRectangle
	width_  float64
	height_ float64
}

var _ IRectangle = (*Rectangle_)(nil)

func SuperRectangle(face IRectangle) Rectangle_ {
	return Rectangle_{
		Shape_: SuperShape(face),
		face:   face,
	}
}

func NewRectangle() *Rectangle_ {
	this := &Rectangle_{}
	this.Shape_ = SuperShape(this)
	this.face = this
	return this
}

func (this *Rectangle_) Rectangle_(color_ string, width_, height_ float64) *Rectangle_ {
	this.Shape_.Shape_(color_)
	this.width_ = width_
	this.height_ = height_
	return this
}

func (this *Rectangle_) CalculateArea_() float64 {
	return this.width_ * this.height_
}

type ISquare interface {
	IRectangle
}

type Square_ struct {
	Rectangle_
	face ISquare
}

var _ ISquare = (*Square_)(nil)

func NewSquare() *Square_ {
	this := &Square_{}
	this.Rectangle_ = SuperRectangle(this)
	this.face = this
	return this
}

func SuperSquare(face ISquare) Square_ {
	return Square_{
		Rectangle_: SuperRectangle(face),
		face:       face,
	}
}

func SquareConstructor(this *Square_, color_ string, side_ float64) *Square_ {
	this.Rectangle_.Rectangle_(color_, side_, side_)
	return this
}

func Example() {
	var rect_ *Rectangle_ = NewRectangle().Rectangle_("red", 1.0, 1.0)
	var circle_ *Circle_ = CircleConstructor(NewCircle(), "blue", 1.0)
	var square_ *Square_ = SquareConstructor(NewSquare(), "green", 4.0)

	describeShape_(&rect_.Shape_)
	describeShape_(&circle_.Shape_)
	describeShape_(&square_.Rectangle_.Shape_)

	// Output:
	// 1
	// red
	// 3.141592653589793
	// blue
	// 16
	// green
}

func describeShape_(shape *Shape_) {
	fmt.Println(shape.face.CalculateArea_())
	fmt.Println(shape.color_)
}
