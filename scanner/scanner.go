package scanner

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var ErrNilReader error = errors.New("provided nil reader")

// State records the UTF-8 encoded Unicode character alongside its position in
// the byte buffer it's been read from.
type State struct {
	Rune  rune
	Start int
	End   int
}

// Token represents a single, self-contained UTF-8 encoded Unicode character
// read from a byte buffer.
type Token struct {
	State
	Buffer *[]byte
}

// Scanner is responsible for scanning the contents of a text file. The
// structure is stateful and is considered unsafe to use in multithreaded
// programs.
type Scanner struct {
	State
	Buffer []byte
}

// NewScanner creates a new instance of the Scanner initialized at its zero
// state.
func NewScanner(r io.Reader) (*Scanner, error) {
	if r == nil {
		return nil, ErrNilReader
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	rd := Scanner{
		State{Rune: '\u0000', Start: 0, End: 0},
		buf,
	}
	return &rd, nil
}

// Pos provides the position of the recorded character in the byte buffer.
func (s State) Pos() State {
	return State{s.Rune, s.Start, s.End}
}

// String returns a text representation of the Token.
func (t Token) String() string {
	repr := fmt.Sprintf("[%c %d:%d]", t.Rune, t.Start, t.End)
	return repr
}

// Reset puts the Scanner back to its initial state with the cursor pointing at
// the very start of the buffer.
func (s *Scanner) Reset() {
	s.Rune, s.Start, s.End = '\u0000', 0, 0
}

// Goto moves the cursor of the Scanner to the position specified by the
// attributes of the Token.
func (s *Scanner) Goto(t Token) {
	s.Rune, s.Start, s.End = t.Rune, t.Start, t.End
}

// Token returns the Token currently pointed at by the cursor.
func (s *Scanner) Token() Token {
	token := Token{
		State{Rune: s.Rune, Start: s.Start, End: s.End},
		&s.Buffer,
	}
	return token
}

// Scan advances the cursor of the Scanner by a single UTF-8 encoded Unicode
// character. The method returns a boolean value so that is can be used
// idiomatically in a for-loop in the fashion as other Go scanners are used.
func (s *Scanner) Scan() bool {
	if s.End >= len(s.Buffer) {
		return false
	}
	r, size := rune(s.Buffer[s.End]), 1
	if r >= utf8.RuneSelf {
		r, size = utf8.DecodeRune(s.Buffer[s.End:])
		if r == utf8.RuneError {
			return false
		}
	}
	s.Rune, s.Start, s.End = r, s.End, s.End+size
	return true
}
