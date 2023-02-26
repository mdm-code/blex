package lexer

import (
	"errors"

	"github.com/mdm-code/blex/scanner"
)

const (
	// Control-flow tokens.
	ERR TokenType = iota
	EOF

	// Single-character tokens.
	AT     // @
	PERC   // %
	LBRACE // {
	RBRACE // }
	LPAREN // (
	RPAREN // )
	EQUALS // =
	COMMA  // ,
	HASH   // #
	QUOTE  // "

	// Literal tokens.
	IDENT
	STRING
	NUMBER
)

const (
	S_NULL state = iota
	S_ERR
	S_EOF
)

var ErrNilScanner error = errors.New("provided nil scanner")

type (
	TokenType uint8
	state     uint8
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	state   state
	scanner *scanner.Scanner
	tokens  chan Token
	states  map[state]func(*Lexer) state
}

func New(s *scanner.Scanner) *Lexer {
	lexer := Lexer{
		state:   S_NULL,
		scanner: s,
		tokens:  make(chan Token, 2),
		states: map[state]func(*Lexer) state{
			S_NULL: (*Lexer).Null,
		},
	}
	return &lexer
}

func (l *Lexer) Null() state {
	for {
		l.tokens <- Token{ERR, ""}
	}
}

func (l *Lexer) Token() Token {
	for {
		select {
		case t := <-l.tokens:
			return t
		default:
			l.state = l.states[l.state](l)
		}
	}
}
