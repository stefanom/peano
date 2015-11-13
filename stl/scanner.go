package stl

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	// If we see a digit then consume as a number.

	// NOTE: while 'e' can appear both in tokens as well as floating point
	// numbers, it will never be the first letter in a floating point
	// number so the order of the matching makes it work.
	// Yeah, I know, this is somewhat hacky and an ill-defined STL file will
	// make this barf, but I don't care right now.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanToken()
	} else if isNumber(ch) {
		s.unread()
		return s.scanNumber()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	}

	return ILLEGAL, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanToken() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent token character into the buffer.
	// Non-letter characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch strings.ToLower(buf.String()) {
	case "solid":
		return SOLID, buf.String()
	case "facet":
		return FACET, buf.String()
	case "normal":
		return NORMAL, buf.String()
	case "outer":
		return OUTER, buf.String()
	case "loop":
		return LOOP, buf.String()
	case "vertex":
		return VERTEX, buf.String()
	case "endloop":
		return ENDLOOP, buf.String()
	case "endfacet":
		return ENDFACET, buf.String()
	case "endsolid":
		return ENDSOLID, buf.String()
	}

	// Otherwise return as a regular identifier.
	return ILLEGAL, buf.String()
}

func (s *Scanner) scanNumber() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent token character into the buffer.
	// Non-digit characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isNumber(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return NUMBER, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_')
}

// isNumber returns true if the rune can be part of a floating point number.
func isNumber(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch == '-') || (ch == '+') || (ch == 'e')
}

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
