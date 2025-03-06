package ast

import (
	"fmt"
	"slices"
)

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

func (si SourceInfo) Head() SourceInfo {
	return SourceInfo{
		Start: si.Start,
		End:   si.Start,
	}
}

func (si SourceInfo) Tail() SourceInfo {
	return SourceInfo{
		Start: si.End,
		End:   si.End,
	}
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

func ErrorSort(errors []Error) []Error {
	slices.SortFunc(errors, func(a, b Error) int {
		return a.Start.Pos - b.Start.Pos
	})
	return slices.CompactFunc(errors, func(a, b Error) bool {
		return a.Error() == b.Error()
	})
}
