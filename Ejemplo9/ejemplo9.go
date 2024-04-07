package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Creamos un enrutador Gorilla Mux
	r := mux.NewRouter()

	// Endpoint 1: Ruta de ejemplo
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "¡Hola desde el backend en Go con Gorilla Mux!")
	})

	// Endpoint 2: Ruta de ejemplo
	r.HandleFunc("/saludo/{nombre}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		nombre := vars["nombre"]
		fmt.Fprintf(w, "¡Hola, %s!", nombre)
	}).Methods("GET")

	// Endpoint 3: Ruta de ejemplo
	r.HandleFunc("/despedida/{nombre}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		nombre := vars["nombre"]
		fmt.Fprintf(w, "¡Adiós, %s!", nombre)
	}).Methods("GET")

	// Definir el manejador de CORS para permitir solicitudes desde cualquier origen
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			next.ServeHTTP(w, r)
		})
	}

	// Aplicar el manejador de CORS al enrutador Gorilla Mux
	http.Handle("/", corsHandler(r))

	// Iniciar el servidor en el puerto 8080
	fmt.Println("Servidor escuchando en el puerto 8080...")
	http.ListenAndServe(":8080", nil)
}
