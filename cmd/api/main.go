package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	
	"analyzer-api/internal/api"
)

func main() {
	// Obtener el puerto del entorno o usar 8080 por defecto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Configurar el router
	router := api.SetupRouter()
	
	// Iniciar el servidor
	fmt.Printf("Servidor iniciado en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}