package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/stefanom/peano/printer"
	"log"
	"math"
	"os"
	"strings"
)

type Curve struct {
	origin Point
	points []Point
}

type Point struct {
	x, y float64
}

func main() {
	filename := flag.String("file", "", "the file containing the movements commands")
	scale := flag.Float64("scale", 5.0, "the scale factor (# of pixels per mm)")

	spread := flag.Float64("spread", 1.0, "the extrusion spread")
	speed := flag.Float64("speed", 20.0, "the head movement speed when extruding (in mm/sec)")
	temp := flag.Float64("temp", 210.0, "the temperature of extrusion (in degrees celcius)")

	flag.Parse()

	curves := make([]*Curve, 0)

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	maxX := -math.MaxFloat64
	maxY := -math.MaxFloat64
	minX := math.MaxFloat64
	minY := math.MaxFloat64

	var curve *Curve
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// ignore comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// parse motion line
		var command string
		p := new(Point)
		if _, err := fmt.Sscan(line, &command, &p.x, &p.y); err != nil {
			log.Fatal(err)
		}

		// scale the point location in mm coordinates from pixel coordinates
		p.x /= *scale
		p.y /= *scale

		// interpret the motion command
		if command == "m" {
			curve = new(Curve)
			curve.origin = *p
			curve.points = make([]Point, 0)
			curves = append(curves, curve)
		} else if command == "d" {
			curve.points = append(curve.points, *p)
		}

		// evaluate whether we have moved beyond our existin limits
		if p.x > maxX {
			maxX = p.x
		}
		if p.x < minX {
			minX = p.x
		}
		if p.y > maxY {
			maxY = p.y
		}
		if p.y < minY {
			minY = p.y
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	p := printer.Printer{
		Output:           os.Stdout,
		Temperature:      *temp,
		TravelSpeed:      150.0,
		PrintSpeed:       *speed,
		FlowCorrection:   1.0,
		CenterX:          175.0,
		CenterY:          100.0,
		LayerHeight:      0.2,
		FilamentDiameter: 2.85,
		RetractionSpeed:  20.0,
		RetractionLength: 2.0,
	}

	skirtDistance := 10.0

	p.Preamble()

	p.Raise()

	// print skirt
	p.Comment("skirt")
	p.Move(minX-skirtDistance-p.LayerHeight, minY-skirtDistance-p.LayerHeight)
	p.Print(maxX+skirtDistance+p.LayerHeight, minY-skirtDistance-p.LayerHeight, 1.0)
	p.Print(maxX+skirtDistance+p.LayerHeight, maxY+skirtDistance+p.LayerHeight, 1.0)
	p.Print(minX-skirtDistance-p.LayerHeight, maxY+skirtDistance+p.LayerHeight, 1.0)
	p.Print(minX-skirtDistance-p.LayerHeight, minY-skirtDistance-p.LayerHeight, 1.0)
	p.Move(minX-skirtDistance, minY-skirtDistance)
	p.Print(maxX+skirtDistance, minY-skirtDistance, 1.0)
	p.Print(maxX+skirtDistance, maxY+skirtDistance, 1.0)
	p.Print(minX-skirtDistance, maxY+skirtDistance, 1.0)
	p.Print(minX-skirtDistance, minY-skirtDistance, 1.0)

	for i := 0; i < 3; i++ {
		p.Comment("layer: %d", i)

		for _, curve := range curves {
			p.MoveAndRetract(curve.origin.x, curve.origin.y)
			for _, point := range curve.points {
				p.Print(point.x, point.y, *spread)
			}
		}

		p.Raise()
	}

	p.Postamble()
}
