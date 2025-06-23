package lexer

import "fmt"

// TokenType representa el tipo de token
type TokenType string

// Definición de tipos de tokens
const (
	// Palabras reservadas
	TOKEN_INT    TokenType = "PR"     // int
	TOKEN_DO     TokenType = "PR"     // do
	TOKEN_WHILE  TokenType = "PR"     // while
	
	// Identificadores
	TOKEN_IDENT  TokenType = "ID"     // identificadores (a, b, c, x, etc.)
	
	// Literales
	TOKEN_NUMBER TokenType = "Numeros" // números (0, 10, 2, 3, etc.)
	
	// Operadores y símbolos
	TOKEN_ASSIGN TokenType = "Simbolos" // =
	TOKEN_PLUS   TokenType = "Simbolos" // +
	TOKEN_MULT   TokenType = "Simbolos" // *
	TOKEN_EQUAL  TokenType = "Simbolos" // ==
	TOKEN_SEMI   TokenType = "Simbolos" // ;
	TOKEN_LBRACE TokenType = "Simbolos" // {
	TOKEN_RBRACE TokenType = "Simbolos" // }
	TOKEN_LPAREN TokenType = "Simbolos" // (
	TOKEN_RPAREN TokenType = "Simbolos" // )
	
	// Especiales
	TOKEN_EOF    TokenType = "EOF"     // Fin de archivo
	TOKEN_ERROR  TokenType = "Error"   // Error
)

// Token representa un token individual identificado por el analizador léxico
type Token struct {
	Type    TokenType // Tipo del token (PR, ID, Numeros, Simbolos, Error)
	Lexeme  string    // El texto literal del token
	Line    int       // Línea donde se encontró el token
	Column  int       // Columna donde se encontró el token
}

// MapaReservadas mapea palabras reservadas a sus respectivos tipos de token
var MapaReservadas = map[string]TokenType{
	"int":   TOKEN_INT,
	"do":    TOKEN_DO,
	"while": TOKEN_WHILE,
}

// Categoría retorna la categoría general del token (PR, ID, Numeros, Simbolos, Error)
func (t *Token) Categoria() string {
	return string(t.Type)
}

// String implementa la interfaz Stringer para facilitar la depuración
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Lexeme: '%s', Line: %d, Column: %d}", 
		t.Type, t.Lexeme, t.Line, t.Column)
}