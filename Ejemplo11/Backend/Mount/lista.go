package Mount

import (
	"strconv"
)

type Nodo struct {
	id        string
	Direccion string
	Nombre    string
	Letra     string
	Num       int
	Siguiente *Nodo
}

type Lista struct {
	Primero *Nodo
	Ultimo  *Nodo
}

// Nuevo nodo con datos de la particion
func New_nodo(id string, direccion string, nombre string, letra string, num int) *Nodo {
	return &Nodo{id, direccion, nombre, letra, num, nil}
}

// Nueva lista para los nodos
func New_lista() *Lista {
	return &Lista{nil, nil}
}

// Inserta particion
func Insertar(nuevo *Nodo, lista *Lista) {
	// Si esta vacia
	if lista.Primero == nil {
		lista.Primero = nuevo
		lista.Ultimo = nuevo
	} else {
		lista.Ultimo.Siguiente = nuevo
		lista.Ultimo = lista.Ultimo.Siguiente
	}
}

// Elimina particion por nombre
func Eliminar(id string, lista *Lista) int {
	aux := lista.Primero

	// Si esta vacia
	if aux != nil {
		if aux.id == id {
			lista.Primero = aux.Siguiente
			return 1
		} else {
			var aux2 *Nodo = nil

			for aux != nil {
				if aux.id == id {
					aux2.Siguiente = aux.Siguiente
					return 1
				}
				aux2 = aux
				aux = aux.Siguiente
			}
		}
	}

	return 0
}

// Busca la particion por direccion y nombre
func Buscar_particion(direccion string, nombre string, lista *Lista) bool {
	aux := lista.Primero

	for aux != nil {
		if (direccion == aux.Direccion) && (nombre == aux.Nombre) {
			return true
		}
		aux = aux.Siguiente
	}

	return false
}

// Busca por direccion y obtiene el numero
func Buscar_letra(direccion string, lista *Lista) string {
	aux := lista.Primero
	num_ascii := 65

	for aux != nil {
		if (direccion == aux.Direccion) && (string(rune(num_ascii)) == aux.Letra) {
			num_ascii++
		}
		aux = aux.Siguiente
	}

	return string(rune(num_ascii))
}

// Busca por direccion y obtiene el numero
func Buscar_numero(direccion string, lista *Lista) int {
	aux := lista.Primero
	retorno := 1

	for aux != nil {
		if direccion == aux.Direccion {
			return retorno
		}

		aux = aux.Siguiente
		retorno++
	}

	return retorno
}

// Busca por direccion y nombre retorna un boolean
func Buscar_nodo(direccion string, nombre string, lista *Lista) bool {
	aux := lista.Primero

	for aux != nil {
		if (direccion == aux.Direccion) && (aux.Nombre == nombre) {
			return true
		}
		aux = aux.Siguiente
	}

	return false
}

// Busca por id y retorna un nodo
func Obtener_nodo(id string, lista *Lista) *Nodo {
	aux := lista.Primero

	for aux != nil {
		//201602530
		if id == aux.id {
			return aux
		}
		aux = aux.Siguiente
	}

	return nil
}

// Imprime el contenido de la lista
func Imprimir_contenido(lista *Lista) string {
	aux := lista.Primero
	cad := ""

	for aux != nil {
		cad += "===========================================================\\n"
		cad += "Id:        " + aux.id + "\\n"
		cad += "Direccion: " + aux.Direccion + "\\n"
		cad += "Nombre:    " + aux.Nombre + "\\n"
		cad += "Letra:     " + aux.Letra + "\\n"
		cad += "Numero:    " + strconv.Itoa(aux.Num) + "\\n"
		cad += "==========================================================\\n"
		aux = aux.Siguiente
	}

	return cad
}
