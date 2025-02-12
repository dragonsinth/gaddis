package ast

import "fmt"

type Node interface {
	HasSourceInfo
	Visit(v Visitor)
}

type HasSourceInfo interface {
	GetSourceInfo() SourceInfo
}

type Position struct {
	Pos    int
	Line   int
	Column int
}

type SourceInfo struct {
	Start, End Position
}

func (si SourceInfo) GetSourceInfo() SourceInfo {
	return si
}

type Comment struct {
	SourceInfo
	Text string
}

type Error struct {
	SourceInfo
	Desc string
}

func (err Error) Error() string {
	start := err.SourceInfo.Start
	return fmt.Sprintf("%d:%d %s", start.Line, start.Column, err.Desc)
}
