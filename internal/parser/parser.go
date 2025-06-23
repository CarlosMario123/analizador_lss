package parser

import (
	"fmt"
	"analyzer-api/internal/lexer"
)

// Parser estructura para el analizador sintáctico
type Parser struct {
	l         *lexer.Lexer
	tokens    []lexer.Token
	position  int
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []ErrorSintactico
}

// New crea un nuevo analizador sintáctico
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		tokens: l.Analizar(),
		errors: []ErrorSintactico{},
	}

	// Leer dos tokens para inicializar curToken y peekToken
	if len(p.tokens) > 0 {
		p.curToken = p.tokens[0]
	}
	if len(p.tokens) > 1 {
		p.peekToken = p.tokens[1]
	}

	return p
}

// Errores devuelve los errores de sintaxis encontrados
func (p *Parser) Errores() []ErrorSintactico {
	return p.errors
}

// nextToken avanza al siguiente token
func (p *Parser) nextToken() {
	prevToken := p.curToken
	p.position++
	if p.position < len(p.tokens) {
		p.curToken = p.tokens[p.position]
	} else {
		p.curToken = lexer.Token{Type: lexer.TOKEN_EOF, Lexeme: "EOF"}
	}
	if p.position+1 < len(p.tokens) {
		p.peekToken = p.tokens[p.position+1]
	} else {
		p.peekToken = lexer.Token{Type: lexer.TOKEN_EOF, Lexeme: "EOF"}
	}
	
	fmt.Printf("Avanzando de '%s' a '%s', siguiente: '%s'\n", 
		prevToken.Lexeme, p.curToken.Lexeme, p.peekToken.Lexeme)
}

// Parse analiza el programa completo
func (p *Parser) Parse() *Programa {
	programa := &Programa{
		Declaraciones: []Declaracion{},
	}

	fmt.Println("Iniciando análisis sintáctico...")

	// Contador de seguridad para evitar bucles infinitos
	maxIteraciones := len(p.tokens) * 2 // Máximo 2 veces el número de tokens
	iteraciones := 0

	for p.curToken.Type != lexer.TOKEN_EOF && iteraciones < maxIteraciones {
		iteraciones++
		fmt.Printf("Procesando token: '%s' (%s) en línea %d, columna %d [iter: %d]\n", 
			p.curToken.Lexeme, p.curToken.Type, p.curToken.Line, p.curToken.Column, iteraciones)
		
		// Guardar posición actual para detectar bucles infinitos
		posicionAnterior := p.position
		
		decl := p.parseDeclaracion()
		if decl != nil {
			programa.Declaraciones = append(programa.Declaraciones, decl)
			fmt.Printf("Declaración agregada exitosamente\n")
		}
		
		// CRÍTICO: Si no avanzamos, forzar avance para evitar bucle infinito
		if p.position == posicionAnterior && p.curToken.Type != lexer.TOKEN_EOF {
			fmt.Printf("RECOVERY: Forzando avance del token '%s' para evitar bucle infinito\n", p.curToken.Lexeme)
			p.nextToken()
		}
	}

	if iteraciones >= maxIteraciones {
		p.agregarError("Análisis interrumpido: demasiadas iteraciones (posible bucle infinito)")
		fmt.Println("ERROR: Análisis interrumpido por exceso de iteraciones - posible bucle infinito")
	}

	fmt.Println("Análisis sintáctico completado.")
	return programa
}

// parseDeclaracion analiza una declaración
func (p *Parser) parseDeclaracion() Declaracion {
	switch p.curToken.Lexeme {
	case "int":
		return p.parseDeclaracionVariable()
	case "do":
		return p.parseDoWhile()
	default:
		if p.curToken.Type == lexer.TOKEN_IDENT && p.peekTokenIs(lexer.TOKEN_ASSIGN) {
			return p.parseDeclaracionAsignacion()
		} else {
			p.agregarError(fmt.Sprintf("Declaración inesperada con token %s", p.curToken.Lexeme))
			fmt.Printf("ERROR: Token inesperado '%s' - saltando\n", p.curToken.Lexeme)
			// SIEMPRE avanzar en caso de error para evitar bucle infinito
			p.nextToken()
			return nil
		}
	}
}

// parseDeclaracionVariable analiza una declaración de variable (int a = 0;)
func (p *Parser) parseDeclaracionVariable() *DeclaracionVariable {
	fmt.Printf("Analizando declaración de variable, tipo: '%s'\n", p.curToken.Lexeme)
	
	decl := &DeclaracionVariable{Tipo: p.curToken.Lexeme}

	// Siguiente debe ser identificador
	if !p.expectPeek(lexer.TOKEN_IDENT) {
		return nil
	}
	
	decl.Nombre = p.curToken.Lexeme
	fmt.Printf("Nombre de la variable: '%s'\n", decl.Nombre)

	// Siguiente debe ser =
	if !p.expectPeek(lexer.TOKEN_ASSIGN) {
		return nil
	}

	// Siguiente debe ser el valor
	if !p.nextTokenExists() {
		p.agregarError("Se esperaba un valor después de '='")
		return nil
	}
	p.nextToken()
	
	fmt.Printf("Analizando valor de la variable, token: '%s'\n", p.curToken.Lexeme)
	decl.Valor = p.parseExpresionSimple()
	if decl.Valor == nil {
		return nil
	}

	// Debe terminar con ;
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		return nil
	}

	// Avanzar después del ;
	p.nextToken()
	
	fmt.Println("Declaración de variable analizada correctamente")
	return decl
}

// parseDoWhile analiza una estructura do-while
func (p *Parser) parseDoWhile() *DeclaracionDoWhile {
	fmt.Println("Iniciando análisis de do-while")
	dowhile := &DeclaracionDoWhile{
		Cuerpo: []Declaracion{},
	}

	// Siguiente debe ser {
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		p.agregarError("Se esperaba '{' después de 'do'")
		fmt.Println("ERROR: Falta '{' después de 'do' - abortando do-while")
		
		// RECOVERY: Buscar hasta encontrar 'while' o final
		for p.curToken.Type != lexer.TOKEN_WHILE && p.curToken.Type != lexer.TOKEN_EOF {
			fmt.Printf("RECOVERY: Saltando token '%s'\n", p.curToken.Lexeme)
			p.nextToken()
		}
		
		if p.curToken.Type == lexer.TOKEN_EOF {
			fmt.Println("RECOVERY: Llegó al final sin encontrar 'while'")
			return nil
		}
		
		fmt.Println("RECOVERY: Encontrado 'while', intentando continuar análisis...")
		// Continuar con el análisis del while desde aquí
		goto parseWhileCondition
	}

	// Avanzar después de {
	p.nextToken()

	// Analizar el cuerpo del do-while hasta encontrar }
	for p.curToken.Type != lexer.TOKEN_RBRACE && p.curToken.Type != lexer.TOKEN_EOF {
		fmt.Printf("Analizando token en cuerpo do-while: '%s' (%s)\n", p.curToken.Lexeme, p.curToken.Type)
		
		if p.curToken.Type == lexer.TOKEN_IDENT && p.peekTokenIs(lexer.TOKEN_ASSIGN) {
			fmt.Printf("Encontrada asignación a '%s' dentro del do-while\n", p.curToken.Lexeme)
			asignacion := p.parseDeclaracionAsignacion()
			if asignacion != nil {
				dowhile.Cuerpo = append(dowhile.Cuerpo, asignacion)
				fmt.Println("Asignación agregada al cuerpo del do-while")
			}
		} else {
			p.agregarError(fmt.Sprintf("Se esperaba una asignación en el cuerpo del do-while, se encontró '%s'", p.curToken.Lexeme))
			p.nextToken()
		}
	}

	if p.curToken.Type != lexer.TOKEN_RBRACE {
		p.agregarError("Se esperaba '}' al final del cuerpo do-while")
		return nil
	}

	// Siguiente debe ser while
	if !p.expectPeek(lexer.TOKEN_WHILE) {
		return nil
	}

parseWhileCondition:
	// Siguiente debe ser (
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		p.agregarError("Se esperaba '(' después de 'while'")
		return nil
	}

	// Siguiente debe ser la condición
	if !p.nextTokenExists() {
		p.agregarError("Se esperaba una condición después de '('")
		return nil
	}
	p.nextToken()
	
	fmt.Printf("Analizando condición del while, token actual: '%s'\n", p.curToken.Lexeme)
	dowhile.Condicion = p.parseExpresionComparacion()
	if dowhile.Condicion == nil {
		// RECOVERY: Si falla la condición, buscar hasta ) y ;
		fmt.Println("RECOVERY: Error en condición, buscando ')' y ';'")
		for p.curToken.Type != lexer.TOKEN_RPAREN && p.curToken.Type != lexer.TOKEN_EOF {
			p.nextToken()
		}
		if p.curToken.Type == lexer.TOKEN_RPAREN {
			p.nextToken()
			if p.curToken.Type == lexer.TOKEN_SEMI {
				p.nextToken()
			}
		}
		return nil
	}

	// Siguiente debe ser )
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Siguiente debe ser ;
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		return nil
	}

	// Avanzar después del ;
	p.nextToken()

	fmt.Println("Análisis de do-while completado con éxito")
	return dowhile
}

// parseDeclaracionAsignacion analiza una asignación (a = 5;)
func (p *Parser) parseDeclaracionAsignacion() Declaracion {
	fmt.Printf("Parseando asignación para variable: '%s'\n", p.curToken.Lexeme)
	
	nombre := p.curToken.Lexeme
	
	// Siguiente debe ser =
	if !p.expectPeek(lexer.TOKEN_ASSIGN) {
		return nil
	}

	// Siguiente debe ser el valor
	if !p.nextTokenExists() {
		p.agregarError("Se esperaba un valor después de '='")
		return nil
	}
	p.nextToken()
	
	valor := p.parseExpresionCompleta()
	if valor == nil {
		return nil
	}

	// Debe terminar con ;
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		return nil
	}

	// Avanzar después del ;
	p.nextToken()

	return &DeclaracionAsignacion{
		Asignacion: &ExpresionAsignacion{
			Nombre: nombre,
			Valor:  valor,
		},
	}
}

// parseExpresionSimple analiza una expresión simple (número o identificador sin operadores)
func (p *Parser) parseExpresionSimple() Expresion {
	fmt.Printf("Analizando expresión simple, token actual: '%s'\n", p.curToken.Lexeme)

	switch p.curToken.Type {
	case lexer.TOKEN_IDENT:
		fmt.Printf("Expresión simple es identificador: '%s'\n", p.curToken.Lexeme)
		return &ExpresionIdentificador{Valor: p.curToken.Lexeme}
	case lexer.TOKEN_NUMBER:
		fmt.Printf("Expresión simple es número: '%s'\n", p.curToken.Lexeme)
		return &ExpresionNumero{Valor: p.curToken.Lexeme}
	default:
		p.agregarError(fmt.Sprintf("Token inesperado en expresión simple: %s", p.curToken.Lexeme))
		return nil
	}
}

// parseExpresionCompleta analiza una expresión que puede incluir operadores aritméticos
func (p *Parser) parseExpresionCompleta() Expresion {
	fmt.Printf("Analizando expresión completa, token actual: '%s'\n", p.curToken.Lexeme)

	// Obtener el lado izquierdo
	izquierda := p.parseExpresionSimple()
	if izquierda == nil {
		return nil
	}

	// Verificar si hay un operador aritmético (* o +)
	if p.peekTokenIs(lexer.TOKEN_MULT) || p.peekTokenIs(lexer.TOKEN_PLUS) {
		fmt.Printf("Encontrado operador aritmético: '%s'\n", p.peekToken.Lexeme)
		
		p.nextToken() // Avanzar al operador
		operador := p.curToken.Lexeme
		
		if !p.nextTokenExists() {
			p.agregarError("Se esperaba un operando después del operador")
			return nil
		}
		p.nextToken() // Avanzar al operando derecho
		
		fmt.Printf("Analizando lado derecho, token: '%s'\n", p.curToken.Lexeme)
		derecha := p.parseExpresionSimple()
		if derecha == nil {
			return nil
		}
		
		return &ExpresionBinaria{
			Izquierda: izquierda,
			Operador:  operador,
			Derecha:   derecha,
		}
	}

	return izquierda
}

// parseExpresionComparacion analiza una expresión de comparación (x == 2)
func (p *Parser) parseExpresionComparacion() Expresion {
	fmt.Printf("Analizando expresión de comparación, token actual: '%s'\n", p.curToken.Lexeme)

	// Obtener el lado izquierdo
	izquierda := p.parseExpresionSimple()
	if izquierda == nil {
		return nil
	}

	// Debe haber un operador de comparación
	if p.peekTokenIs(lexer.TOKEN_EQUAL) {
		fmt.Printf("Encontrado operador de comparación: '%s'\n", p.peekToken.Lexeme)
		
		p.nextToken() // Avanzar al operador
		operador := p.curToken.Lexeme
		
		if !p.nextTokenExists() {
			p.agregarError("Se esperaba un operando después del operador de comparación")
			return nil
		}
		p.nextToken() // Avanzar al operando derecho
		
		fmt.Printf("Analizando lado derecho de comparación, token: '%s'\n", p.curToken.Lexeme)
		derecha := p.parseExpresionSimple()
		if derecha == nil {
			return nil
		}
		
		return &ExpresionBinaria{
			Izquierda: izquierda,
			Operador:  operador,
			Derecha:   derecha,
		}
	}

	p.agregarError("Se esperaba un operador de comparación")
	return nil
}

// nextTokenExists verifica si hay un siguiente token válido
func (p *Parser) nextTokenExists() bool {
	return p.position+1 < len(p.tokens) && p.peekToken.Type != lexer.TOKEN_EOF
}

// expectPeek verifica si el siguiente token es del tipo esperado
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	fmt.Printf("Verificando si el siguiente token es '%s', actual: '%s', siguiente: '%s'\n", 
		t, p.curToken.Type, p.peekToken.Type)
	
	if p.peekTokenIs(t) {
		p.nextToken()
		fmt.Printf("Verificación exitosa, token actual ahora es: '%s'\n", p.curToken.Lexeme)
		return true
	}
	
	p.peekError(t)
	fmt.Printf("Verificación fallida, se esperaba '%s', se encontró '%s'\n", 
		t, p.peekToken.Type)
	return false
}

// peekTokenIs verifica si el siguiente token es del tipo dado
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// curTokenIs verifica si el token actual es del tipo dado
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekError agrega un error cuando el siguiente token no es el esperado
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("esperaba siguiente token %s, obtuvo %s",
		t, p.peekToken.Type)
	p.agregarError(msg)
}

// agregarError agrega un error con la línea y columna actual
func (p *Parser) agregarError(mensaje string) {
	error := ErrorSintactico{
		Mensaje: mensaje,
		Linea:   p.curToken.Line,
		Columna: p.curToken.Column,
	}
	p.errors = append(p.errors, error)
}