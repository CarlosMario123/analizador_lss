package models

import (
	"analyzer-api/internal/lexer"
	"analyzer-api/internal/parser"
	"analyzer-api/internal/semantic"
)

// TokenInfo representa la información de un token para la API
type TokenInfo struct {
	Tipo     string `json:"tipo"`
	Lexema   string `json:"lexema"`
	Linea    int    `json:"linea"`
	Columna  int    `json:"columna"`
	Categoria string `json:"categoria"`
}

// ConteoTokens representa el conteo de tokens por categoría
type ConteoTokens struct {
	PR       int `json:"pr"`
	ID       int `json:"id"`
	Numeros  int `json:"numeros"`
	Simbolos int `json:"simbolos"`
	Error    int `json:"error"`
	Total    int `json:"total"`
}

// ErrorInfo representa la información de un error para la API
type ErrorInfo struct {
	Tipo     string `json:"tipo"` // "sintactico" o "semantico"
	Mensaje  string `json:"mensaje"`
	Linea    int    `json:"linea"`
	Columna  int    `json:"columna"`
}

// SimboloInfo representa la información de un símbolo para la API
type SimboloInfo struct {
	Nombre   string      `json:"nombre"`
	Tipo     string      `json:"tipo"`
	Valor    interface{} `json:"valor"`
	Linea    int         `json:"linea"`
	Columna  int         `json:"columna"`
}

// ResultadoAnalisis representa el resultado completo del análisis
type ResultadoAnalisis struct {
	Tokens       []TokenInfo   `json:"tokens"`
	ConteoTokens ConteoTokens  `json:"conteoTokens"`
	Errores      []ErrorInfo   `json:"errores"`
	Simbolos     []SimboloInfo `json:"simbolos"`
	CodigoFuente string        `json:"codigoFuente"`
}

// NuevoResultadoAnalisis crea un nuevo resultado de análisis a partir de los componentes
func NuevoResultadoAnalisis(
	lex *lexer.Lexer,
	p *parser.Parser, 
	sem *semantic.Analizador,
	codigoFuente string,
) *ResultadoAnalisis {
	resultado := &ResultadoAnalisis{
		Tokens:       []TokenInfo{},
		Errores:      []ErrorInfo{},
		Simbolos:     []SimboloInfo{},
		CodigoFuente: codigoFuente,
	}

	// Convertir tokens a TokenInfo
	tokens := lex.Analizar()
	for _, token := range tokens {
		resultado.Tokens = append(resultado.Tokens, TokenInfo{
			Tipo:     string(token.Type),
			Lexema:   token.Lexeme,
			Linea:    token.Line,
			Columna:  token.Column,
			Categoria: token.Categoria(),
		})
	}

	// Obtener conteo de tokens
	conteo := lex.ContarTokens()
	resultado.ConteoTokens = ConteoTokens{
		PR:       conteo["PR"],
		ID:       conteo["ID"],
		Numeros:  conteo["Numeros"],
		Simbolos: conteo["Simbolos"],
		Error:    conteo["Error"],
		Total:    len(tokens) - 1, // Restamos 1 para no contar EOF
	}

	// Convertir errores sintácticos a ErrorInfo
	for _, err := range p.Errores() {
		resultado.Errores = append(resultado.Errores, ErrorInfo{
			Tipo:    "sintactico",
			Mensaje: err.Mensaje,
			Linea:   err.Linea,
			Columna: err.Columna,
		})
	}

	// Convertir errores semánticos a ErrorInfo
	for _, err := range sem.Analizar() {
		resultado.Errores = append(resultado.Errores, ErrorInfo{
			Tipo:    "semantico",
			Mensaje: err.Mensaje,
			Linea:   err.Linea,
			Columna: err.Columna,
		})
	}

	// Convertir símbolos a SimboloInfo
	for _, simbolo := range sem.TablaSimbolos().ListarSimbolos() {
		resultado.Simbolos = append(resultado.Simbolos, SimboloInfo{
			Nombre:  simbolo.Nombre,
			Tipo:    simbolo.Tipo,
			Valor:   simbolo.Valor,
			Linea:   simbolo.Linea,
			Columna: simbolo.Columna,
		})
	}

	return resultado
}