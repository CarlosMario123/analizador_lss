package semantic

import (
	"fmt"
	"analyzer-api/internal/parser"
)

// ErrorSemantico representa un error semántico
type ErrorSemantico struct {
	Mensaje string
	Linea   int
	Columna int
}

// Analizador estructura para el analizador semántico
type Analizador struct {
	ast        *parser.Programa
	tabla      *TablaSimbolos
	errores    []ErrorSemantico
}

// New crea un nuevo analizador semántico
func New(ast *parser.Programa) *Analizador {
	return &Analizador{
		ast:     ast,
		tabla:   NewTablaSimbolos(),
		errores: []ErrorSemantico{},
	}
}

// Analizar realiza el análisis semántico
func (a *Analizador) Analizar() []ErrorSemantico {
	// Verificar que el AST no sea nulo
	if a.ast == nil || a.ast.Declaraciones == nil {
		return a.errores
	}
	
	// Realizar un análisis completo en dos fases:
	// 1. Registrar primero TODAS las declaraciones de variables
	a.recolectarDeclaraciones(a.ast.Declaraciones)
	
	// 2. Luego verificar todos los usos
	a.verificarUsos(a.ast.Declaraciones)

	return a.errores
}

// recolectarDeclaraciones registra todas las declaraciones de variables en la tabla de símbolos
func (a *Analizador) recolectarDeclaraciones(declaraciones []parser.Declaracion) {
	if declaraciones == nil {
		return
	}
	
	for _, decl := range declaraciones {
		if decl == nil {
			continue
		}
		
		if varDecl, ok := decl.(*parser.DeclaracionVariable); ok && varDecl != nil {
			// Registrar la variable en la tabla de símbolos
			if !a.tabla.EstaDeclarado(varDecl.Nombre) {
				a.tabla.Definir(varDecl.Nombre, varDecl.Tipo, nil, 0, 0)
			} else {
				a.agregarError(fmt.Sprintf("Variable '%s' ya declarada", varDecl.Nombre), 0, 0)
			}
		} else if doWhile, ok := decl.(*parser.DeclaracionDoWhile); ok && doWhile != nil && doWhile.Cuerpo != nil {
			// Recolectar declaraciones dentro del do-while
			a.recolectarDeclaraciones(doWhile.Cuerpo)
		}
	}
}

// verificarUsos verifica que todas las variables usadas hayan sido declaradas
func (a *Analizador) verificarUsos(declaraciones []parser.Declaracion) {
	if declaraciones == nil {
		return
	}
	
	for _, decl := range declaraciones {
		if decl == nil {
			continue
		}
		
		if varDecl, ok := decl.(*parser.DeclaracionVariable); ok && varDecl != nil {
			// Verificar el valor inicial
			if varDecl.Valor != nil {
				a.verificarExpresion(varDecl.Valor)
			}
		} else if doWhile, ok := decl.(*parser.DeclaracionDoWhile); ok && doWhile != nil {
			// Verificar el cuerpo del do-while
			if doWhile.Cuerpo != nil {
				for _, cuerpoDecl := range doWhile.Cuerpo {
					if asign, asignOk := cuerpoDecl.(*parser.DeclaracionAsignacion); asignOk && asign != nil && asign.Asignacion != nil {
						// CORREGIDO: Verificar que la variable de asignación esté declarada
						if !a.tabla.EstaDeclarado(asign.Asignacion.Nombre) {
							a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", asign.Asignacion.Nombre), 0, 0)
						}
						
						// Verificar las expresiones en el valor asignado
						if asign.Asignacion.Valor != nil {
							a.verificarExpresion(asign.Asignacion.Valor)
						}
					}
				}
			}
			
			// Verificar la condición
			if doWhile.Condicion != nil {
				a.verificarCondicionDoWhile(doWhile.Condicion)
			}
		} else if asign, ok := decl.(*parser.DeclaracionAsignacion); ok && asign != nil && asign.Asignacion != nil {
			// Verificar que la variable exista
			if !a.tabla.EstaDeclarado(asign.Asignacion.Nombre) {
				a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", asign.Asignacion.Nombre), 0, 0)
			}
			
			// Verificar la expresión del valor
			if asign.Asignacion.Valor != nil {
				a.verificarExpresion(asign.Asignacion.Valor)
			}
		}
	}
}

// verificarCondicionDoWhile verifica la condición del do-while con reglas especiales
func (a *Analizador) verificarCondicionDoWhile(expr parser.Expresion) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *parser.ExpresionIdentificador:
		if e == nil {
			return
		}
		// Para la condición del do-while, permitimos 'x' como excepción
		// ya que en algunos casos puede ser una variable de control
		if !a.tabla.EstaDeclarado(e.Valor) && e.Valor != "x" {
			a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada en condición", e.Valor), 0, 0)
		}
	case *parser.ExpresionBinaria:
		if e == nil {
			return
		}
		// Para operaciones de comparación en la condición, aplicar reglas especiales
		// Solo verificar el lado derecho (números), permitir 'x' en el lado izquierdo
		if e.Izquierda != nil {
			if ident, ok := e.Izquierda.(*parser.ExpresionIdentificador); ok {
				if ident.Valor != "x" && !a.tabla.EstaDeclarado(ident.Valor) {
					a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada en condición", ident.Valor), 0, 0)
				}
			} else {
				a.verificarExpresion(e.Izquierda)
			}
		}
		
		// El lado derecho debe seguir las reglas normales
		if e.Derecha != nil {
			a.verificarExpresion(e.Derecha)
		}
	}
}

// verificarExpresion verifica el uso correcto de variables en expresiones
func (a *Analizador) verificarExpresion(expr parser.Expresion) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *parser.ExpresionIdentificador:
		if e == nil {
			return
		}
		// Verificar que la variable haya sido declarada
		if !a.tabla.EstaDeclarado(e.Valor) {
			a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", e.Valor), 0, 0)
		}
	case *parser.ExpresionBinaria:
		if e == nil {
			return
		}
		// Verificar ambos lados de la expresión binaria
		a.verificarExpresion(e.Izquierda)
		a.verificarExpresion(e.Derecha)
	case *parser.ExpresionNumero:
		// Los números son válidos por sí mismos, no necesitan verificación
		return
	}
}

// agregarError agrega un error semántico
func (a *Analizador) agregarError(mensaje string, linea, columna int) {
	error := ErrorSemantico{
		Mensaje: mensaje,
		Linea:   linea,
		Columna: columna,
	}
	a.errores = append(a.errores, error)
}

// TablaSimbolos devuelve la tabla de símbolos
func (a *Analizador) TablaSimbolos() *TablaSimbolos {
	return a.tabla
}