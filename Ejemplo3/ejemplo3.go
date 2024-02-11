package main

/*--------------------------Import--------------------------*/

import (
	"bufio"         // Lectura de datos
	"bytes"         // Manipulacion de bytes
	"encoding/gob"  // Codificacion y decodificacion de datos
	"errors"        // Manejo de errores
	"fmt"           // Impresion de datos
	"io"            // Entrada y salida de datos
	"math/rand"     // Numeros aleatorios
	"os"            // Sistema operativo
	"os/exec"       // Ejecucion de comandos
	"path/filepath" // Rutas de archivos
	"strconv"       // Conversion de datos
	"strings"       // Manipulacion de cadenas
	"time"          // Fecha y hora
)

/*--------------------------/Import--------------------------*/

/*--------------------------Structs--------------------------*/

// Master Boot Record (MBR)
type mbr = struct {
	Mbr_tamano         [100]byte
	Mbr_fecha_creacion [100]byte
	Mbr_dsk_signature  [100]byte
}

/*--------------------------/Structs--------------------------*/

/*--------------------------Metodos o Funciones--------------------------*/

// Metodo principal
func main() {
	analizar()
}

// Muestra el mensaje de error
func msg_error(err error) {
	fmt.Println("[ERROR] ", err)
}

// Obtiene y lee el comando
func analizar() {
	finalizar := false
	fmt.Println("Ejemplo 3: MKDISK 1.0 y REP 1.0")
	reader := bufio.NewReader(os.Stdin)

	//  Pide constantemente un comando
	for !finalizar {
		fmt.Print("Ingrese un comando: ")
		// Lee hasta que presione ENTER
		comando, _ := reader.ReadString('\n')

		if strings.Contains(comando, "exit") {
			/* SALIR */
			finalizar = true
		} else if strings.Contains(comando, "EXIT") {
			/* SALIR */
			finalizar = true
		} else {
			// Si no es vacio o el comando EXIT
			if comando != "" && comando != "exit\n" && comando != "EXIT\n" {
				// Obtener comando y parametros
				split_comando(comando)
			}
		}
	}
}

// Separa los diferentes comando con sus parametros si tienen
func split_comando(comando string) {
	var commandArray []string
	// Elimina los saltos de linea y retornos de carro
	comando = strings.Replace(comando, "\n", "", 1)
	comando = strings.Replace(comando, "\r", "", 1)

	// Banderas para verficar comentarios
	band_comentario := false

	if strings.Contains(comando, "pause") {
		// Comando sin Parametros
		commandArray = append(commandArray, comando)
	} else if strings.Contains(comando, "#") {
		// Comentario
		band_comentario = true
		fmt.Println(comando)
	} else {
		// Comando con Parametros
		commandArray = strings.Split(comando, " -")
	}

	// Ejecuta el comando leido si no es un comentario
	if !band_comentario {
		ejecutar_comando(commandArray)
	}
}

// Identifica y ejecuta el comando encontrado
func ejecutar_comando(commandArray []string) {
	// Convierte el comando a minusculas
	data := strings.ToLower(commandArray[0])

	// Identifica el comando a ejecutar
	if data == "mkdisk" {
		/* MKDISK */
		mkdisk(commandArray)
	} else if data == "rep" {
		/* REP */
		rep()
	} else if data == "execute" {
		/* EXECUTE */
		execute()
	} else {
		/* ERROR */
		fmt.Println("[ERROR] El comando no fue reconocido...")
	}
}

/*--------------------------/Metodos o Funciones--------------------------*/

/*--------------------------Comandos--------------------------*/

/* MKDISK */
func mkdisk(commandArray []string) {
	fmt.Println("[MENSAJE] El comando MKDISK aqui inicia")

	// Variables para los valores de los parametros
	val_size := 0
	val_fit := ""
	val_unit := ""
	val_path := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_size := false
	band_fit := false
	band_unit := false
	band_path := false
	band_error := false

	// Obtengo solo los parametros validos
	for i := 1; i < len(commandArray); i++ {
		aux_data := strings.SplitAfter(commandArray[i], "=")
		data := strings.ToLower(aux_data[0])
		val_data := aux_data[1]

		// Identifica los parametos
		switch {
		/* PARAMETRO OBLIGATORIO -> SIZE */
		case strings.Contains(data, "size="):
			// Valido si el parametro ya fue ingresado
			if band_size {
				fmt.Println("[ERROR] El parametro -size ya fue ingresado...")
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_size = true

			// Conversion a entero
			aux_size, err := strconv.Atoi(val_data)
			val_size = aux_size

			// ERROR de conversion
			if err != nil {
				msg_error(err)
			}

			// Valido que el tamaño sea positivo
			if val_size < 0 {
				band_error = true
				fmt.Println("[ERROR] El parametro -size es negativo...")
				break
			}
		/* PARAMETRO OPCIONAL -> FIT */
		case strings.Contains(data, "fit="):
			// Valido si el parametro ya fue ingresado
			if band_fit {
				fmt.Println("[ERROR] El parametro -fit ya fue ingresado...")
				band_error = true
				break
			}

			// Le quito las comillas y lo paso a minusculas
			val_fit = strings.Replace(val_data, "\"", "", 2)
			val_fit = strings.ToLower(val_fit)

			if val_fit == "bf" { // Best Fit
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "b"
			} else if val_fit == "ff" { // First Fit
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "f"
			} else if val_fit == "wf" { // Worst Fit
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "w"
			} else {
				fmt.Println("[ERROR] El Valor del parametro -fit no es valido...")
				band_error = true
				break
			}
		/* PARAMETRO OPCIONAL -> UNIT */
		case strings.Contains(data, "unit="):
			// Valido si el parametro ya fue ingresado
			if band_unit {
				fmt.Println("[ERROR] El parametro -unit ya fue ingresado...")
				band_error = true
				break
			}

			// Reemplaza comillas y lo paso a minusculas
			val_unit = strings.Replace(val_data, "\"", "", 2)
			val_unit = strings.ToLower(val_unit)

			// valido que tenga unidades validas
			if val_unit == "k" || val_unit == "m" { // Kilobytes o Megabytes
				// Activo la bandera del parametro
				band_unit = true
			} else {
				// Parametro no valido
				fmt.Println("[ERROR] El Valor del parametro -unit no es valido...")
				band_error = true
				break
			}
		/* PARAMETRO OBLIGATORIO -> PATH */
		// No se utiliza en el proyecto, es una ruta fija para el disco
		case strings.Contains(data, "path="):
			// Valido si el parametro ya fue ingresado
			if band_path {
				fmt.Println("[ERROR] El parametro -path ya fue ingresado...")
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			fmt.Println("[ERROR] Parametro no valido...")
		}
	}

	// Verifico si no hay errores
	if !band_error {
		// Verifico que el parametro "Path" (Obligatorio) este ingresado
		if band_path {
			// Verifico que el parametro "Size" (Obligatorio) este ingresado
			if band_size {
				total_size := 1024
				master_boot_record := mbr{}

				// Disco -> Archivo Binario
				crear_disco(val_path)

				// Fecha
				fecha := time.Now()
				str_fecha := fecha.Format("02/01/2006 15:04:05")

				// Copio valor al Struct
				copy(master_boot_record.Mbr_fecha_creacion[:], str_fecha)

				// Numero aleatorio
				rand.Seed(time.Now().UnixNano())
				min := 0
				max := 100
				num_random := rand.Intn(max-min+1) + min

				// Copio valor al Struct
				copy(master_boot_record.Mbr_dsk_signature[:], strconv.Itoa(int(num_random)))

				// Verifico si existe el parametro "Unit" (Opcional)
				if band_unit {
					// Megabytes
					if val_unit == "m" {
						copy(master_boot_record.Mbr_tamano[:], strconv.Itoa(int(val_size*1024*1024)))
						total_size = val_size * 1024
					} else {
						// Kilobytes
						copy(master_boot_record.Mbr_tamano[:], strconv.Itoa(int(val_size*1024)))
						total_size = val_size
					}
				} else {
					// Si no especifica -> Megabytes
					copy(master_boot_record.Mbr_tamano[:], strconv.Itoa(int(val_size*1024*1024)))
					total_size = val_size * 1024
				}

				// Convierto de entero a string
				str_total_size := strconv.Itoa(total_size)

				// Comando para definir el tamaño (Kilobytes) y llenarlo de ceros
				cmd := exec.Command("/bin/sh", "-c", "dd if=/dev/zero of=\""+val_path+"\" bs=1024 count="+str_total_size)
				cmd.Dir = "/"
				_, err := cmd.Output()

				// ERROR
				if err != nil {
					msg_error(err)
				}

				// Se escriben los datos en disco

				// Apertura del archivo
				disco, err := os.OpenFile(val_path, os.O_RDWR, 0660)

				// ERROR
				if err != nil {
					msg_error(err)
				}

				// Conversion de struct a bytes
				mbr_byte := struct_a_bytes(master_boot_record)

				// Se posiciona al inicio del archivo para guardar la informacion del disco
				newpos, err := disco.Seek(0, os.SEEK_SET)

				// ERROR
				if err != nil {
					msg_error(err)
				}

				// Escritura de struct en archivo binario
				_, err = disco.WriteAt(mbr_byte, newpos)

				// ERROR
				if err != nil {
					msg_error(err)
				}

				disco.Close()
			}
		}
	}

	fmt.Println("[MENSAJE] El comando MKDISK aqui finaliza")
}

/* EXECUTE */
func execute() {
	fmt.Print("Aquí deberias agregar execute...")
	fmt.Scanln()
}

// Muestra los datos en el disco
func rep() {
	fin_archivo := false
	var empty [100]byte
	mbr_empty := mbr{}
	cont := 0

	fmt.Println("Reporte de MKDISK:")

	// Apertura de archivo
	disco, err := os.OpenFile("/home/andres-pontaza/Escritorio/MIA/P1/C.dsk", os.O_RDWR, 0660)

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Calculo del tamano de struct en bytes
	mbr2 := struct_a_bytes(mbr_empty)
	sstruct := len(mbr2)

	// RECORRE CADA STRUCT DEL ARCHIVO
	for !fin_archivo {
		// Lectrura de conjunto de bytes en archivo binario
		lectura := make([]byte, sstruct)
		_, err = disco.ReadAt(lectura, int64(sstruct*cont))

		// ERROR
		if err != nil && err != io.EOF {
			msg_error(err)
		}

		// Conversion de bytes a struct
		mbr := bytes_a_struct_mbr(lectura)
		sstruct = len(lectura)

		// ERROR
		if err != nil {
			msg_error(err)
		}

		if mbr.Mbr_tamano == empty {
			fin_archivo = true
		} else {
			fmt.Print("Tamaño: ")
			fmt.Println(string(mbr.Mbr_tamano[:]))
			fmt.Print("Fecha: ")
			fmt.Println(string(mbr.Mbr_fecha_creacion[:]))
			fmt.Print("Signature: ")
			fmt.Println(string(mbr.Mbr_dsk_signature[:]))
		}

		cont++
	}
	disco.Close()
}

/*--------------------------/Comandos--------------------------*/

// Crea el archivo que simula ser un disco duro
func crear_disco(ruta string) {
	aux, err := filepath.Abs(ruta)

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Crea el directiorio de forma recursiva
	cmd1 := exec.Command("/bin/sh", "-c", "sudo mkdir -p '"+filepath.Dir(aux)+"'")
	cmd1.Dir = "/"
	_, err1 := cmd1.Output()

	// ERROR
	if err1 != nil {
		msg_error(err)
	}

	// Da los permisos al directorio
	cmd2 := exec.Command("/bin/sh", "-c", "sudo chmod -R 777 '"+filepath.Dir(aux)+"'")
	cmd2.Dir = "/"
	_, err2 := cmd2.Output()

	// ERROR
	if err2 != nil {
		msg_error(err)
	}

	// Verifica si existe la ruta para el archivo
	if _, err := os.Stat(filepath.Dir(aux)); errors.Is(err, os.ErrNotExist) {
		if err != nil {
			fmt.Println("[FAILURE] No se pudo crear el disco...")
		}
	}
}

// Codifica de Struct a []Bytes
func struct_a_bytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)

	// ERROR
	if err != nil && err != io.EOF {
		msg_error(err)
	}

	return buf.Bytes()
}

// Decodifica de [] Bytes a Struct
func bytes_a_struct_mbr(s []byte) mbr {
	p := mbr{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)

	// ERROR
	if err != nil && err != io.EOF {
		msg_error(err)
	}

	return p
}

/*--------------------------/Metodos o Funciones--------------------------*/
