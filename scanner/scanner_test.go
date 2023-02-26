package scanner

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

var (
	bookText string = `@BOOK{Knuth1997,
  title     = "The Art of Computer Programming",
  author    = "Knuth, Donald Ervin",
  publisher = "Addison Wesley",
  address   = "Boston, MA",
  edition   = "3.",
  year      = 1997
}`
	collectionText string = `@COLLECTION{yanagida1975,
  editor    = {柳田聖山},
  title     = {禪學叢書},
  location  = {京都},
  publisher = {中文出版社},
  date      = 1975
}`
)

// See if the string representation of the Token meets the predicted format.
func TestTokenString(t *testing.T) {
	cases := []struct {
		Rune  rune
		Start int
		End   int
	}{
		{'\u0000', 0, 1},
		{'a', 23, 24},
		{'禪', 2, 5},
		{'9', 37, 38},
		{'\uffff', 0, 3},
	}
	for _, c := range cases {
		t.Run(string(c.Rune), func(t *testing.T) {
			tn := Token{
				State{Rune: c.Rune, Start: c.Start, End: c.End}, nil,
			}
			have := fmt.Sprintf("%s", tn)
			want := fmt.Sprintf("[%c %d:%d]", c.Rune, c.Start, c.End)
			if have != want {
				t.Errorf("have %s; want %s", have, want)
			}
		})
	}
}

// Test if the set of tokens returned from the Scanner is aligned with the
// expected output.
func TestScannerScan(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"book", bookText},
		{"collection", collectionText},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := New(strings.NewReader(c.input))
			if err != nil {
				t.Fatal("failed to initialize the scanner")
			}
			result := []byte{}
			for s.Scan() {
				t := s.Token()
				result = utf8.AppendRune(result, t.Rune)
			}
			if string(result) != c.input {
				t.Error("the result and the expected text are not aligned")
			}
		})
	}
}

// Check if the scanner Scan method returns false upon failure.
func TestScanFailure(t *testing.T) {
	s := Scanner{State{'\u0000', 0, 0}, []byte(string('\uFFFD'))}
	if s.Scan() != false {
		t.Error("scan was expected to return false on rune error")
	}
}

// See if the Goto method of the Scanner changes the recorded state
// accordingly.
func TestScannerGoto(t *testing.T) {
	cases := []struct {
		name  string
		token Token
	}{
		{`ù`, Token{State{'ù', 5, 7}, nil}},
		{`æ`, Token{State{'æ', 112, 115}, nil}},
		{`ß`, Token{State{'ß', 0, 2}, nil}},
		{`§`, Token{State{'§', 14, 16}, nil}},
		{`®`, Token{State{'®', 67, 70}, nil}},
	}
	s := Scanner{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s.Goto(c.token)
			if c.token.Pos() != s.Pos() {
				t.Errorf("have %v; want %v", c.token.Pos(), s.Pos())
			}
		})
	}
}

// Check if the Scanner assumes the expected state upon being reset.
func TestScannerReset(t *testing.T) {
	s := Scanner{}
	s.Reset()
	have, want := s.Pos(), State{'\u0000', 0, 0}
	if have != want {
		t.Errorf("have %v; want %v", have, want)
	}
}

type failing struct{}

func (failing) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failed")
}

// Test if the Scanner initialization fails when byte buffer cannot be read.
func TestNewScannerErrors(t *testing.T) {
	cases := []struct {
		name   string
		reader io.Reader
	}{
		{"nil", nil},
		{"failing", failing{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := New(c.reader)
			if err == nil {
				t.Errorf("expected to get an error here")
			}
		})
	}
}
