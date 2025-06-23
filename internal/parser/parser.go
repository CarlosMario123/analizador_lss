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
		// Si llegamos al final, establecer un token EOF
		p.curToken = lexer.Token{Type: lexer.TOKEN_EOF, Lexeme: "EOF"}
	}
	if p.position+1 < len(p.tokens) {
		p.peekToken = p.tokens[p.position+1]
	} else {
		// Si no hay siguiente token, establecer un token EOF
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

	// Para depuración
	fmt.Println("Iniciando análisis sintáctico...")

	// Mientras no lleguemos al final del archivo
	for p.curToken.Type != lexer.TOKEN_EOF {
		// Imprimir TODOS los tokens para depuración
		fmt.Printf("Procesando token: '%s' (%s) en línea %d, columna %d\n", 
			p.curToken.Lexeme, p.curToken.Type, p.curToken.Line, p.curToken.Column)
		
		// Procesar según el lexema del token actual
		if p.curToken.Lexeme == "int" {
			// Declaración de variable
			decl := p.parseDeclaracionVariable()
			if decl != nil {
				programa.Declaraciones = append(programa.Declaraciones, decl)
				fmt.Printf("Agregada declaración de variable: %s\n", decl.Nombre)
			} else {
				fmt.Println("Error al analizar declaración de variable")
			}
		} else if p.curToken.Lexeme == "do" {
			// Estructura do-while
			fmt.Println("Encontrado 'do', analizando estructura do-while...")
			decl := p.parseDoWhile()
			if decl != nil {
				programa.Declaraciones = append(programa.Declaraciones, decl)
				fmt.Printf("Agregada estructura do-while con %d sentencias en el cuerpo\n", len(decl.Cuerpo))
			} else {
				fmt.Println("Error al analizar estructura do-while")
			}
		} else if p.curToken.Type == lexer.TOKEN_IDENT {
			// Asignación (identificador seguido de =)
			if p.peekTokenIs(lexer.TOKEN_ASSIGN) {
				fmt.Printf("Encontrada asignación a '%s'\n", p.curToken.Lexeme)
				decl := p.parseDeclaracionAsignacion()
				if decl != nil {
					programa.Declaraciones = append(programa.Declaraciones, decl)
					fmt.Println("Agregada asignación")
				} else {
					fmt.Println("Error al analizar asignación")
				}
			} else {
				p.agregarError(fmt.Sprintf("Se esperaba un operador de asignación después del identificador '%s'", p.curToken.Lexeme))
				fmt.Printf("Error: se esperaba '=' después de '%s'\n", p.curToken.Lexeme)
				p.nextToken() // Avanzar para evitar bucle infinito
			}
		} else {
			// Si no es ninguno de los anteriores, reportar error y avanzar
			p.agregarError(fmt.Sprintf("Token inesperado: %s", p.curToken.Lexeme))
			fmt.Printf("Error: token inesperado '%s'\n", p.curToken.Lexeme)
			p.nextToken() // Avanzar para evitar bucle infinito
			continue
		}
		
		// Avanzar al siguiente token (solo si no se avanzó en los métodos de parsing)
		if p.curToken.Type != lexer.TOKEN_EOF {
			p.nextToken()
		}
	}

	fmt.Println("Análisis sintáctico completado.")
	return programa
}

// parseDeclaracion analiza una declaración
func (p *Parser) parseDeclaracion() Declaracion {
	if p.curToken.Type == lexer.TOKEN_INT {
		return p.parseDeclaracionVariable()
	} else if p.curToken.Type == lexer.TOKEN_DO {
		return p.parseDoWhile()
	} else if p.curToken.Type == lexer.TOKEN_IDENT && p.peekToken.Type == lexer.TOKEN_ASSIGN {
		// Si es un identificador seguido de un =, es una asignación
		return p.parseDeclaracionAsignacion()
	} else {
		p.agregarError(fmt.Sprintf("Declaración inesperada con token %s", p.curToken.Lexeme))
		return nil
	}
}

// parseDeclaracionVariable analiza una declaración de variable (int a = 0;)
func (p *Parser) parseDeclaracionVariable() *DeclaracionVariable {
	fmt.Printf("Analizando declaración de variable, tipo: '%s'\n", p.curToken.Lexeme)
	
	decl := &DeclaracionVariable{Tipo: p.curToken.Lexeme}

	// Siguiente token debe ser un identificador
	if !p.expectPeek(lexer.TOKEN_IDENT) {
		fmt.Println("Error: se esperaba un identificador después del tipo")
		return nil
	}
	
	decl.Nombre = p.curToken.Lexeme
	fmt.Printf("Nombre de la variable: '%s'\n", decl.Nombre)

	// Siguiente token debe ser un '='
	if !p.expectPeek(lexer.TOKEN_ASSIGN) {
		fmt.Println("Error: se esperaba '=' después del identificador")
		return nil
	}

	// Avanzar al valor
	p.nextToken()
	fmt.Printf("Analizando valor de la variable, token: '%s'\n", p.curToken.Lexeme)
	
	decl.Valor = p.parseExpresion()
	if decl.Valor == nil {
		fmt.Println("Error: no se pudo analizar el valor de la variable")
		return nil
	}

	// Debe terminar con ';'
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		fmt.Println("Error: se esperaba ';' después del valor")
		return nil
	}

	fmt.Println("Declaración de variable analizada correctamente")
	return decl
}

// parseDoWhile analiza una estructura do-while
func (p *Parser) parseDoWhile() *DeclaracionDoWhile {
	fmt.Println("Iniciando análisis de do-while en posición", p.position)
	dowhile := &DeclaracionDoWhile{
		Cuerpo: []Declaracion{},
	}

	// Siguiente token debe ser '{'
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		fmt.Println("Error: se esperaba '{' después de 'do'")
		return nil
	}

	fmt.Println("Analizando cuerpo del do-while")
	
	// Avanzar después de '{'
	p.nextToken()

	// Analizar el cuerpo del do-while hasta encontrar '}'
	for p.curToken.Type != lexer.TOKEN_RBRACE && p.curToken.Type != lexer.TOKEN_EOF {
		fmt.Printf("Analizando token en cuerpo do-while: '%s' (%s)\n", p.curToken.Lexeme, p.curToken.Type)
		
		if p.curToken.Type == lexer.TOKEN_IDENT {
			if p.peekTokenIs(lexer.TOKEN_ASSIGN) {
				// Es una asignación dentro del do-while
				fmt.Printf("Encontrada asignación a '%s' dentro del do-while\n", p.curToken.Lexeme)
				asignacion := p.parseDeclaracionAsignacion()
				if asignacion != nil {
					dowhile.Cuerpo = append(dowhile.Cuerpo, asignacion)
					fmt.Println("Asignación agregada al cuerpo del do-while")
				} else {
					fmt.Println("Error al analizar asignación en do-while")
				}
			} else {
				p.agregarError(fmt.Sprintf("Se esperaba '=' después de '%s' en el cuerpo del do-while", p.curToken.Lexeme))
				fmt.Printf("Error: se esperaba '=' después de '%s' en do-while\n", p.curToken.Lexeme)
				p.nextToken()
			}
		} else {
			p.agregarError(fmt.Sprintf("Se esperaba un identificador en el cuerpo del do-while, se encontró '%s'", p.curToken.Lexeme))
			fmt.Printf("Error: se esperaba un identificador en do-while, se encontró '%s'\n", p.curToken.Lexeme)
			p.nextToken()
		}
	}

	if p.curToken.Type != lexer.TOKEN_RBRACE {
		p.agregarError("Se esperaba '}' al final del cuerpo do-while")
		fmt.Println("Error: se esperaba '}' al final del cuerpo do-while")
		return nil
	}

	fmt.Println("Fin del cuerpo do-while, buscando 'while'")

	// Siguiente token debe ser 'while'
	if !p.expectPeek(lexer.TOKEN_WHILE) {
		fmt.Println("Error: se esperaba 'while' después de '}'")
		return nil
	}

	// Siguiente token debe ser '('
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		fmt.Println("Error: se esperaba '(' después de 'while'")
		return nil
	}

	// Avanzar para leer la condición
	p.nextToken()
	fmt.Printf("Analizando condición del while, token actual: '%s'\n", p.curToken.Lexeme)
	
	// Analizar la condición del while
	dowhile.Condicion = p.parseExpresion()
	if dowhile.Condicion == nil {
		fmt.Println("Error: no se pudo analizar la condición del while")
		return nil
	}

	fmt.Println("Condición analizada correctamente")

	// Siguiente token debe ser ')'
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		fmt.Println("Error: se esperaba ')' después de la condición")
		return nil
	}

	// Siguiente token debe ser ';'
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		fmt.Println("Error: se esperaba ';' al final del do-while")
		return nil
	}

	fmt.Println("Análisis de do-while completado con éxito")
	return dowhile
}

// parseAsignacion analiza una asignación (a = 5;)
func (p *Parser) parseExpresionAsignacion() *ExpresionAsignacion {
	asign := &ExpresionAsignacion{
		Nombre: p.curToken.Lexeme,
	}

	// Siguiente token debe ser '='
	if !p.expectPeek(lexer.TOKEN_ASSIGN) {
		return nil
	}

	p.nextToken()
	asign.Valor = p.parseExpresion()

	// Debe terminar con ';'
	if !p.expectPeek(lexer.TOKEN_SEMI) {
		return nil
	}

	return asign
}

// parseDeclaracionAsignacion convierte una expresión de asignación en una declaración
func (p *Parser) parseDeclaracionAsignacion() Declaracion {
	return &DeclaracionAsignacion{
		Asignacion: p.parseExpresionAsignacion(),
	}
}

// parseExpresion analiza una expresión
func (p *Parser) parseExpresion() Expresion {
	// Implementación básica para las expresiones
	// Este es un parser simple, en una implementación real se usaría
	// precedencia de operadores y análisis recursivo
	var exp Expresion

	fmt.Printf("Analizando expresión, token actual: '%s'\n", p.curToken.Lexeme)

	switch p.curToken.Type {
	case lexer.TOKEN_IDENT:
		exp = &ExpresionIdentificador{Valor: p.curToken.Lexeme}
		fmt.Printf("Expresión es identificador: '%s'\n", p.curToken.Lexeme)
		
		// Si el siguiente token es un operador, es una expresión binaria
		if p.peekTokenIs(lexer.TOKEN_PLUS) || p.peekTokenIs(lexer.TOKEN_MULT) || p.peekTokenIs(lexer.TOKEN_EQUAL) {
			fmt.Printf("Encontrado operador: '%s', creando expresión binaria\n", p.peekToken.Lexeme)
			p.nextToken()
			binaria := &ExpresionBinaria{
				Izquierda: exp,
				Operador:  p.curToken.Lexeme,
			}
			
			p.nextToken()
			fmt.Printf("Analizando lado derecho de la expresión binaria, token: '%s'\n", p.curToken.Lexeme)
			binaria.Derecha = p.parseExpresion()
			
			if binaria.Derecha == nil {
				fmt.Println("Error: lado derecho de la expresión binaria es nulo")
				return nil
			}
			
			fmt.Println("Expresión binaria creada exitosamente")
			return binaria
		}
		
	case lexer.TOKEN_NUMBER:
		exp = &ExpresionNumero{Valor: p.curToken.Lexeme}
		fmt.Printf("Expresión es número: '%s'\n", p.curToken.Lexeme)
		
		// Si el siguiente token es un operador, es una expresión binaria
		if p.peekTokenIs(lexer.TOKEN_PLUS) || p.peekTokenIs(lexer.TOKEN_MULT) || p.peekTokenIs(lexer.TOKEN_EQUAL) {
			fmt.Printf("Encontrado operador: '%s', creando expresión binaria\n", p.peekToken.Lexeme)
			p.nextToken()
			binaria := &ExpresionBinaria{
				Izquierda: exp,
				Operador:  p.curToken.Lexeme,
			}
			
			p.nextToken()
			fmt.Printf("Analizando lado derecho de la expresión binaria, token: '%s'\n", p.curToken.Lexeme)
			binaria.Derecha = p.parseExpresion()
			
			if binaria.Derecha == nil {
				fmt.Println("Error: lado derecho de la expresión binaria es nulo")
				return nil
			}
			
			fmt.Println("Expresión binaria creada exitosamente")
			return binaria
		}
		
	default:
		p.agregarError(fmt.Sprintf("Expresión inesperada con token %s", p.curToken.Lexeme))
		fmt.Printf("Error: token inesperado en expresión: '%s'\n", p.curToken.Lexeme)
		return nil
	}

	return exp
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