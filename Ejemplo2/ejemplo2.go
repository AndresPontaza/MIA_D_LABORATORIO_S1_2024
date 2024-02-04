package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	analizar()
}

func msg_error(err error) {
	fmt.Println("Error: ", err)
}

func analizar() {
	finalizar := false
	fmt.Println("    Ejemplo 2: Interprete de comandos")
	fmt.Println("")
	reader := bufio.NewReader(os.Stdin)
	//  Ciclo para lectura de multiples comandos
	for !finalizar {
		fmt.Print("<mia>: ")
		comando, _ := reader.ReadString('\n')
		if strings.Contains(comando, "exit") {
			finalizar = true
		} else {
			if comando != "" && comando != "exit\n" {
				//  Separacion de comando y parametros
				split_comando(comando)
			}
		}
	}
}

func split_comando(comando string) {
	var commandArray []string
	// Eliminacion de saltos de linea
	comando = strings.Replace(comando, "\n", "", 1)
	comando = strings.Replace(comando, "\r", "", 1)
	// Guardado de parametros
	if strings.Contains(comando, "comando2") {
		commandArray = append(commandArray, comando)
	} else {
		commandArray = strings.Split(comando, " ")
	}
	// Ejecicion de comando leido
	ejecucion_comando(commandArray)
}

func ejecucion_comando(commandArray []string) {
	// Identificacion de comando y ejecucion (caso no casesensitive)
	data := strings.ToLower(commandArray[0])
	if data == "comando1" {
		comandoConAtributos(commandArray)
	} else if data == "comando2" {
		comandoSinAtributos()
	} else {
		fmt.Println("Comando ingresado no valido...")
	}
}

// comando1 -int=1 -str="hola"
func comandoConAtributos(commandArray []string) {
	paramint := 0
	paramstr := ""

	// Lectura de parametros del comando
	for i := 0; i < len(commandArray); i++ {
		data := strings.ToLower(commandArray[i])
		if strings.Contains(data, "-int=") {
			// Eliminacion de prefijo
			strint := strings.Replace(data, "-int=", "", 1)
			// Eliminacion de comillas y saltos de linea
			strint = strings.Replace(strint, "\"", "", 2)
			strint = strings.Replace(strint, "\r", "", 1)
			// Conversion de string a int
			aux, err := strconv.Atoi(strint)
			paramint = aux
			if err != nil {
				msg_error(err)
			}
		} else if strings.Contains(data, "-str=") {
			paramstr = strings.Replace(data, "-str=", "", 1)
			paramstr = strings.Replace(paramstr, "\"", "", 2)
		}
	}

	// Resumen de accion realizada
	fmt.Print("Ejecucion de comando con atributos")
	fmt.Print(" Parametro 1: ")
	fmt.Print(paramint)
	fmt.Print(" Parametro 2: ")
	fmt.Println(paramstr)
}

// comandoSinAtributos
func comandoSinAtributos() {
	fmt.Println("Ejecucion de comando sin atributos")
}
