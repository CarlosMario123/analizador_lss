package api

import (
	"net/http"
)

// Router configura las rutas de la API
func SetupRouter() http.Handler {
	// Crear el manejador
	handler := NewAnalyzerHandler()
	
	// Crear el multiplexor
	mux := http.NewServeMux()
	
	// Configurar rutas
	mux.HandleFunc("/api/analyze", handler.AnalizarCodigo)
	
	// Agregar middleware para logging
	return loggingMiddleware(mux)
}

// loggingMiddleware registra las solicitudes HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Loggear la solicitud (en una implementación real usaríamos un logger)
		println(r.Method, r.URL.Path)
		
		// Llamar al siguiente handler
		next.ServeHTTP(w, r)
	})
}