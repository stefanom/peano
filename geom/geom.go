package geom

import (
	"fmt"
	"math"
)

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
	Start   Point
	End     Point
	Normal  *Vector
	Visited bool
}

type Path []Point

func SliceAngle(a, b, c, normal *Vector, layerZ float32) Segment {
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

func FacetsByLayer(facets *[]Facet, layerHeight float64) (*map[int32]map[Point]Segment, *map[int32][]Segment, int32, int32) {
	m := make(map[int32]map[Point]Segment)
	segments := make(map[int32][]Segment)
	globalMinLayer := int32(0)
	globalMaxLayer := int32(0)

	for _, facet := range *facets {
		if facet.Normal[0] == 0 && facet.Normal[1] == 0 {
			// degenerate facet parallel to the slicing plane, so we ignore it for now
			// TODO: I think we need this to understand roofs, so ignoring it is not
			// ideal, we need to find a way to record this info somehow
			fmt.Println("Skipping facet:", facet)
			continue
		}

		var max *Vector
		var mid *Vector
		var min *Vector
		var maxIndex int
		var midIndex int
		var minIndex int
		facetType := "general"

		if facet.Vertex1[2] > facet.Vertex2[2] {
			max = &facet.Vertex1
			maxIndex = 1
			mid = &facet.Vertex2
			midIndex = 2
		} else {
			max = &facet.Vertex2
			maxIndex = 2
			mid = &facet.Vertex1
			midIndex = 1
		}

		if facet.Vertex3[2] > max[2] {
			min = mid
			minIndex = midIndex
			mid = max
			midIndex = maxIndex
			max = &facet.Vertex3
			maxIndex = 3
		} else if facet.Vertex3[2] > mid[2] {
			min = mid
			minIndex = midIndex
			mid = &facet.Vertex3
			midIndex = 3
		} else {
			min = &facet.Vertex3
			minIndex = 3
		}

		if min[2] == mid[2] {
			facetType = "resting_top"
		} else if max[2] == mid[2] {
			facetType = "resting_bottom"
		}

		fmt.Println(facet, facetType)
		fmt.Println(min, mid, max, minIndex, midIndex, maxIndex)

		// lower part of the facet
		if facetType != "resting_top" {
			fmt.Println("Processing lower part of the facet")
			minLayer := int32(math.Ceil(float64(min[2]) / layerHeight))
			midLayerBelow := int32(math.Floor(float64(mid[2]) / layerHeight))

			if minLayer < globalMinLayer {
				globalMinLayer = minLayer
			}

			origin := min

			var right *Vector
			var left *Vector

			if midIndex == (minIndex+1)%3 {
				fmt.Println("mid to the left")
				right = max
				left = mid
			} else {
				fmt.Println("mid to the right")
				right = mid
				left = max
			}

			for layer := minLayer; layer <= midLayerBelow; layer += 1 {
				fmt.Println("slicing layer ", layer)
				layerSegments, ok := m[layer]
				if !ok {
					layerSegments = make(map[Point]Segment)
					m[layer] = layerSegments
				}
				segment := SliceAngle(origin, right, left, &facet.Normal, float32(layer)*float32(layerHeight))
				fmt.Println("obtained segment: ", segment)
				if segment.Start[0] == segment.End[0] && segment.Start[1] == segment.End[1] {
					fmt.Println("degenerate segment: skipping")
				} else {
					layerSegments[segment.Start] = segment
					segments[layer] = append(segments[layer], segment)
				}
			}
		}

		// upper part of the facet
		if facetType != "resting_bottom" {
			fmt.Println("Processing upper part of the facet")
			midLayerAbove := int32(math.Ceil(float64(mid[2]) / layerHeight))
			maxLayer := int32(math.Floor(float64(max[2]) / layerHeight))

			if maxLayer > globalMaxLayer {
				globalMaxLayer = maxLayer
			}

			origin := max

			var right *Vector
			var left *Vector

			if midIndex == (maxIndex+1)%3 {
				fmt.Println("mid to the right")
				right = mid
				left = min
			} else {
				fmt.Println("mid to the left")
				right = min
				left = mid
			}

			for layer := midLayerAbove; layer <= maxLayer; layer += 1 {
				fmt.Println("slicing layer ", layer)
				layerSegments, ok := m[layer]
				if !ok {
					layerSegments = make(map[Point]Segment)
					m[layer] = layerSegments
				}
				segment := SliceAngle(origin, right, left, &facet.Normal, float32(layer)*float32(layerHeight))
				fmt.Println("obtained segment: ", segment)
				if segment.Start[0] == segment.End[0] && segment.Start[1] == segment.End[1] {
					fmt.Println("degenerate segment: skipping")
				} else {
					layerSegments[segment.Start] = segment
					segments[layer] = append(segments[layer], segment)
				}
			}
		}
	}

	return &m, &segments, globalMinLayer, globalMaxLayer
}

func PathsFromSegments(segments map[Point]Segment) *[]Path {
	var paths []Path
	for point, segment := range segments {
		fmt.Println(segment)
		if segment.Visited {
			continue
		}
		start := point
		segment.Visited = true
		s := segment
		var path Path
		path = append(path, segment.Start)
		for {
			s, ok := segments[s.End]
			if !ok {
				fmt.Println("path doesn't close!")
				break
			}
			s.Visited = true
			path = append(path, s.Start)
			if s.End == start {
				// path closed, so we're done with it
				break
			}
		}
		paths = append(paths, path)
	}
	return &paths
}
