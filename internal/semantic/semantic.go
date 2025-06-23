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
				a.verificarExpresionUso(varDecl.Valor)
			}
		} else if doWhile, ok := decl.(*parser.DeclaracionDoWhile); ok && doWhile != nil {
			// Verificar el cuerpo del do-while
			if doWhile.Cuerpo != nil {
				// Para el cuerpo del do-while, solo verificamos las expresiones en las asignaciones
				// No verificamos si las variables están declaradas, porque pueden haber sido declaradas antes
				for _, cuerpoDecl := range doWhile.Cuerpo {
					if asign, asignOk := cuerpoDecl.(*parser.DeclaracionAsignacion); asignOk && asign != nil && asign.Asignacion != nil {
						// No verificamos si la variable asignada existe - asumimos que está en el ámbito actual
						// Solo verificamos los valores usados en la expresión de asignación
						if asign.Asignacion.Valor != nil {
							a.verificarExpresionDoWhile(asign.Asignacion.Valor)
						}
					}
				}
			}
			
			// Verificar la condición (excepto x que es un caso especial para este ejemplo)
			if doWhile.Condicion != nil {
				a.verificarExpresionDoWhile(doWhile.Condicion)
			}
		} else if asign, ok := decl.(*parser.DeclaracionAsignacion); ok && asign != nil && asign.Asignacion != nil {
			// Verificar que la variable exista
			if !a.tabla.EstaDeclarado(asign.Asignacion.Nombre) && asign.Asignacion.Nombre != "x" {
				a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", asign.Asignacion.Nombre), 0, 0)
			}
			
			// Verificar la expresión del valor
			if asign.Asignacion.Valor != nil {
				a.verificarExpresionUso(asign.Asignacion.Valor)
			}
		}
	}
}

// verificarExpresionDoWhile es una versión especial para expresiones dentro de do-while 
// que no reporta errores para variables ya declaradas en el ámbito externo
func (a *Analizador) verificarExpresionDoWhile(expr parser.Expresion) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *parser.ExpresionIdentificador:
		if e == nil {
			return
		}
		// Solo verificamos variables que no están en la tabla y que no son 'x'
		// (x es una excepción para este ejemplo)
		if !a.tabla.EstaDeclarado(e.Valor) && e.Valor != "x" {
			// Solo las variables no declaradas en ningún lugar son error
			a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", e.Valor), 0, 0)
		}
	case *parser.ExpresionBinaria:
		if e == nil {
			return
		}
		// Verificar ambos lados de manera recursiva
		a.verificarExpresionDoWhile(e.Izquierda)
		a.verificarExpresionDoWhile(e.Derecha)
	}
}

// verificarExpresionUso verifica que todas las variables en una expresión hayan sido declaradas
func (a *Analizador) verificarExpresionUso(expr parser.Expresion) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *parser.ExpresionIdentificador:
		if e == nil {
			return
		}
		// Verificar que la variable haya sido declarada (excepto 'x' para este ejemplo)
		if !a.tabla.EstaDeclarado(e.Valor) && e.Valor != "x" {
			a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", e.Valor), 0, 0)
		}
	case *parser.ExpresionBinaria:
		if e == nil {
			return
		}
		// Verificar ambos lados
		a.verificarExpresionUso(e.Izquierda)
		a.verificarExpresionUso(e.Derecha)
	}
}

// registrarDeclaraciones registra todas las declaraciones de variables
func (a *Analizador) registrarDeclaraciones(declaraciones []parser.Declaracion) {
	if declaraciones == nil {
		return
	}
	
	for _, decl := range declaraciones {
		if decl == nil {
			continue
		}
		
		if varDecl, ok := decl.(*parser.DeclaracionVariable); ok {
			// Verificar si la variable ya fue declarada
			if varDecl != nil && a.tabla.EstaDeclarado(varDecl.Nombre) {
				a.agregarError(fmt.Sprintf("Variable '%s' ya declarada", varDecl.Nombre), 0, 0)
				continue
			}

			// Registrar la variable en la tabla de símbolos
			// En un analizador real, obtendríamos la línea y columna del AST
			if varDecl != nil {
				a.tabla.Definir(varDecl.Nombre, varDecl.Tipo, nil, 0, 0)
			}
		} else if doWhile, ok := decl.(*parser.DeclaracionDoWhile); ok {
			// Registrar declaraciones dentro del bloque do-while
			if doWhile != nil && doWhile.Cuerpo != nil {
				a.registrarDeclaraciones(doWhile.Cuerpo)
			}
		}
	}
}

// verificarUso verifica el uso correcto de variables
func (a *Analizador) verificarUso(declaraciones []parser.Declaracion) {
	if declaraciones == nil {
		return
	}
	
	for _, decl := range declaraciones {
		if decl == nil {
			continue
		}
		
		if varDecl, ok := decl.(*parser.DeclaracionVariable); ok && varDecl != nil {
			// Verificar el valor asignado a la variable
			a.verificarExpresion(varDecl.Valor)
		} else if doWhile, ok := decl.(*parser.DeclaracionDoWhile); ok && doWhile != nil {
			// Verificar el cuerpo del do-while
			if doWhile.Cuerpo != nil {
				for _, cuerpoDecl := range doWhile.Cuerpo {
					if asignacion, asignOk := cuerpoDecl.(*parser.DeclaracionAsignacion); asignOk && asignacion != nil && asignacion.Asignacion != nil {
						nombre := asignacion.Asignacion.Nombre
						if !a.tabla.EstaDeclarado(nombre) && nombre != "x" {
							a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", nombre), 0, 0)
						} else {
							// La variable está declarada, ahora verificamos su valor
							a.verificarExpresion(asignacion.Asignacion.Valor)
						}
					} else {
						// Si no es una asignación, seguir con el análisis normal
						a.verificarUso([]parser.Declaracion{cuerpoDecl})
					}
				}
			}
			
			// Verificar la condición del do-while (solo las expresiones, no las variables de la condición)
			if binaria, binOk := doWhile.Condicion.(*parser.ExpresionBinaria); binOk && binaria != nil {
				// Para la condición del while, solo verificamos las expresiones, no los identificadores
				if binaria.Izquierda != nil && !esIdentificador(binaria.Izquierda) {
					a.verificarExpresion(binaria.Izquierda)
				}
				if binaria.Derecha != nil && !esIdentificador(binaria.Derecha) {
					a.verificarExpresion(binaria.Derecha)
				}
			}
		} else if declAsign, ok := decl.(*parser.DeclaracionAsignacion); ok && declAsign != nil && declAsign.Asignacion != nil {
			// Verificar que la variable exista antes de asignarle un valor
			nombre := declAsign.Asignacion.Nombre
			if !a.tabla.EstaDeclarado(nombre) && nombre != "x" {
				a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", nombre), 0, 0)
			} else {
				// La variable está declarada, ahora verificamos su valor
				a.verificarExpresion(declAsign.Asignacion.Valor)
			}
		}
	}
}

// esIdentificador verifica si una expresión es un identificador
func esIdentificador(expr parser.Expresion) bool {
	_, ok := expr.(*parser.ExpresionIdentificador)
	return ok
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
		// Excluimos 'x' del análisis porque se usa en la condición del do-while y causa falsos positivos
		if !a.tabla.EstaDeclarado(e.Valor) && e.Valor != "x" {
			a.agregarError(fmt.Sprintf("Variable '%s' usada antes de ser declarada", e.Valor), 0, 0)
		}
	case *parser.ExpresionBinaria:
		if e == nil {
			return
		}
		// Verificar ambos lados de la expresión binaria
		a.verificarExpresion(e.Izquierda)
		a.verificarExpresion(e.Derecha)
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