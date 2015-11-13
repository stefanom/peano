package stl

import (
	"bufio"
	"fmt"
	"io"
)

// Serializer represents a way to serialize an STL model.
type Serializer struct {
	w *bufio.Writer
}

// NewSerializer returns a new instance of Serializer.
func NewSerializer(w io.Writer) *Serializer {
	return &Serializer{w: bufio.NewWriter(w)}
}

func (s *Serializer) SerializeAsAscii(name string, model *Model) {
	fmt.Fprintf(s.w, "solid %v\n", name)
	for _, facet := range model.Facets {
		fmt.Fprintf(s.w, "  facet normal %E %E %E\n", facet.Normal[0], facet.Normal[1], facet.Normal[2])
		fmt.Fprintf(s.w, "    outer loop\n")
		fmt.Fprintf(s.w, "      vertex %E %E %E\n", facet.Vertex1[0], facet.Vertex1[1], facet.Vertex1[2])
		fmt.Fprintf(s.w, "      vertex %E %E %E\n", facet.Vertex2[0], facet.Vertex2[1], facet.Vertex2[2])
		fmt.Fprintf(s.w, "      vertex %E %E %E\n", facet.Vertex3[0], facet.Vertex3[1], facet.Vertex3[2])
		fmt.Fprintf(s.w, "    endloop\n")
		fmt.Fprintf(s.w, "  endfacet\n")
	}
	fmt.Fprintf(s.w, "endsolid %v\n", name)
}
