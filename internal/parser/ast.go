package parser

// Nodo es la interfaz base para todos los nodos del AST
type Nodo interface {
	TokenLiteral() string
}

// Expresion representa una expresión en el AST
type Expresion interface {
	Nodo
	esExpresion()
}

// Declaracion representa una declaración en el AST
type Declaracion interface {
	Nodo
	esDeclaracion()
}

// Programa representa el programa completo
type Programa struct {
	Declaraciones []Declaracion
}

func (p *Programa) TokenLiteral() string {
	if len(p.Declaraciones) > 0 {
		return p.Declaraciones[0].TokenLiteral()
	}
	return ""
}

// DeclaracionVariable representa una declaración de variable (int a = 0;)
type DeclaracionVariable struct {
	Tipo     string    // Tipo de la variable (int, float, etc.)
	Nombre   string    // Nombre de la variable
	Valor    Expresion // Valor asignado
}

func (dv *DeclaracionVariable) esDeclaracion() {}
func (dv *DeclaracionVariable) TokenLiteral() string { return dv.Tipo }

// ExpresionIdentificador representa un identificador
type ExpresionIdentificador struct {
	Valor string
}

func (ei *ExpresionIdentificador) esExpresion() {}
func (ei *ExpresionIdentificador) TokenLiteral() string { return ei.Valor }

// ExpresionNumero representa un literal numérico
type ExpresionNumero struct {
	Valor string
}

func (en *ExpresionNumero) esExpresion() {}
func (en *ExpresionNumero) TokenLiteral() string { return en.Valor }

// ExpresionBinaria representa una operación binaria (a + b, a * b)
type ExpresionBinaria struct {
	Izquierda Expresion
	Operador  string
	Derecha   Expresion
}

func (eb *ExpresionBinaria) esExpresion() {}
func (eb *ExpresionBinaria) TokenLiteral() string { return eb.Operador }

// ExpresionAsignacion representa una asignación (a = 5)
type ExpresionAsignacion struct {
	Nombre string
	Valor  Expresion
}

func (ea *ExpresionAsignacion) esExpresion() {}
func (ea *ExpresionAsignacion) TokenLiteral() string { return "=" }

// DeclaracionAsignacion envuelve una ExpresionAsignacion para implementar Declaracion
type DeclaracionAsignacion struct {
	Asignacion *ExpresionAsignacion
}

func (da *DeclaracionAsignacion) esDeclaracion() {}
func (da *DeclaracionAsignacion) TokenLiteral() string { return "=" }

// DeclaracionDoWhile representa una estructura do-while
type DeclaracionDoWhile struct {
	Cuerpo      []Declaracion
	Condicion   Expresion
}

func (dw *DeclaracionDoWhile) esDeclaracion() {}
func (dw *DeclaracionDoWhile) TokenLiteral() string { return "do" }

// Error de sintaxis
type ErrorSintactico struct {
	Mensaje string
	Linea   int
	Columna int
}