package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

func mergeSourceInfo(a ast.HasSourceInfo, b ast.HasSourceInfo) ast.SourceInfo {
	return ast.SourceInfo{Start: a.GetSourceInfo().Start, End: b.GetSourceInfo().End}
}

func spanAst(start lex.Result, end ast.HasSourceInfo) ast.SourceInfo {
	return ast.SourceInfo{
		Start: toPos(start.Pos),
		End:   end.GetSourceInfo().End,
	}
}

func spanResult(start lex.Result, end lex.Result) ast.SourceInfo {
	return ast.SourceInfo{
		Start: toPos(start.Pos),
		End:   toSourceInfo(end).End,
	}
}

func toSourceInfo(r lex.Result) ast.SourceInfo {
	start := toPos(r.Pos)
	end := start
	end.Pos += len(r.Text)
	end.Column += len(r.Text)
	return ast.SourceInfo{
		Start: start,
		End:   end,
	}
}

func toPos(p lex.Position) ast.Position {
	return ast.Position{
		Pos:    p.Pos,
		Line:   p.Line,
		Column: p.Column,
	}
}
