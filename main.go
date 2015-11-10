package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Model struct {
	Header [80]byte
	Length int32
	Facets []Facet
}

type Vector [3]float32

type Facet struct {
	Normal    Vector
	Vertex1   Vector
	Vertex2   Vector
	Vertex3   Vector
	Attribute uint16
}

type Point [2]float32

type Segment struct {
	Start  Point
	End    Point
	Normal *Vector
}

func ParseSTL(filename string) *Model {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	m := new(Model)

	err = binary.Read(bytes.NewBuffer(data[0:80]), binary.LittleEndian, &m.Header)
	if err != nil {
		panic(err)
	}

	err = binary.Read(bytes.NewBuffer(data[80:84]), binary.LittleEndian, &m.Length)
	if err != nil {
		panic(err)
	}

	m.Facets = make([]Facet, m.Length)

	err = binary.Read(bytes.NewBuffer(data[84:]), binary.LittleEndian, &m.Facets)
	if err != nil {
		panic(err)
	}

	return m
}

func SaveASCII(name string, model *Model) {
	fmt.Printf("solid %v\n", name)
	for _, facet := range model.Facets {
		fmt.Printf("  facet normal %E %E %E\n", facet.Normal[0], facet.Normal[1], facet.Normal[2])
		fmt.Printf("    outer loop\n")
		fmt.Printf("      vertex %E %E %E\n", facet.Vertex1[0], facet.Vertex1[1], facet.Vertex1[2])
		fmt.Printf("      vertex %E %E %E\n", facet.Vertex2[0], facet.Vertex2[1], facet.Vertex2[2])
		fmt.Printf("      vertex %E %E %E\n", facet.Vertex3[0], facet.Vertex3[1], facet.Vertex3[2])
		fmt.Printf("    endloop\n")
		fmt.Printf("  endfacet\n")
	}
	fmt.Printf("endsolid %v\n", name)
}

func Intersect(a, b, c, normal *Vector, layerZ float32) Segment {
	s := new(Segment)
	s.Start = *new(Point)
	s.End = *new(Point)
	s.Normal = normal

	dlaz := layerZ - a[2]
	dbaz := b[2] - a[2]
	dcaz := c[2] - a[2]

	t1 := dlaz / dbaz
	s.Start[0] = t1*(b[0]-a[0]) + a[0]
	s.Start[1] = t1*(b[1]-a[1]) + a[1]

	t2 := dlaz / dcaz
	s.End[0] = t2*(c[0]-a[0]) + a[0]
	s.End[1] = t2*(c[1]-a[1]) + a[1]

	return *s
}

func FacetsByLayer(model *Model, layerHeight float64) (*map[int32][]Segment, int32, int32) {
	m := make(map[int32][]Segment)
	globalMinLayer := int32(0)
	globalMaxLayer := int32(0)
	for _, facet := range model.Facets {
		var max *Vector
		var mid *Vector
		var min *Vector

		if facet.Vertex1[2] > facet.Vertex2[2] {
			max = &facet.Vertex1
			mid = &facet.Vertex2
		} else {
			max = &facet.Vertex2
			mid = &facet.Vertex1
		}
		if facet.Vertex3[2] > max[2] {
			min = mid
			mid = max
			max = &facet.Vertex3
		} else if facet.Vertex3[2] > mid[2] {
			min = mid
			mid = &facet.Vertex3
		} else {
			min = &facet.Vertex3
		}

		// fmt.Println(min, mid, max)

		minLayer := int32(math.Ceil(float64(min[2]) / layerHeight))
		midLayerBelow := int32(math.Floor(float64(mid[2]) / layerHeight))
		midLayerAbove := int32(math.Ceil(float64(mid[2]) / layerHeight))
		maxLayer := int32(math.Floor(float64(max[2]) / layerHeight))

		if minLayer < globalMinLayer {
			globalMinLayer = minLayer
		}
		if maxLayer > globalMaxLayer {
			globalMaxLayer = maxLayer
		}

		// fmt.Println(minLayer, midLayerBelow, midLayerAbove, maxLayer)

		for layer := minLayer; layer <= midLayerBelow; layer += 1 {
			segment := Intersect(min, mid, max, &facet.Normal, float32(layer)*float32(layerHeight))
			m[layer] = append(m[layer], segment)
		}
		for layer := midLayerAbove; layer <= maxLayer; layer += 1 {
			segment := Intersect(max, mid, min, &facet.Normal, float32(layer)*float32(layerHeight))
			m[layer] = append(m[layer], segment)
		}
	}

	return &m, globalMinLayer, globalMaxLayer
}

func ExportSliceAsSVG(slicename string, segments []Segment) {
	f, err := os.Create(slicename)
	check(err)
	defer f.Close()

	fmt.Fprintln(f, "<svg xmlns=\"http://www.w3.org/2000/svg\">")
	fmt.Fprintln(f, " <g transform=\"translate(200,200)\">")
	for _, segment := range segments {
		fmt.Fprintf(f, "  <line x1=\"%v\" y1=\"%v\" x2=\"%v\" y2=\"%v\" stroke=\"black\" stroke-width=\"1\" />\n", segment.Start[0], segment.Start[1], segment.End[0], segment.End[1])
	}
	fmt.Fprintln(f, " </g>")
	fmt.Fprintln(f, "</svg>")
}

var filename string
var layerHeight float64
var exportAscii bool

func init() {
	flag.StringVar(&filename, "file", "", "The filename of the STL file to parse.")
	flag.Float64Var(&layerHeight, "layerHeight", 0.2, "The layer height to use for slicing.")
	flag.BoolVar(&exportAscii, "exportAscii", false, "Whether to export the parsed STL file as ASCII.")
	flag.Parse()
}

func main() {
	model := ParseSTL(filename)
	facetsByLayer, minLayer, maxLayer := FacetsByLayer(model, layerHeight)

	for i := minLayer; i <= maxLayer; i += 1 {
		slicename := fmt.Sprintf("%s.%d.svg", filename, i)
		segments := (*facetsByLayer)[i]
		ExportSliceAsSVG(slicename, segments)
	}

	if exportAscii {
		SaveASCII(filename, model)
	}
}

// TODO
//  - implement guard agaist ascii STL parsing
