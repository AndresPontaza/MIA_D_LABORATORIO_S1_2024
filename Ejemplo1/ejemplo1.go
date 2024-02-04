package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Definición de la estructura de datos
type Profesor struct {
	Tipo        int32
	Id_profesor int32
	CUI         [13]byte
	Nombre      [25]byte
	Curso       [25]byte
}

type Estudiante struct {
	Tipo          int32
	Id_estudiante int32
	CUI           [13]byte
	Nombre        [25]byte
	Carnet        [25]byte
}

func main() {
	CrearArchivo()
	Menu()
}

// Función Menu
func Menu() {
	var ValorMenu string

	fmt.Println("     Ejemplo 1: Archivos Binarios GO")
	fmt.Println("     =================================")
	fmt.Println("     1. Crear Registro de profesor")
	fmt.Println("     2. Crear Registro de estudiante")
	fmt.Println("     3. Mostrar Registros")
	fmt.Println("     4. Salir")
	fmt.Println("     =================================")
	fmt.Print("     Ingrese una opción: ")
	fmt.Scanln(&ValorMenu)

	if ValorMenu == "1" {
		RegistroProfesor()
	} else if ValorMenu == "2" {
		RegistroEstudiante()
	} else if ValorMenu == "3" {
		VerRegistros()
	} else if ValorMenu == "4" {
		os.Exit(0)
	} else {
		fmt.Println("Opción no válida")
	}
	Menu()
}

// Función RegistroProfesor
func RegistroProfesor() {
	var id int32
	var cui string
	var nombre string
	var curso string

	// abrir archivo en modo escritura
	arch, err := os.OpenFile("Registros.dat", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	// Verificar si hubo un error
	if err != nil {
		fmt.Println(err)
		return
	}
	// Cerrar el archivo al terminar la función
	defer arch.Close()

	// Mover el cursor al final del archivo
	arch.Seek(0, io.SeekEnd)

	// crear un profesor
	var profesorNuevo Profesor
	profesorNuevo.Tipo = int32(1)

	fmt.Println("***************************************************************************************")

	// Solicitar ID
	fmt.Print("Ingrese el ID del profesor: ")
	fmt.Scanln(&id)
	profesorNuevo.Id_profesor = id

	// Solicitar CUI
	fmt.Print("Ingrese el CUI del profesor: ")
	fmt.Scanln(&cui)
	copy(profesorNuevo.CUI[:], cui)

	// Solicitar Nombre
	fmt.Print("Ingrese el nombre del profesor: ")
	fmt.Scanln(&nombre)
	copy(profesorNuevo.Nombre[:], nombre)

	// Solicitar Curso
	fmt.Print("Ingrese el curso del profesor: ")
	fmt.Scanln(&curso)
	copy(profesorNuevo.Curso[:], curso)

	// Escribir el profesor en el archivo
	binary.Write(arch, binary.LittleEndian, &profesorNuevo)
	arch.Close()
	fmt.Println("Profesor registrado con éxito")

	fmt.Println("***************************************************************************************")
}

// Función RegistroEstudiante
func RegistroEstudiante() {
	// Completar función
}

// Función VerRegistros
func VerRegistros() {
	// abrir archivo en modo lectura
	arch, err := os.OpenFile("Registros.dat", os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer arch.Close()

	// leer el archivo con bucle para leer todos los registros
	for {
		var profesor Profesor
		err = binary.Read(arch, binary.LittleEndian, &profesor)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		// Imprimir información del registro
		fmt.Println("*****************************************************************")
		fmt.Println("Tipo: ", profesor.Tipo)
		if profesor.Tipo == 1 {
			fmt.Println("Rol: Profesor")
			fmt.Println("ID: ", profesor.Id_profesor)
			fmt.Println("CUI: ", string(profesor.CUI[:]))
			fmt.Println("Nombre: ", string(profesor.Nombre[:]))
			fmt.Println("Curso: ", string(profesor.Curso[:]))
			fmt.Println("*****************************************************************")
			fmt.Println("")
		} else if profesor.Tipo == 2 {
			fmt.Println("Rol: Estudiante")
			fmt.Println("ID: ", profesor.Id_profesor) // Debería ser Id_estudiante
			fmt.Println("CUI: ", string(profesor.CUI[:]))
			fmt.Println("Nombre: ", string(profesor.Nombre[:]))
			fmt.Println("Carnet: ", string(profesor.Curso[:]))
			fmt.Println("*****************************************************************")
			fmt.Println("")
		}
	}
}

// Función CrearArchivo
func CrearArchivo() {
	if _, err := os.Stat("Registros.dat"); os.IsNotExist(err) {
		// Crear archivo si no existe
		arch, err := os.Create("Registros.dat")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer arch.Close()
	}
}
