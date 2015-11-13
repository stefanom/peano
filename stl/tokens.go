package stl

// Token represents a lexical token.
type Token int

const (
	ILLEGAL Token = iota
	EOF
	WS
	NUMBER
	SOLID
	FACET
	NORMAL
	OUTER
	LOOP
	VERTEX
	ENDLOOP
	ENDFACET
	ENDSOLID
)
