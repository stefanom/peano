package main

import (
	"flag"
	"fmt"
	"github.com/stefanom/slicer/geom"
	"github.com/stefanom/slicer/stl"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ExportSliceAsSVG(slicename string, segments *[]geom.Segment, paths *[]geom.Path) {
	f, err := os.Create(slicename)
	check(err)
	defer f.Close()

	fmt.Fprintln(f, "<svg xmlns=\"http://www.w3.org/2000/svg\">")
	// fmt.Fprintln(f, " <g transform=\"translate(200,200)\">")
	// fmt.Fprintf(f, "  <path stroke=\"black\" d=\"M %v,%v ", (*paths)[0][0], (*paths)[0][1])
	// for _, path := range (*paths)[1:] {
	// 	fmt.Fprintf(f, "L %v,%v", path[0], path[1])
	// }
	// fmt.Fprintln(f, "\"/>")
	// fmt.Fprintln(f, " </g>")
	fmt.Fprintln(f, " <g transform=\"translate(400,200)\">")
	for _, segment := range *segments {
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
	reader, err := os.Open(filename)
	check(err)
	parser := stl.NewParser(reader)
	model, err := parser.Parse()
	check(err)

	//	facetsByLayer, segmentsByLayer, minLayer, maxLayer := FacetsByLayer(model, layerHeight)
	_, segmentsByLayer, minLayer, maxLayer := geom.FacetsByLayer(&model.Facets, layerHeight)

	fmt.Println("got slices")

	for i := minLayer; i <= maxLayer; i += 1 {
		slicename := fmt.Sprintf("%s.%d.svg", filename, i)
		fmt.Println("writing: ", slicename)
		// for point, segment := range (*facetsByLayer)[i] {
		// 	fmt.Printf("  %v: %v\n", point, segment)
		// }
		segments := (*segmentsByLayer)[i]
		for _, segment := range segments {
			fmt.Printf("  %v -> %v\n", segment.Start, segment.End)
		}
		//paths := new([]Path)
		// paths := PathsFromSegments(segments)
		//ExportSliceAsSVG(slicename, &segments, paths)
	}

	if exportAscii {
		serializer := stl.NewSerializer(os.Stdout)
		serializer.SerializeAsAscii(filename, model)
	}
}
