package stl

import (
	"fmt"
	"os"
	"testing"
)

func TestParsingAscii(t *testing.T) {
	reader, err := os.Open("test_data/cube.ascii.stl")
	if err != nil {
		panic(err)
	}
	parser := NewParser(reader)
	model, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	if len(model.Facets) != 12 {
		t.Error(fmt.Sprintf("Expected %v facets, got %v", 12, len(model.Facets)))
	}
}

func TestParsingBinary(t *testing.T) {
	reader, err := os.Open("test_data/cube.binary.stl")
	if err != nil {
		panic(err)
	}
	parser := NewParser(reader)
	model, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	if len(model.Facets) != 12 {
		t.Error(fmt.Sprintf("Expected %v facets, got %v", 12, len(model.Facets)))
	}
}
