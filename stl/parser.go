package stl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/stefanom/peano/geom"
	"io"
	"io/ioutil"
	"strconv"
)

type Model struct {
	Header [80]byte
	Length int32
	Facets []geom.Facet
}

type Parser struct {
	r   io.Reader
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

// Parse the STL file into a Model
func (p *Parser) Parse() (*Model, error) {
	data, err := ioutil.ReadAll(p.r)
	if err != nil {
		panic(err)
	}

	m := new(Model)

	err = binary.Read(bytes.NewBuffer(data[0:80]), binary.LittleEndian, &m.Header)
	if err != nil {
		return m, err
	}

	start := string(m.Header[:6])
	if start == "solid " {
		p.s = NewScanner(bytes.NewReader(data))
		m.Facets = make([]geom.Facet, 0)

		// First token should be the "solid" keyword.
		if tok, lit := p.scanIgnoreWhitespace(); tok != SOLID {
			return nil, fmt.Errorf("found %q, expected 'solid'", lit)
		}

		// Next is the name of the solid, which we ignore.
		p.scanIgnoreWhitespace()

		// Then we loop over the facets.
		for {
			// Read a field.
			tok, lit := p.scanIgnoreWhitespace()
			if tok != FACET && tok != ENDSOLID {
				return nil, fmt.Errorf("found %q, expected 'facet' or 'endsolid'", lit)
			}

			if tok == ENDSOLID {
				p.unscan()
				break
			}

			facet := new(geom.Facet)

			// Now we read the facet normal.
			if tok, lit := p.scanIgnoreWhitespace(); tok != NORMAL {
				return nil, fmt.Errorf("found %q, expected 'normal'", lit)
			}

			normal := new(geom.Vector)
			for j := 0; j < 3; j++ {
				tok, lit := p.scanIgnoreWhitespace()
				if tok != NUMBER {
					return nil, fmt.Errorf("found %q, expected number", lit)
				}
				coordinate, err := strconv.ParseFloat(lit, 32)
				if err != nil {
					return nil, err
				}
				normal[j] = float32(coordinate)
			}
			facet.Normal = *normal

			// Now we read the facet vertices.
			if tok, lit := p.scanIgnoreWhitespace(); tok != OUTER {
				return nil, fmt.Errorf("found %q, expected 'outer'", lit)
			}
			if tok, lit := p.scanIgnoreWhitespace(); tok != LOOP {
				return nil, fmt.Errorf("found %q, expected 'loop'", lit)
			}

			vectors := make([]geom.Vector, 3)
			for i := 0; i < 3; i++ {
				if tok, lit := p.scanIgnoreWhitespace(); tok != VERTEX {
					return nil, fmt.Errorf("found %q, expected 'vertex'", lit)
				}
				for j := 0; j < 3; j++ {
					tok, lit := p.scanIgnoreWhitespace()
					if tok != NUMBER {
						return nil, fmt.Errorf("found %q, expected number", lit)
					}
					coordinate, err := strconv.ParseFloat(lit, 32)
					if err != nil {
						return nil, err
					}
					vectors[i][j] = float32(coordinate)
				}
			}
			facet.Vertex1 = vectors[0]
			facet.Vertex2 = vectors[1]
			facet.Vertex3 = vectors[2]

			if tok, lit := p.scanIgnoreWhitespace(); tok != ENDLOOP {
				return nil, fmt.Errorf("found %q, expected 'endloop'", lit)
			}

			if tok, lit := p.scanIgnoreWhitespace(); tok != ENDFACET {
				return nil, fmt.Errorf("found %q, expected 'endfacet'", lit)
			}

			m.Facets = append(m.Facets, *facet)
		}

		// Next we should see the "FROM" keyword.
		if tok, lit := p.scanIgnoreWhitespace(); tok != ENDSOLID {
			return nil, fmt.Errorf("found %q, expected 'endsolid'", lit)
		}

		// The very end is the solid name but we can ignore that.

		// Make sure the model length reflects the facets found.
		m.Length = int32(len(m.Facets))

		// Return the successfully parsed model.
		return m, nil

	} else {
		// Obtain the number of facets this model contains.
		err = binary.Read(bytes.NewBuffer(data[80:84]), binary.LittleEndian, &m.Length)
		if err != nil {
			return m, err
		}

		// Create the slice of Facets.
		m.Facets = make([]geom.Facet, m.Length)

		// Read the slice of Facets directly from their binary reprsentation.
		err = binary.Read(bytes.NewBuffer(data[84:]), binary.LittleEndian, &m.Facets)
		if err != nil {
			return m, err
		}
	}

	return m, nil
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
