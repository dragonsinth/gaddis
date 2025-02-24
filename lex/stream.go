package lex

type Stream struct {
	buf    string
	pos    int
	line   int
	column int
}

func (s *Stream) Peek() byte {
	return s.buf[s.pos]
}

func (s *Stream) Next() byte {
	ret := s.buf[s.pos]
	s.pos++
	s.column++
	if ret == '\n' {
		s.line++
		s.column = 0
	}
	return ret
}

func (s *Stream) Eof() bool {
	return s.pos >= len(s.buf)
}
