package api

import (
	"encoding/json"
	"net/http"
	 "fmt"
	"analyzer-api/internal/lexer"
	"analyzer-api/internal/parser"
	"analyzer-api/internal/semantic"
	"analyzer-api/pkg/models"
)

// AnalyzerHandler maneja las solicitudes a la API del analizador
type AnalyzerHandler struct{}

// NewAnalyzerHandler crea un nuevo manejador para el analizador
func NewAnalyzerHandler() *AnalyzerHandler {
	return &AnalyzerHandler{}
}

// SolicitudAnalisis representa la solicitud para analizar código
type SolicitudAnalisis struct {
	Codigo string `json:"codigo"`
}

// AnalizarCodigo analiza el código fuente
func (h *AnalyzerHandler) AnalizarCodigo(w http.ResponseWriter, r *http.Request) {
	// Establecer encabezados CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	// Manejar solicitudes OPTIONS (preflight)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Verificar que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	
	// Decodificar la solicitud
	var solicitud SolicitudAnalisis
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&solicitud); err != nil {
		http.Error(w, "Error al decodificar la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	fmt.Println("Código a analizar:", solicitud.Codigo)
	
	// Realizar el análisis
	l := lexer.New(solicitud.Codigo)
	fmt.Println("Analizador léxico creado")
	
	p := parser.New(l)
	fmt.Println("Analizador sintáctico creado")
	
	ast := p.Parse()
	fmt.Println("AST generado")
	
	sem := semantic.New(ast)
	fmt.Println("Analizador semántico creado")
	
	// Crear el resultado
	resultado := models.NuevoResultadoAnalisis(l, p, sem, solicitud.Codigo)
	fmt.Println("Resultado generado")
	
	// Establecer encabezado de tipo de contenido
	w.Header().Set("Content-Type", "application/json")
	
	// Codificar y enviar la respuesta
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resultado); err != nil {
		http.Error(w, "Error al codificar la respuesta: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	fmt.Println("Respuesta enviada correctamente")
}