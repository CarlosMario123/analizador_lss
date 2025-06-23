package lexer
import (
	"fmt"
)

// Lexer estructura para el analizador léxico
type Lexer struct {
	input        string
	position     int  // posición actual en el input
	readPosition int  // siguiente posición a leer
	ch           byte // caracter actual
	line         int  // línea actual
	column       int  // columna actual
	tokens       []Token // tokens encontrados
}

// New crea un nuevo analizador léxico
func New(input string) *Lexer {
	l := &Lexer{
		input:    input,
		line:     1,
		column:   0,
		tokens:   []Token{},
	}
	l.readChar() // Lee el primer caracter
	return l
}

// Analizar realiza el análisis léxico completo y devuelve todos los tokens
func (l *Lexer) Analizar() []Token {
	for {
		tok := l.NextToken()
		l.tokens = append(l.tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return l.tokens
}

// NextToken obtiene el siguiente token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	// Añadir depuración
	fmt.Printf("Analizando caracter: '%c' en posición %d, línea %d, columna %d\n", 
		l.ch, l.position, l.line, l.column)

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TOKEN_EQUAL, Lexeme: string(ch) + string(l.ch)}
		} else {
			tok = Token{Type: TOKEN_ASSIGN, Lexeme: string(l.ch)}
		}
	case '+':
		tok = Token{Type: TOKEN_PLUS, Lexeme: string(l.ch)}
	case '*':
		tok = Token{Type: TOKEN_MULT, Lexeme: string(l.ch)}
	case ';':
		tok = Token{Type: TOKEN_SEMI, Lexeme: string(l.ch)}
	case '{':
		tok = Token{Type: TOKEN_LBRACE, Lexeme: string(l.ch)}
	case '}':
		tok = Token{Type: TOKEN_RBRACE, Lexeme: string(l.ch)}
	case '(':
		tok = Token{Type: TOKEN_LPAREN, Lexeme: string(l.ch)}
	case ')':
		tok = Token{Type: TOKEN_RPAREN, Lexeme: string(l.ch)}
	case 0:
		tok = Token{Type: TOKEN_EOF, Lexeme: ""}
	default:
		if isLetter(l.ch) {
			tok.Lexeme = l.readIdentifier()
			// Verifica si es una palabra reservada
			if tokenType, ok := MapaReservadas[tok.Lexeme]; ok {
				tok.Type = tokenType
			} else {
				tok.Type = TOKEN_IDENT
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TOKEN_NUMBER
			tok.Lexeme = l.readNumber()
			return tok
		} else {
			tok = Token{Type: TOKEN_ERROR, Lexeme: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}

// readChar lee el siguiente caracter y avanza la posición
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII 'NUL'
	} else {
		l.ch = l.input[l.readPosition]
	}
	
	l.position = l.readPosition
	l.readPosition++
	l.column++
	
	// Actualiza línea y columna en caso de nueva línea
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar mira el siguiente caracter sin avanzar
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// readIdentifier lee un identificador o palabra reservada
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber lee un número
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipWhitespace ignora espacios en blanco
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter verifica si un caracter es una letra o guión bajo
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit verifica si un caracter es un dígito
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// ContarTokens cuenta el número de tokens por categoría
func (l *Lexer) ContarTokens() map[string]int {
	conteo := map[string]int{
		"PR":      0,
		"ID":      0,
		"Numeros": 0,
		"Simbolos": 0,
		"Error":   0,
	}

	for _, token := range l.tokens {
		categoria := token.Categoria()
		if _, ok := conteo[categoria]; ok {
			conteo[categoria]++
		}
	}

	return conteo
}