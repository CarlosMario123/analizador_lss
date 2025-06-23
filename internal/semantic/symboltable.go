package semantic

// Símbolo representa una entrada en la tabla de símbolos
type Simbolo struct {
	Nombre    string
	Tipo      string
	Valor     interface{}
	Declarado bool
	Linea     int
	Columna   int
}

// TablaSimbolos mantiene un registro de las variables declaradas
type TablaSimbolos struct {
	simbolos map[string]Simbolo
}

// NewTablaSimbolos crea una nueva tabla de símbolos
func NewTablaSimbolos() *TablaSimbolos {
	return &TablaSimbolos{
		simbolos: make(map[string]Simbolo),
	}
}

// Definir agrega un símbolo a la tabla
func (ts *TablaSimbolos) Definir(nombre, tipo string, valor interface{}, linea, columna int) {
	ts.simbolos[nombre] = Simbolo{
		Nombre:    nombre,
		Tipo:      tipo,
		Valor:     valor,
		Declarado: true,
		Linea:     linea,
		Columna:   columna,
	}
}

// Actualizar actualiza el valor de un símbolo existente
func (ts *TablaSimbolos) Actualizar(nombre string, valor interface{}) bool {
	if simbolo, ok := ts.simbolos[nombre]; ok {
		simbolo.Valor = valor
		ts.simbolos[nombre] = simbolo
		return true
	}
	return false
}

// Obtener obtiene un símbolo de la tabla
func (ts *TablaSimbolos) Obtener(nombre string) (Simbolo, bool) {
	simbolo, ok := ts.simbolos[nombre]
	return simbolo, ok
}

// EstaDeclarado verifica si un símbolo ya fue declarado
func (ts *TablaSimbolos) EstaDeclarado(nombre string) bool {
	_, ok := ts.simbolos[nombre]
	return ok
}

// ListarSimbolos retorna todos los símbolos de la tabla
func (ts *TablaSimbolos) ListarSimbolos() map[string]Simbolo {
	return ts.simbolos
}