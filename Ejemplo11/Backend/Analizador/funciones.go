package Analizador

import (
	"Ejemplo11/Mount"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
)

/*-------------------------- Structs --------------------------*/

// Master Boot Record
type MBR struct {
	Mbr_tamano         [100]byte
	Mbr_fecha_creacion [100]byte
	Mbr_dsk_signature  [100]byte
	Dsk_fit            [100]byte
	Mbr_partition      [4]Partition
}

// Partition
type Partition struct {
	Part_status [100]byte
	Part_type   [100]byte
	Part_fit    [100]byte
	Part_start  [100]byte
	Part_size   [100]byte
	Part_name   [100]byte
}

// Extended Boot Record
type EBR struct {
	Part_status [100]byte
	Part_fit    [100]byte
	Part_start  [100]byte
	Part_size   [100]byte
	Part_next   [100]byte
	Part_name   [100]byte
}

// Super Bloque
type Super_bloque struct {
	S_filesystem_type   [100]byte
	S_inodes_count      [100]byte
	S_blocks_count      [100]byte
	S_free_blocks_count [100]byte
	S_free_inodes_count [100]byte
	S_mtime             [100]byte
	S_mnt_count         [100]byte
	S_magic             [100]byte
	S_inode_size        [100]byte
	S_block_size        [100]byte
	S_firts_ino         [100]byte
	S_first_blo         [100]byte
	S_bm_inode_start    [100]byte
	S_bm_block_start    [100]byte
	S_inode_start       [100]byte
	S_block_start       [100]byte
}

// Tablas de Inodos
type Inodo struct {
	I_uid   [100]byte
	I_gid   [100]byte
	I_size  [100]byte
	I_atime [100]byte
	I_ctime [100]byte
	I_mtime [100]byte
	I_block [15]byte
	I_type  [100]byte
	I_perm  [100]byte
}

// Bloques de Carpetas
type Bloque_carpeta struct {
	B_content [4]Cotent
}

// Content
type Cotent struct {
	B_name  [100]byte
	B_inodo [100]byte
}

// Bloques de Archivos
type Bloque_archivo struct {
	B_content [100]byte
}

// Estructura para el API
type Cmd_API struct {
	Cmd string `json:"cmd"`
}

/*-------------------------- Variables Globales --------------------------*/
var lista_montajes *Mount.Lista = Mount.New_lista()
var salida_comando string = ""
var graphDot string = ""

/*-------------------------- Analizador --------------------------*/

// Obtiene y lee el comando
func Analizar() {
	fmt.Println("Bienvenido al API de Sistemia")

	mux := http.NewServeMux()

	// Endpoint
	mux.HandleFunc("/analizar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var Content Cmd_API
		body, _ := io.ReadAll(r.Body)
		// Arreglo  de bytes a Json
		json.Unmarshal(body, &Content)
		split_cmd(Content.Cmd)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "` + salida_comando + `" }`))
		// Limpio la salida de comandos
		salida_comando = ""
	})

	fmt.Println("Servidor en el puerto 5000")
	// Configuracion de cors
	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":5000", handler))
}

func split_cmd(cmd string) {
	arr_com := strings.Split(cmd, "\n")

	for i := 0; i < len(arr_com); i++ {
		if arr_com[i] != "" {
			split_comando(arr_com[i])
			salida_comando += "\\n"
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
	} else if strings.Contains(comando, "logout") {
		// Comando sin Parametros
		commandArray = append(commandArray, comando)
	} else if strings.Contains(comando, "#") {
		// Comentario
		band_comentario = true
		salida_comando += comando + "\\n"
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
	} else if data == "rmdisk" {
		/* RMDISK */
		rmdisk(commandArray)
	} else if data == "fdisk" {
		/* FDISK */
		fdisk(commandArray)
	} else if data == "mount" {
		/* MOUNT */
		mount(commandArray)
	} else if data == "mkfs" {
		/* MKFS */
		mkfs(commandArray)
	} else if data == "rep" {
		/* REP */
		rep(commandArray)
	} else if data == "pause" {
		/* PAUSE */
		pause()
	} else {
		/* ERROR */
		salida_comando += "[ERROR] El comando no fue reconocido...\\n"
	}
}

/*-------------------------- Comandos --------------------------*/

/* MKDISK */
func mkdisk(commandArray []string) {
	salida_comando += "[MENSAJE] El comando MKDISK aqui inicia\\n"

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
				salida_comando += "[ERROR] El parametro -size ya fue ingresado...\\n"
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
				band_error = true
				salida_comando += "[ERROR] En la conversio a entero\\n"
				break
			}

			// Valido que el tamaño sea positivo
			if val_size < 0 {
				band_error = true
				salida_comando += "[ERROR] El parametro -size es negativo...\\n"
			}
		/* PARAMETRO OPCIONAL -> FIT */
		case strings.Contains(data, "fit="):
			// Valido si el parametro ya fue ingresado
			if band_fit {
				salida_comando += "[ERROR] El parametro -fit ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Le quito las comillas y lo paso a minusculas
			val_fit = strings.Replace(val_data, "\"", "", 2)
			val_fit = strings.ToLower(val_fit)

			if val_fit == "bf" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "b"
			} else if val_fit == "ff" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "f"
			} else if val_fit == "wf" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "w"
			} else {
				salida_comando += "[ERROR] El Valor del parametro -fit no es valido...\\n"
				band_error = true
				break
			}
		/* PARAMETRO OPCIONAL -> UNIT */
		case strings.Contains(data, "unit="):
			// Valido si el parametro ya fue ingresado
			if band_unit {
				salida_comando += "[ERROR] El parametro -unit ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Reemplaza comillas y lo paso a minusculas
			val_unit = strings.Replace(val_data, "\"", "", 2)
			val_unit = strings.ToLower(val_unit)

			if val_unit == "k" || val_unit == "m" {
				// Activo la bandera del parametro
				band_unit = true
			} else {
				// Parametro no valido
				salida_comando += "[ERROR] El Valor del parametro -unit no es valido...\\n"
				band_error = true
			}
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	// Verifico si no hay errores
	if !band_error {
		// Verifico que el parametro "Path" (Obligatorio) este ingresado
		if band_path {
			// Verifico que el parametro "Size" (Obligatorio) este ingresado
			if band_size {
				total_size := 1024
				master_boot_record := MBR{}

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

				// Verifico si existe el parametro "Fit" (Opcional)
				if band_fit {
					// Copio valor al Struct
					copy(master_boot_record.Dsk_fit[:], val_fit)
				} else {
					// Si no especifica -> "Primer Ajuste"
					copy(master_boot_record.Dsk_fit[:], "f")
				}

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

				// Inicializar Parcticiones
				for i := 0; i < 4; i++ {
					copy(master_boot_record.Mbr_partition[i].Part_status[:], "0")
					copy(master_boot_record.Mbr_partition[i].Part_type[:], "0")
					copy(master_boot_record.Mbr_partition[i].Part_fit[:], "0")
					copy(master_boot_record.Mbr_partition[i].Part_start[:], "-1")
					copy(master_boot_record.Mbr_partition[i].Part_size[:], "0")
					copy(master_boot_record.Mbr_partition[i].Part_name[:], "")
				}

				// Convierto de entero a string
				str_total_size := strconv.Itoa(total_size)

				// Comando para definir el tamaño (Kilobytes) y llenarlo de ceros
				cmd := exec.Command("/bin/sh", "-c", "dd if=/dev/zero of=\""+val_path+"\" bs=1024 count="+str_total_size)
				cmd.Dir = "/"
				_, err := cmd.Output()

				// ERROR
				if err != nil {
					salida_comando += "[ERROR] Al ejecuatar comando en consola\\n"
				}

				// Se escriben los datos en disco

				// Apertura del archivo
				f, err := os.OpenFile(val_path, os.O_RDWR, 0660)

				// ERROR
				if err != nil {
					salida_comando += "[ERROR] Al abrir el archivo\\n"
				} else {
					// Conversion de struct a bytes
					mbr_byte := struct_a_bytes(master_boot_record)

					// Escribo el mbr desde el inicio del archivos
					f.Seek(0, io.SeekStart)
					f.Write(mbr_byte)
					f.Close()

					salida_comando += "[SUCCES] El disco fue creado con exito!\\n"
				}
			}
		}
	}

	salida_comando += "[MENSAJE] El comando MKDISK aqui finaliza\\n"
}

/* RMDISK */
func rmdisk(commandArray []string) {
	salida_comando += "[MENSAJE] El comando RMDISK aqui inicia\\n"

	// Variables para los valores de los parametros
	val_path := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_path := false
	band_error := false

	// Obtengo solo los parametros validos
	for i := 1; i < len(commandArray); i++ {
		aux_data := strings.SplitAfter(commandArray[i], "=")
		data := strings.ToLower(aux_data[0])
		val_data := aux_data[1]

		// Identifica los parametos
		switch {
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	// Verifico si no hay errores
	if !band_error {
		// Verifico que el parametro "Path" (Obligatorio) este ingresado
		if band_path {
			// Si existe el archivo binario
			_, e := os.Stat(val_path)

			if e != nil {
				// Si no existe
				if os.IsNotExist(e) {
					salida_comando += "[ERROR] No existe el disco que desea eliminar...\\n"
					band_path = false
				}
			} else {
				// Elimina el archivo
				cmd := exec.Command("/bin/sh", "-c", "rm \""+val_path+"\"")
				cmd.Dir = "/"
				_, err := cmd.Output()

				// ERROR
				if err != nil {
					salida_comando += "[ERROR] Al ejecutar un comando en consola\\n"
				} else {
					salida_comando += "[SUCCES] El Disco fue eliminado!\\n"
				}

				band_path = false
			}
		} else {
			salida_comando += "[ERROR] el parametro -path no fue ingresado...\\n"
		}
	}

	salida_comando += "[MENSAJE] El comando RMDISK aqui finaliza\\n"
}

/* FDISK */
func fdisk(commandArray []string) {
	salida_comando += "[MENSAJE] El comando FDISK aqui inicia\\n"

	// Variables para los valores de los parametros
	val_size := 0
	val_unit := ""
	val_path := ""
	val_type := ""
	val_fit := ""
	val_name := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_size := false
	band_unit := false
	band_path := false
	band_type := false
	band_fit := false
	band_name := false
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
				salida_comando += "[ERROR] El parametro -size ya fue ingresado...\\n"
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
				salida_comando += "[ERROR] Al convertir a entero\\n"
				band_error = true
				break
			}

			// Valido que el tamaño sea positivo
			if val_size < 0 {
				band_error = true
				salida_comando += "[ERROR] El parametro -size es negativo...\\n"
			}
		/* PARAMETRO OPCIONAL -> UNIT */
		case strings.Contains(data, "unit="):
			// Valido si el parametro ya fue ingresado
			if band_unit {
				salida_comando += "[ERROR] El parametro -unit ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Reemplaza comillas y lo paso a minusculas
			val_unit = strings.Replace(val_data, "\"", "", 2)
			val_unit = strings.ToLower(val_unit)

			if val_unit == "b" || val_unit == "k" || val_unit == "m" {
				// Activo la bandera del parametro
				band_unit = true
			} else {
				// Parametro no valido
				salida_comando += "[ERROR] El Valor del parametro -unit no es valido...\\n"
				band_error = true
			}
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO OPCIONAL -> TYPE */
		case strings.Contains(data, "type="):
			if band_type {
				salida_comando += "[ERROR] El parametro -type ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Reemplaza comillas y lo paso a minusculas
			val_type = strings.Replace(val_data, "\"", "", 2)
			val_type = strings.ToLower(val_type)

			if val_type == "p" || val_type == "e" || val_type == "l" {
				// Activo la bandera del parametro
				band_type = true
			} else {
				// Parametro no valido
				salida_comando += "[ERROR] El Valor del parametro -type no es valido...\\n"
				band_error = true
			}
		/* PARAMETRO OPCIONAL -> FIT */
		case strings.Contains(data, "fit="):
			// Valido si el parametro ya fue ingresado
			if band_fit {
				salida_comando += "[ERROR] El parametro -fit ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Le quito las comillas y lo paso a minusculas
			val_fit = strings.Replace(val_data, "\"", "", 2)
			val_fit = strings.ToLower(val_fit)

			if val_fit == "bf" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "b"
			} else if val_fit == "ff" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "f"
			} else if val_fit == "wf" {
				// Activo la bandera del parametro y obtengo el caracter que me interesa
				band_fit = true
				val_fit = "w"
			} else {
				salida_comando += "[ERROR] El Valor del parametro -fit no es valido...\\n"
				band_error = true
				break
			}
		/* PARAMETRO OBLIGATORIO -> NAME */
		case strings.Contains(data, "name="):
			// Valido si el parametro ya fue ingresado
			if band_name {
				salida_comando += "[ERROR] El parametro -name ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_name = true

			// Reemplaza comillas
			val_name = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	// Verifico si no hay errores
	if !band_error {
		if band_size {
			if band_path {
				if band_name {
					if band_type {
						if val_type == "p" {
							// Primaria
							crear_particion_primaria(val_path, val_name, val_size, val_fit, val_unit)
						} else if val_type == "e" {
							// Extendida
							crear_particion_extendia(val_path, val_name, val_size, val_fit, val_unit)
						} else {
							// Logica
							crear_particion_logica(val_path, val_name, val_size, val_fit, val_unit)
						}
					} else {
						// Si no lo indica se tomara como Primaria
						crear_particion_primaria(val_path, val_name, val_size, val_fit, val_unit)
					}
				} else {
					salida_comando += "[ERROR] El parametro -name no fue ingresado\\n"
				}
			} else {
				salida_comando += "[ERROR] El parametro -path no fue ingresado\\n"
			}
		} else {
			salida_comando += "[ERROR] El parametro -size no fue ingresado\\n"
		}
	}

	salida_comando += "[MENSAJE] El comando FDISK aqui finaliza\\n"
}

/* MOUNT */
func mount(commandArray []string) {
	salida_comando += "[MENSAJE] El comando MOUNT aqui inicia\\n"

	// Variables para los valores de los parametros
	val_path := ""
	val_name := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_path := false
	band_name := false
	band_error := false

	// Obtengo solo los parametros validos
	for i := 1; i < len(commandArray); i++ {
		aux_data := strings.SplitAfter(commandArray[i], "=")
		data := strings.ToLower(aux_data[0])
		val_data := aux_data[1]

		// Identifica los parametos
		switch {
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO OBLIGATORIO -> NAME */
		case strings.Contains(data, "name="):
			// Valido si el parametro ya fue ingresado
			if band_name {
				salida_comando += "[ERROR] El parametro -name ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_name = true

			// Reemplaza comillas
			val_name = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	// Si no hay reptidos
	if !band_error {
		//Parametro obligatorio
		if band_path {
			if band_name {
				index_p := buscar_particion_p_e(val_path, val_name)
				// Si existe
				if index_p != -1 {
					// Apertura del archivo
					f, err := os.OpenFile(val_path, os.O_RDWR, 0660)

					if err == nil {
						mbr_empty := MBR{}

						// Calculo del tamaño de struct en bytes
						mbr2 := struct_a_bytes(mbr_empty)
						sstruct := len(mbr2)

						// Lectrura del archivo binario desde el inicio
						lectura := make([]byte, sstruct)
						f.Seek(0, io.SeekStart)
						f.Read(lectura)

						// Conversion de bytes a struct
						master_boot_record := bytes_a_struct_mbr(lectura)

						// Colocamos la particion ocupada
						copy(master_boot_record.Mbr_partition[index_p].Part_status[:], "2")

						// Conversion de struct a bytes
						mbr_byte := struct_a_bytes(master_boot_record)

						// Se posiciona al inicio del archivo para guardar la informacion del disco
						f.Seek(0, io.SeekStart)
						f.Write(mbr_byte)
						f.Close()

						if Mount.Buscar_particion(val_path, val_name, lista_montajes) {
							salida_comando += "[ERROR] La particion ya esta montada...\\n"
						} else {
							num := Mount.Buscar_numero(val_path, lista_montajes)
							letra := Mount.Buscar_letra(val_path, lista_montajes)
							id := "30" + strconv.Itoa(num) + letra

							var n *Mount.Nodo = Mount.New_nodo(id, val_path, val_name, letra, num)
							Mount.Insertar(n, lista_montajes)
							salida_comando += "[SUCCES] Particion montada con exito!\\n"
							salida_comando += Mount.Imprimir_contenido(lista_montajes)
						}
					} else {
						salida_comando += "[ERROR] No se encuentra el disco...\\n"
					}
				} else {
					//Posiblemente logica
					index_p := buscar_particion_l(val_path, val_name)
					if index_p != -1 {
						// Apertura del archivo
						f, err := os.OpenFile(val_path, os.O_RDWR, 0660)

						if err == nil {
							ebr_empty := EBR{}

							// Calculo del tamaño de struct en bytes
							ebr2 := struct_a_bytes(ebr_empty)
							sstruct := len(ebr2)

							// Lectrura del archivo binario desde el inicio
							lectura := make([]byte, sstruct)
							f.Seek(int64(index_p), io.SeekStart)
							f.Read(lectura)

							// Conversion de bytes a struct
							extended_boot_record := bytes_a_struct_ebr(lectura)

							// Colocamos la particion ocupada
							copy(extended_boot_record.Part_status[:], "2")

							// Conversion de struct a bytes
							mbr_byte := struct_a_bytes(extended_boot_record)

							// Se posiciona al inicio del archivo para guardar la informacion del disco
							f.Seek(int64(index_p), io.SeekStart)
							f.Write(mbr_byte)
							f.Close()

							if Mount.Buscar_particion(val_path, val_name, lista_montajes) {
								salida_comando += "[ERROR] La particion ya esta montada...\\n"
							} else {
								num := Mount.Buscar_numero(val_path, lista_montajes)
								letra := Mount.Buscar_letra(val_path, lista_montajes)
								id := "30" + strconv.Itoa(num) + letra

								var n *Mount.Nodo = Mount.New_nodo(id, val_path, val_name, letra, num)
								Mount.Insertar(n, lista_montajes)
								salida_comando += "[SUCCES] Particion montada con exito!\\n"
								salida_comando += Mount.Imprimir_contenido(lista_montajes)
							}
						} else {
							salida_comando += "[ERROR] No se encuentra el disco...\\n"
						}

					} else {
						salida_comando += "[ERROR] No se encuentra la particion a montar...\\n"
					}
				}
			} else {
				salida_comando += "[ERROR] Parametro -name no definido...\\n"
			}
		} else {
			salida_comando += "[ERROR] Parametro -path no definido...\\n"
		}
	}

	salida_comando += "[MENSAJE] El comando MOUNT aqui finaliza\\n"
}

/* MKFS */
func mkfs(commandArray []string) {
	salida_comando += "[MENSAJE] El comando MKFS aqui inicia\\n"

	// Variables para los valores de los parametros
	val_id := ""
	val_type := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_id := false
	band_type := false
	band_error := false

	// Obtengo solo los parametros validos
	for i := 1; i < len(commandArray); i++ {
		aux_data := strings.SplitAfter(commandArray[i], "=")
		data := strings.ToLower(aux_data[0])
		val_data := aux_data[1]

		// Identifica los parametos
		switch {
		/* PARAMETRO OBLIGATORIO -> ID */
		case strings.Contains(data, "id="):
			if band_id {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_id = true
			val_id = val_data
		/* PARAMETRO OBLIGATORIO -> TYPE */
		case strings.Contains(data, "type="):
			// Valido si el parametro ya fue ingresado
			if band_type {
				salida_comando += "[ERROR] El parametro -name ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_type = true
			val_type = val_data
		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	if !band_error {
		if band_id {
			var aux *Mount.Nodo = Mount.Obtener_nodo(val_id, lista_montajes)
			if aux != nil {
				index := buscar_particion_p_e(aux.Direccion, aux.Nombre)

				// Si existe la particion
				if index != -1 {
					// Apertura del archivo
					f, err := os.OpenFile(aux.Direccion, os.O_RDWR, 0660)

					if err == nil {
						mbr_empty := MBR{}

						// Calculo del tamaño de struct en bytes
						mbr2 := struct_a_bytes(mbr_empty)
						sstruct := len(mbr2)

						// Lectrura del archivo binario desde el inicio
						lectura := make([]byte, sstruct)
						f.Seek(0, io.SeekStart)
						f.Read(lectura)

						// Conversion de bytes a struct
						master_boot_record := bytes_a_struct_mbr(lectura)

						// Obtengo el inicio
						s_part_start := string(master_boot_record.Mbr_partition[index].Part_start[:])
						// Le quito los caracteres null
						s_part_start = strings.Trim(s_part_start, "\x00")
						inicio, _ := strconv.Atoi(s_part_start)

						// Obtengo el espacio utilizado
						s_part_size := string(master_boot_record.Mbr_partition[index].Part_size[:])
						// Le quito los caracteres null
						s_part_size = strings.Trim(s_part_size, "\x00")
						tamano, _ := strconv.Atoi(s_part_size)

						salida_comando += "[MENSAJE] Formateando " + val_type + "\\n"

						formatear_ext2(inicio, tamano, aux.Direccion)

						f.Close()
					} else {
						salida_comando += "[ERROR] No se puede abrir el archivo...\\n"
					}

				} else {
					index = buscar_particion_l(aux.Direccion, aux.Nombre)
					salida_comando += "[MENSAJE] Index de la logica" + strconv.Itoa(index) + "\\n"
				}
			} else {
				salida_comando += "[ERROR] No se encuentra ninguna particion montada con ese id...\\n"
			}
		} else {
			salida_comando += "[ERROR] El Parametro -id no fue ingresado...\\n"
		}
	}

	salida_comando += "[MENSAJE] El comando MKFS aqui finaliza\\n"
}

/* REP */
func rep(commandArray []string) {
	// Variables para los valores de los parametros
	val_name := ""
	val_path := ""
	val_id := ""

	// Banderas para verificar los parametros y ver si se repiten
	band_name := false
	band_path := false
	band_id := false
	band_ruta := false
	band_error := false

	// Limpio la variable global
	graphDot = ""

	// Obtengo solo los parametros validos
	for i := 1; i < len(commandArray); i++ {
		aux_data := strings.SplitAfter(commandArray[i], "=")
		data := strings.ToLower(aux_data[0])
		val_data := aux_data[1]

		// Identifica los parametos
		switch {
		/* PARAMETRO OBLIGATORIO -> NAME */
		case strings.Contains(data, "name="):
			// Valido si el parametro ya fue ingresado
			if band_name {
				salida_comando += "[ERROR] El parametro -name ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_name = true

			// Reemplaza comillas
			val_name = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				salida_comando += "[ERROR] El parametro -path ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_path = true

			// Reemplaza comillas
			val_path = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO OBLIGATORIO -> ID */
		case strings.Contains(data, "id="):
			// Valido si el parametro ya fue ingresado
			if band_id {
				salida_comando += "[ERROR] El parametro -id ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_id = true

			// Reemplaza comillas
			val_id = val_data
		/* PARAMETRO OBLIGATORIO -> RUTA */
		case strings.Contains(data, "ruta="):
			if band_ruta {
				salida_comando += "[ERROR] El parametro -ruta ya fue ingresado...\\n"
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_ruta = true

		/* PARAMETRO NO VALIDO */
		default:
			salida_comando += "[ERROR] Parametro no valido...\\n"
		}
	}

	if !band_error {
		if band_path {
			if band_name {
				if band_id {
					var aux *Mount.Nodo = Mount.Obtener_nodo(val_id, lista_montajes)

					if aux != nil {
						// Reportes validos
						if val_name == "disk" {
							graficar_disk(aux.Direccion, val_path, "jpg")
						} else {
							salida_comando += "[ERROR] Reporte no valido...\\n"
						}
					} else {
						salida_comando += "[ERROR] No encuentra la particion...\\n"
					}
				} else {
					salida_comando += "[ERROR] El parametro -id no fue ingresado...\\n"
				}
			} else {
				salida_comando += "[ERROR] El parametro -name no fue ingresado...\n"
			}
		} else {
			salida_comando += "[ERROR] El parametro -path no fue ingresado...\\n"
		}
	}
}

/* PAUSE */
func pause() {
	salida_comando += "[MENSAJE] Pausa presione enter para continuar...\\n"
}

/*-------------------------- Funciones Auxiliares --------------------------*/

// Verifica o crea la ruta para el disco duro
func crear_disco(ruta string) {
	aux, err := filepath.Abs(ruta)

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] Al abrir el archivo\\n"
	}

	// Crea el directiorio de forma recursiva
	cmd1 := exec.Command("/bin/sh", "-c", "echo 253097 | sudo -S mkdir -p '"+filepath.Dir(aux)+"'")
	cmd1.Dir = "/"
	_, err = cmd1.Output()

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] Al ejecutar el comando\\n"
	}

	// Da los permisos al directorio
	cmd2 := exec.Command("/bin/sh", "-c", "echo 253097 | sudo -S chmod -R 777 '"+filepath.Dir(aux)+"'")
	cmd2.Dir = "/"
	_, err = cmd2.Output()

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] Error al ejecutar el comando\\n"
	}

	// Verifica si existe la ruta para el archivo
	if _, err := os.Stat(filepath.Dir(aux)); errors.Is(err, os.ErrNotExist) {
		if err != nil {
			salida_comando += "[FAILURE] No se pudo crear el disco...\\n"
		}
	}
}

// Crea la Particion Primaria
func crear_particion_primaria(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := MBR{}
	var empty [100]byte

	// Verifico si tiene Ajuste
	if fit != "" {
		aux_fit = fit
	} else {
		// Por default es Peor ajuste
		aux_fit = "w"
	}

	// Verifico si tiene Unidad
	if unit != "" {
		aux_unit = unit

		// *Bytes
		if aux_unit == "b" {
			size_bytes = size
		} else if aux_unit == "k" {
			// *Kilobytes
			size_bytes = size * 1024
		} else {
			// *Megabytes
			size_bytes = size * 1024 * 1024
		}
	} else {
		// Por default Kilobytes
		size_bytes = size * 1024
	}

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] No existe un disco duro con ese nombre...\\n"
	} else {
		// Bandera para ver si hay una particion disponible
		band_particion := false
		// Valor del numero de particion
		num_particion := 0

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		// Si el disco esta creado
		if master_boot_record.Mbr_tamano != empty {
			s_part_start := ""

			// Recorro las 4 particiones
			for i := 0; i < 4; i++ {
				// Antes de comparar limpio la cadena
				s_part_start = string(master_boot_record.Mbr_partition[i].Part_start[:])
				s_part_start = strings.Trim(s_part_start, "\x00")

				// Verifico si en las particiones hay espacio
				if s_part_start == "-1" {
					band_particion = true
					num_particion = i
					break
				}
			}

			// Si hay una particion disponible
			if band_particion {
				espacio_usado := 0
				s_part_size := ""
				i_part_size := 0
				s_part_status := ""

				// Recorro las 4 particiones
				for i := 0; i < 4; i++ {
					// Obtengo el espacio utilizado
					s_part_size = string(master_boot_record.Mbr_partition[i].Part_size[:])
					// Le quito los caracteres null
					s_part_size = strings.Trim(s_part_size, "\x00")
					i_part_size, _ = strconv.Atoi(s_part_size)

					// Obtengo el status de la particion
					s_part_status = string(master_boot_record.Mbr_partition[i].Part_status[:])
					// Le quito los caracteres null
					s_part_status = strings.Trim(s_part_status, "\x00")

					if s_part_status != "1" {
						// Le sumo el valor al espacio
						espacio_usado += i_part_size
					}
				}

				/* Tamaño del disco */

				// Obtengo el tamaño del disco
				s_tamaño_disco := string(master_boot_record.Mbr_tamano[:])
				// Le quito los caracteres null
				s_tamaño_disco = strings.Trim(s_tamaño_disco, "\x00")
				i_tamaño_disco, _ := strconv.Atoi(s_tamaño_disco)

				espacio_disponible := i_tamaño_disco - espacio_usado

				salida_comando += "[ESPACIO DISPONIBLE] " + strconv.Itoa(espacio_disponible) + " Bytes\\n"
				salida_comando += "[ESPACIO NECESARIO] " + strconv.Itoa(size_bytes) + " Bytes\\n"

				// Verifico que haya espacio suficiente
				if espacio_disponible >= size_bytes {
					// Verifico si no existe una particion con ese nombre
					if !existe_particion(direccion, nombre) {
						// Antes de comparar limpio la cadena
						s_dsk_fit := string(master_boot_record.Dsk_fit[:])
						s_dsk_fit = strings.Trim(s_dsk_fit, "\x00")

						/*  Primer Ajuste  */
						if s_dsk_fit == "f" {
							copy(master_boot_record.Mbr_partition[num_particion].Part_type[:], "p")
							copy(master_boot_record.Mbr_partition[num_particion].Part_fit[:], aux_fit)

							// Si esta iniciando
							if num_particion == 0 {
								// Guardo el inicio de la particion y dejo un espacio de separacion
								mbr_empty_byte := struct_a_bytes(mbr_empty)
								copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
							} else {
								// Obtengo el inicio de la particion anterior
								s_part_start_ant := string(master_boot_record.Mbr_partition[num_particion-1].Part_start[:])
								// Le quito los caracteres null
								s_part_start_ant = strings.Trim(s_part_start_ant, "\x00")
								i_part_start_ant, _ := strconv.Atoi(s_part_start_ant)

								// Obtengo el tamaño de la particion anterior
								s_part_size_ant := string(master_boot_record.Mbr_partition[num_particion-1].Part_size[:])
								// Le quito los caracteres null
								s_part_size_ant = strings.Trim(s_part_size_ant, "\x00")
								i_part_size_ant, _ := strconv.Atoi(s_part_size_ant)

								copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(i_part_start_ant+i_part_size_ant))
							}

							copy(master_boot_record.Mbr_partition[num_particion].Part_size[:], strconv.Itoa(size_bytes))
							copy(master_boot_record.Mbr_partition[num_particion].Part_status[:], "0")
							copy(master_boot_record.Mbr_partition[num_particion].Part_name[:], nombre)

							// Se guarda de nuevo el MBR

							// Conversion de struct a bytes
							mbr_byte := struct_a_bytes(master_boot_record)

							// Se posiciona al inicio del archivo para guardar la informacion del disco
							f.Seek(0, io.SeekStart)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion
							s_part_start = string(master_boot_record.Mbr_partition[num_particion].Part_start[:])
							// Le quito los caracteres null
							s_part_start = strings.Trim(s_part_start, "\x00")
							i_part_start, _ := strconv.Atoi(s_part_start)

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start), io.SeekStart)

							// Lo llena de unos
							for i := 0; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							salida_comando += "[SUCCES] La Particion primaria fue creada con exito!\\n"
						} else if s_dsk_fit == "b" {
							/*  Mejor Ajuste  */
							best_index := num_particion

							// Variables para conversiones
							s_part_start_act := ""
							s_part_status_act := ""
							s_part_size_act := ""
							i_part_size_act := 0
							s_part_start_best := ""
							i_part_start_best := 0
							s_part_start_best_ant := ""
							i_part_start_best_ant := 0
							s_part_size_best := ""
							i_part_size_best := 0
							s_part_size_best_ant := ""
							i_part_size_best_ant := 0

							for i := 0; i < 4; i++ {
								// Obtengo el inicio de la particion actual
								s_part_start_act = string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start_act = strings.Trim(s_part_start_act, "\x00")

								// Obtengo el size de la particion actual
								s_part_status_act = string(master_boot_record.Mbr_partition[i].Part_status[:])
								// Le quito los caracteres null
								s_part_status_act = strings.Trim(s_part_status_act, "\x00")

								// Obtengo la posicion de la particion actual
								s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_size_act = strings.Trim(s_part_size_act, "\x00")
								i_part_size_act, _ = strconv.Atoi(s_part_size_act)

								if s_part_start_act == "-1" || (s_part_status_act == "1" && i_part_size_act >= size_bytes) {
									if i != num_particion {
										// Obtengo el tamaño de la particion del mejor indice
										s_part_size_best = string(master_boot_record.Mbr_partition[best_index].Part_size[:])
										// Le quito los caracteres null
										s_part_size_best = strings.Trim(s_part_size_best, "\x00")
										i_part_size_best, _ = strconv.Atoi(s_part_size_best)

										// Obtengo la posicion de la particion actual
										s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
										// Le quito los caracteres null
										s_part_size_act = strings.Trim(s_part_size_act, "\x00")
										i_part_size_act, _ = strconv.Atoi(s_part_size_act)

										if i_part_size_best > i_part_size_act {
											best_index = i
											break
										}
									}
								}
							}

							// Primaria
							copy(master_boot_record.Mbr_partition[best_index].Part_type[:], "p")
							copy(master_boot_record.Mbr_partition[best_index].Part_fit[:], aux_fit)

							// Si esta iniciando
							if best_index == 0 {
								// Guardo el inicio de la particion y dejo un espacio de separacion
								mbr_empty_byte := struct_a_bytes(mbr_empty)
								copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
							} else {
								// Obtengo el inicio de la particion actual
								s_part_start_best_ant = string(master_boot_record.Mbr_partition[best_index-1].Part_start[:])
								// Le quito los caracteres null
								s_part_start_best_ant = strings.Trim(s_part_start_best_ant, "\x00")
								i_part_start_best_ant, _ = strconv.Atoi(s_part_start_best_ant)

								// Obtengo el inicio de la particion actual
								s_part_size_best_ant = string(master_boot_record.Mbr_partition[best_index-1].Part_size[:])
								// Le quito los caracteres null
								s_part_size_best_ant = strings.Trim(s_part_size_best_ant, "\x00")
								i_part_size_best_ant, _ = strconv.Atoi(s_part_size_best_ant)

								copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(i_part_start_best_ant+i_part_size_best_ant))
							}

							copy(master_boot_record.Mbr_partition[best_index].Part_size[:], strconv.Itoa(size_bytes))
							copy(master_boot_record.Mbr_partition[best_index].Part_status[:], "0")
							copy(master_boot_record.Mbr_partition[best_index].Part_name[:], nombre)

							// Se guarda de nuevo el MBR

							// Conversion de struct a bytes
							mbr_byte := struct_a_bytes(master_boot_record)

							// Se posiciona al inicio del archivo para guardar la informacion del disco
							f.Seek(0, io.SeekStart)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion best
							s_part_start_best = string(master_boot_record.Mbr_partition[best_index].Part_start[:])
							// Le quito los caracteres null
							s_part_start_best = strings.Trim(s_part_start_best, "\x00")
							i_part_start_best, _ = strconv.Atoi(s_part_start_best)

							// Conversion de struct a bytes

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start_best), io.SeekStart)

							// Lo llena de unos
							for i := 1; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							salida_comando += "[SUCCES] La Particion primaria fue creada con exito!\\n"
						} else {
							/*  Peor ajuste  */
							worst_index := num_particion

							// Variables para conversiones
							s_part_start_act := ""
							s_part_status_act := ""
							s_part_size_act := ""
							i_part_size_act := 0
							s_part_start_worst := ""
							i_part_start_worst := 0
							s_part_start_worst_ant := ""
							i_part_start_worst_ant := 0
							s_part_size_worst := ""
							i_part_size_worst := 0
							s_part_size_worst_ant := ""
							i_part_size_worst_ant := 0

							for i := 0; i < 4; i++ {
								// Obtengo el inicio de la particion actual
								s_part_start_act = string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start_act = strings.Trim(s_part_start_act, "\x00")

								// Obtengo el size de la particion actual
								s_part_status_act = string(master_boot_record.Mbr_partition[i].Part_status[:])
								// Le quito los caracteres null
								s_part_status_act = strings.Trim(s_part_status_act, "\x00")

								// Obtengo la posicion de la particion actual
								s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_size_act = strings.Trim(s_part_size_act, "\x00")
								i_part_size_act, _ = strconv.Atoi(s_part_size_act)

								if s_part_start_act == "-1" || (s_part_status_act == "1" && i_part_size_act >= size_bytes) {
									if i != num_particion {
										// Obtengo el tamaño de la particion del mejor indice
										s_part_size_worst = string(master_boot_record.Mbr_partition[worst_index].Part_size[:])
										// Le quito los caracteres null
										s_part_size_worst = strings.Trim(s_part_size_worst, "\x00")
										i_part_size_worst, _ = strconv.Atoi(s_part_size_worst)

										// Obtengo la posicion de la particion actual
										s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
										// Le quito los caracteres null
										s_part_size_act = strings.Trim(s_part_size_act, "\x00")
										i_part_size_act, _ = strconv.Atoi(s_part_size_act)

										if i_part_size_worst < i_part_size_act {
											worst_index = i
											break
										}
									}
								}
							}

							// Particiones Primarias
							copy(master_boot_record.Mbr_partition[worst_index].Part_type[:], "p")
							copy(master_boot_record.Mbr_partition[worst_index].Part_fit[:], aux_fit)

							// Se esta iniciando
							if worst_index == 0 {
								// Guardo el inicio de la particion y dejo un espacio de separacion
								mbr_empty_byte := struct_a_bytes(mbr_empty)
								copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
							} else {
								// Obtengo el inicio de la particion anterior
								s_part_start_worst_ant = string(master_boot_record.Mbr_partition[worst_index-1].Part_start[:])
								// Le quito los caracteres null
								s_part_start_worst_ant = strings.Trim(s_part_start_worst_ant, "\x00")
								i_part_start_worst_ant, _ = strconv.Atoi(s_part_start_worst_ant)

								// Obtengo el tamaño de la particion anterior
								s_part_size_worst_ant = string(master_boot_record.Mbr_partition[worst_index-1].Part_size[:])
								// Le quito los caracteres null
								s_part_size_worst_ant = strings.Trim(s_part_size_worst_ant, "\x00")
								i_part_size_worst_ant, _ = strconv.Atoi(s_part_size_worst_ant)

								copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(i_part_start_worst_ant+i_part_size_worst_ant))
							}

							copy(master_boot_record.Mbr_partition[worst_index].Part_size[:], strconv.Itoa(size_bytes))
							copy(master_boot_record.Mbr_partition[worst_index].Part_status[:], "0")
							copy(master_boot_record.Mbr_partition[worst_index].Part_name[:], nombre)

							// Se guarda de nuevo el MBR

							// Conversion de struct a bytes
							mbr_byte := struct_a_bytes(master_boot_record)

							// Escribe desde el inicio del archivo
							f.Seek(0, io.SeekStart)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion best
							s_part_start_worst = string(master_boot_record.Mbr_partition[worst_index].Part_start[:])
							// Le quito los caracteres null
							s_part_start_worst = strings.Trim(s_part_start_worst, "\x00")
							i_part_start_worst, _ = strconv.Atoi(s_part_start_worst)

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start_worst), io.SeekStart)

							// Lo llena de unos
							for i := 1; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							salida_comando += "[SUCCES] La Particion primaria fue creada con exito!\\n"
						}
					} else {
						salida_comando += "[ERROR] Ya existe una particion creada con ese nombre...\\n"
					}
				} else {
					salida_comando += "[ERROR] La particion que desea crear excede el espacio disponible...\\n"
				}
			} else {
				salida_comando += "[ERROR] La suma de particiones primarias y extendidas no debe exceder de 4 particiones...\\n"
				salida_comando += "[MENSAJE] Se recomienda eliminar alguna particion para poder crear otra particion primaria o extendida\\n"
			}
		} else {
			salida_comando += "[ERROR] el disco se encuentra vacio...\\n"
		}

		f.Close()
	}
}

// Crea la Particion Extendida
func crear_particion_extendia(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := MBR{}
	var empty [100]byte

	// Verifico si tiene Ajuste
	if fit != "" {
		aux_fit = fit
	} else {
		// Por default es Peor ajuste
		aux_fit = "w"
	}

	// Verifico si tiene Unidad
	if unit != "" {
		aux_unit = unit

		// *Bytes
		if aux_unit == "b" {
			size_bytes = size
		} else if aux_unit == "k" {
			// *Kilobytes
			size_bytes = size * 1024
		} else {
			// *Megabytes
			size_bytes = size * 1024 * 1024
		}
	} else {
		// Por default Kilobytes
		size_bytes = size * 1024
	}

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] Al abrir el archivo\\n"
	} else {
		// Procede a leer el archivo
		band_particion := false
		band_extendida := false
		num_particion := 0

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		// Si el disco esta creado
		if master_boot_record.Mbr_tamano != empty {
			s_part_type := ""

			// Recorro las 4 particiones
			for i := 0; i < 4; i++ {
				// Antes de comparar limpio la cadena
				s_part_type = string(master_boot_record.Mbr_partition[i].Part_type[:])
				s_part_type = strings.Trim(s_part_type, "\x00")

				if s_part_type == "e" {
					band_extendida = true
					break
				}
			}

			// Si no es extendida
			if !band_extendida {
				s_part_start := ""
				s_part_status := ""
				s_part_size := ""
				i_part_size := 0

				// Recorro las 4 particiones
				for i := 0; i < 4; i++ {
					// Antes de comparar limpio la cadena
					s_part_start = string(master_boot_record.Mbr_partition[i].Part_start[:])
					s_part_start = strings.Trim(s_part_start, "\x00")

					s_part_status = string(master_boot_record.Mbr_partition[i].Part_status[:])
					s_part_status = strings.Trim(s_part_status, "\x00")

					s_part_size = string(master_boot_record.Mbr_partition[i].Part_size[:])
					s_part_size = strings.Trim(s_part_size, "\x00")
					i_part_size, _ = strconv.Atoi(s_part_size)

					// Verifica si existe una particion disponible
					if s_part_start == "-1" || (s_part_status == "1" && i_part_size >= size_bytes) {
						band_particion = true
						num_particion = i
						break
					}
				}

				// Si hay una particion
				if band_particion {
					espacio_usado := 0

					// Recorro las 4 particiones
					for i := 0; i < 4; i++ {
						s_part_status = string(master_boot_record.Mbr_partition[i].Part_status[:])
						s_part_status = strings.Trim(s_part_status, "\x00")

						if s_part_status != "1" {
							// Obtengo el espacio utilizado
							s_part_size = string(master_boot_record.Mbr_partition[i].Part_size[:])
							// Le quito los caracteres null
							s_part_size = strings.Trim(s_part_size, "\x00")
							i_part_size, _ = strconv.Atoi(s_part_size)

							// Le sumo el valor al espacio
							espacio_usado += i_part_size
						}
					}

					/* Tamaño del disco */

					// Obtengo el tamaño del disco
					s_tamaño_disco := string(master_boot_record.Mbr_tamano[:])
					// Le quito los caracteres null
					s_tamaño_disco = strings.Trim(s_tamaño_disco, "\x00")
					i_tamaño_disco, _ := strconv.Atoi(s_tamaño_disco)

					espacio_disponible := i_tamaño_disco - espacio_usado

					salida_comando += "[ESPACIO DISPONIBLE] " + strconv.Itoa(espacio_disponible) + " Bytes\\n"
					salida_comando += "[ESPACIO NECESARIO] " + strconv.Itoa(size_bytes) + " Bytes\\n"

					// Verifico que haya espacio suficiente
					if espacio_disponible >= size_bytes {
						// Verifico si no existe una particion con ese nombre
						if !existe_particion(direccion, nombre) {
							// Antes de comparar limpio la cadena
							s_dsk_fit := string(master_boot_record.Dsk_fit[:])
							s_dsk_fit = strings.Trim(s_dsk_fit, "\x00")

							/*  Primer Ajuste  */
							if s_dsk_fit == "f" {
								copy(master_boot_record.Mbr_partition[num_particion].Part_type[:], "e")
								copy(master_boot_record.Mbr_partition[num_particion].Part_fit[:], aux_fit)

								// Si esta iniciando
								if num_particion == 0 {
									// Guardo el inicio de la particion y dejo un espacio de separacion
									mbr_empty_byte := struct_a_bytes(mbr_empty)
									copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
								} else {
									// Obtengo el inicio de la particion anterior
									s_part_start_ant := string(master_boot_record.Mbr_partition[num_particion-1].Part_start[:])
									// Le quito los caracteres null
									s_part_start_ant = strings.Trim(s_part_start_ant, "\x00")
									i_part_start_ant, _ := strconv.Atoi(s_part_start_ant)

									// Obtengo el tamaño de la particion anterior
									s_part_size_ant := string(master_boot_record.Mbr_partition[num_particion-1].Part_size[:])
									// Le quito los caracteres null
									s_part_size_ant = strings.Trim(s_part_size_ant, "\x00")
									i_part_size_ant, _ := strconv.Atoi(s_part_size_ant)

									copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(i_part_start_ant+i_part_size_ant))
								}

								copy(master_boot_record.Mbr_partition[num_particion].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[num_particion].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[num_particion].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Escribe en la posicion inicial del archivo
								f.Seek(0, io.SeekStart)
								f.Write(mbr_byte)

								// Obtengo el tamaño de la particion
								s_part_start = string(master_boot_record.Mbr_partition[num_particion].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start), io.SeekStart)

								extended_boot_record := EBR{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								// Lo llena de unos
								for i := 0; i < (size_bytes - len(ebr_byte)); i++ {
									f.Write([]byte{1})
								}

								salida_comando += "[SUCCES] La Particion extendida fue creada con exito!\\n"
							} else if s_dsk_fit == "b" {
								/*  Mejor Ajuste  */
								best_index := num_particion

								// Variables para conversiones
								s_part_start_act := ""
								s_part_status_act := ""
								s_part_size_act := ""
								i_part_size_act := 0
								s_part_start_best := ""
								i_part_start_best := 0
								s_part_start_best_ant := ""
								i_part_start_best_ant := 0
								s_part_size_best := ""
								i_part_size_best := 0
								s_part_size_best_ant := ""
								i_part_size_best_ant := 0

								for i := 0; i < 4; i++ {
									// Obtengo el inicio de la particion actual
									s_part_start_act = string(master_boot_record.Mbr_partition[i].Part_start[:])
									// Le quito los caracteres null
									s_part_start_act = strings.Trim(s_part_start_act, "\x00")

									// Obtengo el size de la particion actual
									s_part_status_act = string(master_boot_record.Mbr_partition[i].Part_status[:])
									// Le quito los caracteres null
									s_part_status_act = strings.Trim(s_part_status_act, "\x00")

									// Obtengo la posicion de la particion actual
									s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
									// Le quito los caracteres null
									s_part_size_act = strings.Trim(s_part_size_act, "\x00")
									i_part_size_act, _ = strconv.Atoi(s_part_size_act)

									if s_part_start_act == "-1" || (s_part_status_act == "1" && i_part_size_act >= size_bytes) {
										if i != num_particion {
											// Obtengo el tamaño de la particion del mejor indice
											s_part_size_best = string(master_boot_record.Mbr_partition[best_index].Part_size[:])
											// Le quito los caracteres null
											s_part_size_best = strings.Trim(s_part_size_best, "\x00")
											i_part_size_best, _ = strconv.Atoi(s_part_size_best)

											// Obtengo la posicion de la particion actual
											s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
											// Le quito los caracteres null
											s_part_size_act = strings.Trim(s_part_size_act, "\x00")
											i_part_size_act, _ = strconv.Atoi(s_part_size_act)

											if i_part_size_best > i_part_size_act {
												best_index = i
												break
											}
										}
									}
								}

								// Extendida
								copy(master_boot_record.Mbr_partition[best_index].Part_type[:], "e")
								copy(master_boot_record.Mbr_partition[best_index].Part_fit[:], aux_fit)

								// Si esta iniciando
								if best_index == 0 {
									// Guardo el inicio de la particion y dejo un espacio de separacion
									mbr_empty_byte := struct_a_bytes(mbr_empty)
									copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
								} else {
									// Obtengo el inicio de la particion actual
									s_part_start_best_ant = string(master_boot_record.Mbr_partition[best_index-1].Part_start[:])
									// Le quito los caracteres null
									s_part_start_best_ant = strings.Trim(s_part_start_best_ant, "\x00")
									i_part_start_best_ant, _ = strconv.Atoi(s_part_start_best_ant)

									// Obtengo el inicio de la particion actual
									s_part_size_best_ant = string(master_boot_record.Mbr_partition[best_index-1].Part_size[:])
									// Le quito los caracteres null
									s_part_size_best_ant = strings.Trim(s_part_size_best_ant, "\x00")
									i_part_size_best_ant, _ = strconv.Atoi(s_part_size_best_ant)

									copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(i_part_start_best_ant+i_part_size_best_ant))
								}

								copy(master_boot_record.Mbr_partition[best_index].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[best_index].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[best_index].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Se escribe al inicio del archivo
								f.Seek(0, io.SeekStart)
								f.Write(mbr_byte)

								// Obtengo el inicio de la particion best
								s_part_start_best = string(master_boot_record.Mbr_partition[best_index].Part_start[:])
								// Le quito los caracteres null
								s_part_start_best = strings.Trim(s_part_start_best, "\x00")
								i_part_start_best, _ = strconv.Atoi(s_part_start_best)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start_best), io.SeekStart)

								extended_boot_record := EBR{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start_best)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								// Lo llena de unos
								for i := 0; i < (size_bytes - len(ebr_byte)); i++ {
									f.Write([]byte{1})
								}

								salida_comando += "[SUCCES] La Particion extendida fue creada con exito!\\n"
							} else {
								/*  Peor ajuste  */
								worst_index := num_particion

								// Variables para conversiones
								s_part_start_act := ""
								s_part_status_act := ""
								s_part_size_act := ""
								i_part_size_act := 0
								s_part_start_worst := ""
								i_part_start_worst := 0
								s_part_start_worst_ant := ""
								i_part_start_worst_ant := 0
								s_part_size_worst := ""
								i_part_size_worst := 0
								s_part_size_worst_ant := ""
								i_part_size_worst_ant := 0

								for i := 0; i < 4; i++ {
									// Obtengo el inicio de la particion actual
									s_part_start_act = string(master_boot_record.Mbr_partition[i].Part_start[:])
									// Le quito los caracteres null
									s_part_start_act = strings.Trim(s_part_start_act, "\x00")

									// Obtengo el size de la particion actual
									s_part_status_act = string(master_boot_record.Mbr_partition[i].Part_status[:])
									// Le quito los caracteres null
									s_part_status_act = strings.Trim(s_part_status_act, "\x00")

									// Obtengo la posicion de la particion actual
									s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
									// Le quito los caracteres null
									s_part_size_act = strings.Trim(s_part_size_act, "\x00")
									i_part_size_act, _ = strconv.Atoi(s_part_size_act)

									if s_part_start_act == "-1" || (s_part_status_act == "1" && i_part_size_act >= size_bytes) {
										if i != num_particion {
											// Obtengo el tamaño de la particion del mejor indice
											s_part_size_worst = string(master_boot_record.Mbr_partition[worst_index].Part_size[:])
											// Le quito los caracteres null
											s_part_size_worst = strings.Trim(s_part_size_worst, "\x00")
											i_part_size_worst, _ = strconv.Atoi(s_part_size_worst)

											// Obtengo la posicion de la particion actual
											s_part_size_act = string(master_boot_record.Mbr_partition[i].Part_size[:])
											// Le quito los caracteres null
											s_part_size_act = strings.Trim(s_part_size_act, "\x00")
											i_part_size_act, _ = strconv.Atoi(s_part_size_act)

											if i_part_size_worst < i_part_size_act {
												worst_index = i
												break
											}
										}
									}
								}

								// Particiones Extendidas
								copy(master_boot_record.Mbr_partition[worst_index].Part_type[:], "e")
								copy(master_boot_record.Mbr_partition[worst_index].Part_fit[:], aux_fit)

								// Se esta iniciando
								if worst_index == 0 {
									// Guardo el inicio de la particion y dejo un espacio de separacion
									mbr_empty_byte := struct_a_bytes(mbr_empty)
									copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(len(mbr_empty_byte)))
								} else {
									// Obtengo el inicio de la particion actual
									s_part_start_worst_ant = string(master_boot_record.Mbr_partition[worst_index-1].Part_start[:])
									// Le quito los caracteres null
									s_part_start_worst_ant = strings.Trim(s_part_start_worst_ant, "\x00")
									i_part_start_worst_ant, _ = strconv.Atoi(s_part_start_worst_ant)

									// Obtengo el inicio de la particion actual
									s_part_size_worst_ant = string(master_boot_record.Mbr_partition[worst_index-1].Part_size[:])
									// Le quito los caracteres null
									s_part_size_worst_ant = strings.Trim(s_part_size_worst_ant, "\x00")
									i_part_size_worst_ant, _ = strconv.Atoi(s_part_size_worst_ant)

									copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(i_part_start_worst_ant+i_part_size_worst_ant))
								}

								copy(master_boot_record.Mbr_partition[worst_index].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[worst_index].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[worst_index].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Se escribe desde el inicio del archivo
								f.Seek(0, io.SeekStart)
								f.Write(mbr_byte)

								// Obtengo el inicio de la particion best
								s_part_start_worst = string(master_boot_record.Mbr_partition[worst_index].Part_start[:])
								// Le quito los caracteres null
								s_part_start_worst = strings.Trim(s_part_start_worst, "\x00")
								i_part_start_worst, _ = strconv.Atoi(s_part_start_worst)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start_worst), io.SeekStart)

								extended_boot_record := EBR{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start_worst)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								// Lo llena de unos
								for i := 0; i < (size_bytes - len(ebr_byte)); i++ {
									f.Write([]byte{1})
								}

								salida_comando += "[SUCCES] La Particion extendida fue creada con exito!\\n"
							}
						} else {
							salida_comando += "[ERROR] Ya existe una particion creada con ese nombre...\\n"
						}
					} else {
						salida_comando += "[ERROR] La particion que desea crear excede el espacio disponible...\\n"
					}
				} else {
					salida_comando += "[ERROR] La suma de particiones primarias y extendidas no debe exceder de 4 particiones...\\n"
					salida_comando += "[MENSAJE] Se recomienda eliminar alguna particion para poder crear otra particion primaria o extendida\\n"
				}
			} else {
				salida_comando += "[ERROR] Solo puede haber una particion extendida por disco...\\n"
			}
		} else {
			salida_comando += "[ERROR] el disco se encuentra vacio...\\n"
		}
		f.Close()
	}
}

// Crea la Particion Logica
func crear_particion_logica(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := MBR{}
	ebr_empty := EBR{}
	var empty [100]byte

	// Verifico si tiene Ajuste
	if fit != "" {
		aux_fit = fit
	} else {
		// Por default es Peor ajuste
		aux_fit = "w"
	}

	// Verifico si tiene Unidad
	if unit != "" {
		aux_unit = unit

		// *Bytes
		if aux_unit == "b" {
			size_bytes = size
		} else if aux_unit == "k" {
			// *Kilobytes
			size_bytes = size * 1024
		} else {
			// *Megabytes
			size_bytes = size * 1024 * 1024
		}
	} else {
		// Por default Kilobytes
		size_bytes = size * 1024
	}

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	// ERROR
	if err != nil {
		salida_comando += "[ERROR] No existe el disco duro con ese nombre...\\n"
	} else {
		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		// Si el disco esta creado
		if master_boot_record.Mbr_tamano != empty {
			s_part_type := ""
			num_extendida := -1

			// Recorro las 4 particiones
			for i := 0; i < 4; i++ {
				// Antes de comparar limpio la cadena
				s_part_type = string(master_boot_record.Mbr_partition[i].Part_type[:])
				s_part_type = strings.Trim(s_part_type, "\x00")

				if s_part_type == "e" {
					num_extendida = i
					break
				}
			}

			if !existe_particion(direccion, nombre) {
				if num_extendida != -1 {
					s_part_start := string(master_boot_record.Mbr_partition[num_extendida].Part_start[:])
					s_part_start = strings.Trim(s_part_start, "\x00")
					i_part_start, _ := strconv.Atoi(s_part_start)

					cont := i_part_start

					// Se posiciona en el inicio de la particion
					f.Seek(int64(cont), io.SeekStart)

					// Calculo del tamaño de struct en bytes
					ebr2 := struct_a_bytes(ebr_empty)
					sstruct := len(ebr2)

					// Lectrura del archivo binario desde el inicio
					lectura := make([]byte, sstruct)
					f.Read(lectura)

					// Conversion de bytes a struct
					extended_boot_record := bytes_a_struct_ebr(lectura)

					// Obtencion de datos
					s_part_size_ext := string(extended_boot_record.Part_size[:])
					s_part_size_ext = strings.Trim(s_part_size_ext, "\x00")

					if s_part_size_ext == "0" {
						// Obtencion de datos
						s_part_size := string(master_boot_record.Mbr_partition[num_extendida].Part_size[:])
						s_part_size = strings.Trim(s_part_size, "\x00")
						i_part_size, _ := strconv.Atoi(s_part_size)

						salida_comando += "[ESPACIO DISPONIBLE] " + strconv.Itoa(i_part_size) + " Bytes\\n"
						salida_comando += "[ESPACIO NECESARIO] " + strconv.Itoa(size_bytes) + " Bytes\\n"

						// Si excede el tamaño de la extendida
						if i_part_size < size_bytes {
							salida_comando += "[ERROR] La particion logica a crear excede el espacio disponible de la particion extendida...\\n"
						} else {
							copy(extended_boot_record.Part_status[:], "0")
							copy(extended_boot_record.Part_fit[:], aux_fit)

							// Posicion actual en el archivo
							pos_actual, _ := f.Seek(0, io.SeekCurrent)
							ebr_empty_byte := struct_a_bytes(ebr_empty)

							copy(extended_boot_record.Part_start[:], strconv.Itoa(int(pos_actual)-len(ebr_empty_byte)))
							copy(extended_boot_record.Part_size[:], strconv.Itoa(size_bytes))
							copy(extended_boot_record.Part_next[:], "-1")
							copy(extended_boot_record.Part_name[:], nombre)

							// Obtencion de datos
							s_part_start := string(master_boot_record.Mbr_partition[num_extendida].Part_start[:])
							s_part_start = strings.Trim(s_part_start, "\x00")
							i_part_start, _ := strconv.Atoi(s_part_start)

							// Se posiciona en el inicio de la particion
							ebr_byte := struct_a_bytes(extended_boot_record)
							f.Seek(int64(i_part_start), io.SeekStart)
							f.Write(ebr_byte)

							salida_comando += "[SUCCES] La Particion logica fue creada con exito!\\n"
						}
					} else {
						// Obtencion de datos
						s_part_size := string(master_boot_record.Mbr_partition[num_extendida].Part_size[:])
						s_part_size = strings.Trim(s_part_size, "\x00")
						i_part_size, _ := strconv.Atoi(s_part_size)

						// Obtencion de datos
						s_part_start := string(master_boot_record.Mbr_partition[num_extendida].Part_start[:])
						s_part_start = strings.Trim(s_part_start, "\x00")
						i_part_start, _ := strconv.Atoi(s_part_start)

						salida_comando += "[ESPACIO DISPONIBLE] " + strconv.Itoa(i_part_size+i_part_start) + " Bytes\\n"
						salida_comando += "[ESPACIO NECESARIO] " + strconv.Itoa(size_bytes) + " Bytes\\n"

						// Obtencion de datos
						s_part_next := string(extended_boot_record.Part_next[:])
						s_part_next = strings.Trim(s_part_next, "\x00")
						i_part_next, _ := strconv.Atoi(s_part_next)

						pos_actual, _ := f.Seek(0, io.SeekCurrent)

						for (i_part_next != -1) && (int(pos_actual) < (i_part_size + i_part_start)) {
							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_next), io.SeekStart)

							// Calculo del tamaño de struct en bytes
							ebr2 := struct_a_bytes(ebr_empty)
							sstruct := len(ebr2)

							// Lectrura del archivo binario desde el inicio
							lectura := make([]byte, sstruct)
							f.Read(lectura)

							// Posicion actual
							pos_actual, _ = f.Seek(0, io.SeekCurrent)

							// Conversion de bytes a struct
							extended_boot_record = bytes_a_struct_ebr(lectura)

							if extended_boot_record.Part_next == empty {
								break
							}

							// Obtencion de datos
							s_part_next = string(extended_boot_record.Part_next[:])
							s_part_next = strings.Trim(s_part_next, "\x00")
							i_part_next, _ = strconv.Atoi(s_part_next)
						}

						// Obtencion de datos
						s_part_start_ext := string(extended_boot_record.Part_start[:])
						s_part_start_ext = strings.Trim(s_part_start_ext, "\x00")
						i_part_start_ext, _ := strconv.Atoi(s_part_start_ext)

						// Obtencion de datos
						s_part_size_ext := string(extended_boot_record.Part_size[:])
						s_part_size_ext = strings.Trim(s_part_size_ext, "\x00")
						i_part_size_ext, _ := strconv.Atoi(s_part_size_ext)

						// Obtencion de datos
						s_part_size_mbr := string(master_boot_record.Mbr_partition[num_extendida].Part_size[:])
						s_part_size_mbr = strings.Trim(s_part_size_mbr, "\x00")
						i_part_size_mbr, _ := strconv.Atoi(s_part_size_mbr)

						// Obtencion de datos
						s_part_start_mbr := string(master_boot_record.Mbr_partition[num_extendida].Part_start[:])
						s_part_start_mbr = strings.Trim(s_part_start_mbr, "\x00")
						i_part_start_mbr, _ := strconv.Atoi(s_part_start_mbr)

						espacio_necesario := i_part_start_ext + i_part_size_ext + size_bytes

						if espacio_necesario <= (i_part_size_mbr + i_part_start_mbr) {
							copy(extended_boot_record.Part_next[:], strconv.Itoa(i_part_start_ext+i_part_size_ext))

							// Escribo el nedxto del ultimo ebr
							pos_actual, _ = f.Seek(0, io.SeekCurrent)
							ebr_byte := struct_a_bytes(extended_boot_record)
							// Escribo el next del ultimo EBR
							f.Seek(int64(int(pos_actual)-len(ebr_byte)), io.SeekStart)
							f.Write(ebr_byte)

							// Escribo el nuevo EBR
							f.Seek(int64(i_part_start_ext+i_part_size_ext), io.SeekStart)
							copy(extended_boot_record.Part_status[:], "0")
							copy(extended_boot_record.Part_fit[:], aux_fit)
							// Posicion actual del archivo
							pos_actual, _ = f.Seek(0, io.SeekCurrent)
							copy(extended_boot_record.Part_start[:], strconv.Itoa(int(pos_actual)))
							copy(extended_boot_record.Part_size[:], strconv.Itoa(size_bytes))
							copy(extended_boot_record.Part_next[:], "-1")
							copy(extended_boot_record.Part_name[:], nombre)
							ebr_byte = struct_a_bytes(extended_boot_record)
							f.Write(ebr_byte)

							salida_comando += "[SUCCES] La Particion logica fue creada con exito!\\n"
						} else {
							salida_comando += "[ERROR] La particion logica a crear excede el espacio disponible de la particion extendida...\\n"
						}
					}
				} else {
					salida_comando += "[ERROR] No se puede crear una particion logica si no hay una extendida...\\n"
				}
			} else {
				salida_comando += "[ERROR] Ya existe una particion con ese nombre...\\n"
			}
		} else {
			salida_comando += "[ERROR] el disco se encuentra vacio...\\n"
		}
		f.Close()
	}
}

// Verifica si el nombre de la particion esta disponible
func existe_particion(direccion string, nombre string) bool {
	extendida := -1
	mbr_empty := MBR{}
	ebr_empty := EBR{}
	var empty [100]byte

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	if err == nil {
		// Procedo a leer el archivo

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		// Si el disco esta creado
		if master_boot_record.Mbr_tamano != empty {
			s_part_name := ""
			s_part_type := ""

			// Recorro las 4 particiones
			for i := 0; i < 4; i++ {
				// Antes de comparar limpio la cadena
				s_part_name = string(master_boot_record.Mbr_partition[i].Part_name[:])
				s_part_name = strings.Trim(s_part_name, "\x00")

				// Verifico si ya existe una particion con ese nombre
				if s_part_name == nombre {
					f.Close()
					return true
				}

				// Antes de comparar limpio la cadena
				s_part_type = string(master_boot_record.Mbr_partition[i].Part_type[:])
				s_part_type = strings.Trim(s_part_type, "\x00")

				// Verifico si de tipo extendida
				if s_part_type == "e" {
					extendida = i
				}
			}

			// Si es extendida
			if extendida != -1 {
				// Obtengo el inicio de la particion
				s_part_start := string(master_boot_record.Mbr_partition[extendida].Part_start[:])
				// Le quito los caracteres null
				s_part_start = strings.Trim(s_part_start, "\x00")
				i_part_start, _ := strconv.Atoi(s_part_start)

				// Obtengo el espacio de la partcion
				s_part_size := string(master_boot_record.Mbr_partition[extendida].Part_size[:])
				// Le quito los caracteres null
				s_part_size = strings.Trim(s_part_size, "\x00")
				i_part_size, _ := strconv.Atoi(s_part_size)

				// Calculo del tamaño de struct en bytes
				ebr2 := struct_a_bytes(ebr_empty)
				sstruct := len(ebr2)

				// Lectrura de conjunto de bytes en archivo binario
				lectura := make([]byte, sstruct)
				// Lee a partir del inicio de la particion
				n_leidos, _ := f.Read(lectura)

				// Posicion actual en el archivo
				f.Seek(int64(i_part_start), io.SeekStart)

				// Posicion actual en el archivo
				pos_actual, _ := f.Seek(0, io.SeekCurrent)

				// Lectrura de conjunto de bytes desde el inicio de la particion
				for n_leidos != 0 && (pos_actual < int64(i_part_size+i_part_start)) {
					// Lectrura de conjunto de bytes en archivo binario
					lectura := make([]byte, sstruct)
					// Lee a partir del inicio de la particion
					n_leidos, _ = f.Read(lectura)

					// Posicion actual en el archivo
					pos_actual, _ = f.Seek(0, io.SeekCurrent)

					// Conversion de bytes a struct
					extended_boot_record := bytes_a_struct_ebr(lectura)

					if extended_boot_record.Part_size == empty {
						break
					} else {
						// Antes de comparar limpio la cadena
						s_part_name = string(extended_boot_record.Part_name[:])
						s_part_name = strings.Trim(s_part_name, "\x00")

						// Verifico si ya existe una particion con ese nombre
						if s_part_name == nombre {
							f.Close()
							return true
						}

						// Obtengo el espacio utilizado
						s_part_next := string(extended_boot_record.Part_next[:])
						// Le quito los caracteres null
						s_part_next = strings.Trim(s_part_next, "\x00")

						// Si ya termino
						if s_part_next != "-1" {
							f.Close()
							return false
						}
					}
				}
			}
		} else {
			salida_comando += "[ERROR] el disco se encuentra vacio...\\n"
		}
	}

	f.Close()
	return false
}

// Busca particiones Primarias o Extendidas
func buscar_particion_p_e(direccion string, nombre string) int {
	// Apertura del archivo
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	if err == nil {
		mbr_empty := MBR{}

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		s_part_status := ""
		s_part_name := ""

		// Recorro las 4 particiones
		for i := 0; i < 4; i++ {
			// Antes de comparar limpio la cadena
			s_part_status = string(master_boot_record.Mbr_partition[i].Part_status[:])
			s_part_status = strings.Trim(s_part_status, "\x00")

			if s_part_status != "1" {
				// Antes de comparar limpio la cadena
				s_part_name = string(master_boot_record.Mbr_partition[i].Part_name[:])
				s_part_name = strings.Trim(s_part_name, "\x00")
				if s_part_name == nombre {
					return i
				}
			}

		}
	}

	return -1
}

// Busca particiones Logicas
func buscar_particion_l(direccion string, nombre string) int {
	// Apertura del archivo
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	if err == nil {
		extendida := -1
		mbr_empty := MBR{}

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, io.SeekStart)
		f.Read(lectura)

		// Conversion de bytes a struct
		master_boot_record := bytes_a_struct_mbr(lectura)

		s_part_type := ""

		// Recorro las 4 particiones
		for i := 0; i < 4; i++ {
			// Antes de comparar limpio la cadena
			s_part_type = string(master_boot_record.Mbr_partition[i].Part_type[:])
			s_part_type = strings.Trim(s_part_type, "\x00")

			if s_part_type != "e" {
				extendida = i
				break
			}
		}

		// Si hay extendida
		if extendida != -1 {
			ebr_empty := EBR{}

			ebr2 := struct_a_bytes(ebr_empty)
			sstruct := len(ebr2)

			// Lectrura del archivo binario desde el inicio
			lectura := make([]byte, sstruct)

			s_part_start := string(master_boot_record.Mbr_partition[extendida].Part_start[:])
			s_part_start = strings.Trim(s_part_start, "\x00")
			i_part_start, _ := strconv.Atoi(s_part_start)

			f.Seek(int64(i_part_start), io.SeekStart)

			n_leidos, _ := f.Read(lectura)
			pos_actual, _ := f.Seek(0, io.SeekCurrent)

			s_part_size := string(master_boot_record.Mbr_partition[extendida].Part_start[:])
			s_part_size = strings.Trim(s_part_size, "\x00")
			i_part_size, _ := strconv.Atoi(s_part_size)

			for (n_leidos != 0) && (pos_actual < int64(i_part_start+i_part_size)) {
				n_leidos, _ = f.Read(lectura)
				pos_actual, _ = f.Seek(0, io.SeekCurrent)

				// Conversion de bytes a struct
				extended_boot_record := bytes_a_struct_ebr(lectura)

				s_part_name_ext := string(extended_boot_record.Part_name[:])
				s_part_name_ext = strings.Trim(s_part_name_ext, "\x00")

				ebr_empty_byte := struct_a_bytes(ebr_empty)

				if s_part_name_ext == nombre {
					return int(pos_actual) - len(ebr_empty_byte)
				}
			}
		}
		f.Close()
	}

	return -1
}

// Formatea en EXT2
func formatear_ext2(inicio int, tamano int, direccion string) {
	sb_empty := Super_bloque{}
	sb := struct_a_bytes(sb_empty)

	in_empty := Inodo{}
	in := struct_a_bytes(in_empty)

	ba_empty := Bloque_archivo{}
	ba := struct_a_bytes(ba_empty)

	// Despejo n de la formula tamaño particion
	n := (tamano - len(sb)) / (4 + len(in) + 3*len(ba))
	num_estructuras := n
	num_bloques := 3 * num_estructuras

	Super_bloque := Super_bloque{}

	fecha := time.Now()
	str_fecha := fecha.Format("02/01/2006 15:04:05")

	copy(Super_bloque.S_filesystem_type[:], "2")
	copy(Super_bloque.S_inodes_count[:], strconv.Itoa(int(num_estructuras)))
	copy(Super_bloque.S_blocks_count[:], strconv.Itoa(int(num_bloques)))
	copy(Super_bloque.S_free_blocks_count[:], strconv.Itoa(int(num_bloques-2)))
	copy(Super_bloque.S_free_inodes_count[:], strconv.Itoa(int(num_estructuras-2)))
	copy(Super_bloque.S_mtime[:], str_fecha)
	copy(Super_bloque.S_mnt_count[:], "0")
	copy(Super_bloque.S_magic[:], "0xEF53")
	copy(Super_bloque.S_inode_size[:], strconv.Itoa(len(in)))
	copy(Super_bloque.S_block_size[:], strconv.Itoa(len(ba)))
	copy(Super_bloque.S_firts_ino[:], "2")
	copy(Super_bloque.S_first_blo[:], "2")
	copy(Super_bloque.S_bm_inode_start[:], strconv.Itoa(inicio+len(sb)))
	copy(Super_bloque.S_bm_block_start[:], strconv.Itoa(inicio+len(sb)+int(num_estructuras)))
	copy(Super_bloque.S_inode_start[:], strconv.Itoa(inicio+len(sb)+int(num_estructuras)+int(num_bloques)))
	copy(Super_bloque.S_block_start[:], strconv.Itoa(inicio+len(sb)+int(num_estructuras)+int(num_bloques)+len(in)+int(num_estructuras)))

	inodo := Inodo{}
	bloque := Bloque_carpeta{}

	// Libre
	buffer := "0"
	// Usado o archivo
	buffer2 := "1"
	// Carpeta
	buffer3 := "2"

	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	if err == nil {
		// Super Bloque
		f.Seek(int64(inicio), io.SeekStart)
		super_bloque_byte := struct_a_bytes(Super_bloque)
		f.Write(super_bloque_byte)

		s_bm_inode_start := string(Super_bloque.S_bm_inode_start[:])
		s_bm_inode_start = strings.Trim(s_bm_inode_start, "\x00")
		i_bm_inode_start, _ := strconv.Atoi(s_bm_inode_start)

		for i := 0; i < num_estructuras; i++ {
			f.Seek(int64(i_bm_inode_start+i), io.SeekStart)
			f.Write([]byte(buffer))
		}

		f.Seek(int64(i_bm_inode_start), io.SeekStart)
		f.Write([]byte(buffer2))
		f.Write([]byte(buffer2))

		s_bm_block_start := string(Super_bloque.S_bm_block_start[:])
		s_bm_block_start = strings.Trim(s_bm_block_start, "\x00")
		i_bm_block_start, _ := strconv.Atoi(s_bm_block_start)

		for i := 0; i < num_bloques; i++ {
			f.Seek(int64(i_bm_block_start+i), io.SeekStart)
			f.Write([]byte(buffer))
		}

		// Marcando el Bitmap de Inodos para la carpeta "/" y el archivo "users.txt"
		f.Seek(int64(i_bm_block_start), io.SeekStart)
		f.Write([]byte(buffer2))
		f.Write([]byte(buffer3))

		/* Inodo para la carpeta "/" */
		copy(inodo.I_uid[:], "1")
		copy(inodo.I_gid[:], "1")
		copy(inodo.I_size[:], "0")
		copy(inodo.I_atime[:], str_fecha)
		copy(inodo.I_ctime[:], str_fecha)
		copy(inodo.I_mtime[:], str_fecha)
		copy(inodo.I_block[0:1], "0")

		// Nodos libres
		for i := 1; i < 15; i++ {
			copy(inodo.I_block[i:i+1], "$")
		}

		// 1 = archivo o 0 = Carpeta
		copy(inodo.I_type[:], "0")
		copy(inodo.I_perm[:], "664")

		s_inode_start := string(Super_bloque.S_inode_start[:])
		s_inode_start = strings.Trim(s_inode_start, "\x00")
		i_inode_start, _ := strconv.Atoi(s_inode_start)

		f.Seek(int64(i_inode_start), io.SeekStart)
		inodo_byte := struct_a_bytes(inodo)
		f.Write(inodo_byte)

		/* Bloque Para Carpeta "/" */
		copy(bloque.B_content[0].B_name[:], ".")
		copy(bloque.B_content[0].B_inodo[:], "0")

		// Directorio Padre
		copy(bloque.B_content[1].B_name[:], "..")
		copy(bloque.B_content[1].B_inodo[:], "0")

		// Nombre de la carpeta o archivo
		copy(bloque.B_content[2].B_name[:], "users.txt")
		copy(bloque.B_content[2].B_inodo[:], "1")
		copy(bloque.B_content[3].B_name[:], ".")
		copy(bloque.B_content[3].B_inodo[:], "-1")

		s_block_start := string(Super_bloque.S_block_start[:])
		s_block_start = strings.Trim(s_block_start, "\x00")
		i_block_start, _ := strconv.Atoi(s_block_start)

		f.Seek(int64(i_block_start), io.SeekStart)
		bloque_byte := struct_a_bytes(bloque)
		f.Write(bloque_byte)

		/* Inodo Para "users.txt" */
		copy(inodo.I_uid[:], "1")
		copy(inodo.I_gid[:], "1")
		copy(inodo.I_size[:], "29")
		copy(inodo.I_atime[:], str_fecha)
		copy(inodo.I_ctime[:], str_fecha)
		copy(inodo.I_mtime[:], str_fecha)
		copy(inodo.I_block[0:1], "1")

		// Nodos libres
		for i := 1; i < 15; i++ {
			copy(inodo.I_block[i:i+1], "$")
		}

		copy(inodo.I_type[:], "1")
		copy(inodo.I_perm[:], "755")

		f.Seek(int64(i_inode_start+len(in)), io.SeekStart)
		inodo_byte = struct_a_bytes(inodo)
		f.Write(inodo_byte)

		/* Bloque Para "users.txt" */
		archivo := Bloque_archivo{}

		for i := 0; i < 100; i++ {
			copy(archivo.B_content[i:i+1], "0")
		}

		bc_empty := Bloque_carpeta{}
		bc := struct_a_bytes(bc_empty)

		// GID, TIPO, GRUPO
		// UID, TIPO, GRUPO, USUARIO, CONTRASEÑA
		copy(archivo.B_content[:], "1,G,root\n1,U,root,root,123\n")
		f.Seek(int64(i_block_start+len(bc)), io.SeekStart)
		archivo_byte := struct_a_bytes(archivo)
		f.Write(archivo_byte)

		salida_comando += "[SUCCES] El Disco se formateo en el sistema EXT2 con exito!\\n"
		f.Close()
	}
}

// Reporte DISK
func graficar_disk(direccion string, destino string, extension string) {
	mbr_empty := MBR{}
	var empty [100]byte

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	// Calculo del tamaño de struct en bytes
	mbr2 := struct_a_bytes(mbr_empty)
	sstruct := len(mbr2)

	// Lectrura del archivo binario desde el inicio
	lectura := make([]byte, sstruct)
	f.Seek(0, io.SeekStart)
	f.Read(lectura)

	// Conversion de bytes a struct
	master_boot_record := bytes_a_struct_mbr(lectura)

	if master_boot_record.Mbr_tamano != empty {
		if err == nil {
			graphDot += "digraph G{\\n\\n"
			graphDot += "  tbl [\\n    shape=box\\n    label=<\\n"
			graphDot += "     <table border='0' cellborder='2' width='600' height=\\\"150\\\" color='dodgerblue1'>\\n"
			graphDot += "     <tr>\\n"
			graphDot += "     <td height='150' width='110'> MBR </td>\\n"

			// Obtengo el espacio utilizado
			s_mbr_tamano := string(master_boot_record.Mbr_tamano[:])
			// Le quito los caracteres null
			s_mbr_tamano = strings.Trim(s_mbr_tamano, "\x00")
			i_mbr_tamano, _ := strconv.Atoi(s_mbr_tamano)
			total := i_mbr_tamano

			var espacioUsado float64
			espacioUsado = 0

			for i := 0; i < 4; i++ {
				// Obtengo el espacio utilizado
				s_part_s := string(master_boot_record.Mbr_partition[i].Part_size[:])
				// Le quito los caracteres null
				s_part_s = strings.Trim(s_part_s, "\x00")
				i_part_s, _ := strconv.Atoi(s_part_s)

				parcial := i_part_s

				// Obtengo el espacio utilizado
				s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
				// Le quito los caracteres null
				s_part_start = strings.Trim(s_part_start, "\x00")

				if s_part_start != "-1" {
					var porcentaje_real float64
					porcentaje_real = (float64(parcial) * 100) / float64(total)
					var porcentaje_aux float64
					porcentaje_aux = (porcentaje_real * 500) / 100

					espacioUsado += porcentaje_real

					// Obtengo el espacio utilizado
					s_part_status := string(master_boot_record.Mbr_partition[i].Part_status[:])
					// Le quito los caracteres null
					s_part_status = strings.Trim(s_part_status, "\x00")

					if s_part_status != "1" {
						// Obtengo el espacio utilizado
						s_part_type := string(master_boot_record.Mbr_partition[i].Part_type[:])
						// Le quito los caracteres null
						s_part_type = strings.Trim(s_part_type, "\x00")

						if s_part_type == "p" {
							graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Primaria <br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"

							if i != 3 {
								// Obtengo el espacio utilizado
								s_part_s = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_s = strings.Trim(s_part_s, "\x00")
								i_part_s, _ = strconv.Atoi(s_part_s)

								// Obtengo el espacio utilizado
								s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								p1 := i_part_start + i_part_s

								// Obtengo el espacio utilizado
								s_part_start = string(master_boot_record.Mbr_partition[i+1].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ = strconv.Atoi(s_part_start)

								p2 := i_part_start

								if s_part_start != "-1" {
									if (p2 - p1) != 0 {
										fragmentacion := p2 - p1
										porcentaje_real = float64(fragmentacion) * 100 / float64(total)
										porcentaje_aux = (porcentaje_real * 500) / 100

										graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Libre<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
									}
								}
							} else {
								// Obtengo el espacio utilizado
								s_part_s = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_s = strings.Trim(s_part_s, "\x00")
								i_part_s, _ = strconv.Atoi(s_part_s)

								// Obtengo el espacio utilizado
								s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								p1 := i_part_start + i_part_s

								mbr_empty_byte := struct_a_bytes(mbr_empty)
								mbr_size := total + len(mbr_empty_byte)

								// Si esta libre
								if (mbr_size - p1) != 0 {
									libre := (float64(mbr_size) - float64(p1)) + float64(len(mbr_empty_byte))
									porcentaje_real = (float64(libre) * 100) / float64(total)
									porcentaje_aux = (porcentaje_real * 500) / 100
									graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Libre<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
								}
							}
						} else {
							// Si es extendida
							graphDot += "     <td  height='200' width='" + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + "'>\\n     <table border='0'  height='200' WIDTH='" + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + "' cellborder='1'>\\n"
							graphDot += "     <tr>  <td height='60' colspan='15'>Extendida</td>  </tr>\\n     <tr>\\n"

							// Obtengo el espacio utilizado
							s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
							// Le quito los caracteres null
							s_part_start = strings.Trim(s_part_start, "\x00")
							i_part_start, _ := strconv.Atoi(s_part_start)

							f.Seek(int64(i_part_start), io.SeekStart)

							ebr_empty := EBR{}

							// Calculo del tamaño de struct en bytes
							ebr2 := struct_a_bytes(ebr_empty)
							sstruct := len(ebr2)

							// Lectrura del archivo binario desde el inicio
							lectura := make([]byte, sstruct)
							f.Read(lectura)

							// Conversion de bytes a struct
							extended_boot_record := bytes_a_struct_ebr(lectura)

							// Obtengo el espacio utilizado
							s_part_size := string(extended_boot_record.Part_size[:])
							// Le quito los caracteres null
							s_part_size = strings.Trim(s_part_size, "\x00")
							i_part_size, _ := strconv.Atoi(s_part_size)

							if i_part_size != 0 {
								// Obtengo el espacio utilizado
								s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								f.Seek(int64(i_part_start), io.SeekStart)

								band := true

								// Obtengo el espacio utilizado
								s_part_s = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_s = strings.Trim(s_part_s, "\x00")
								i_part_s, _ = strconv.Atoi(s_part_s)

								// Obtengo el espacio utilizado
								s_part_start = string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ = strconv.Atoi(s_part_start)

								for band {
									// Calculo del tamaño de struct en bytes
									ebr2 := struct_a_bytes(ebr_empty)
									sstruct := len(ebr2)

									// Lectrura del archivo binario desde el inicio
									lectura := make([]byte, sstruct)
									f.Seek(0, io.SeekStart)
									n, _ := f.Read(lectura)

									// Posicion actual en el archivo
									pos_actual, _ := f.Seek(0, io.SeekCurrent)

									if n != 0 && pos_actual < int64(i_part_start)+int64(i_part_s) {
										band = false
										break
									}

									// Obtengo el espacio utilizado
									s_part_s = string(extended_boot_record.Part_size[:])
									// Le quito los caracteres null
									s_part_s = strings.Trim(s_part_s, "\x00")
									i_part_s, _ = strconv.Atoi(s_part_s)

									parcial = i_part_start
									porcentaje_real = float64(parcial) * 100 / float64(total)

									if porcentaje_real != 0 {
										// Obtengo el espacio utilizado
										s_part_status = string(extended_boot_record.Part_status[:])
										// Le quito los caracteres null
										s_part_status = strings.Trim(s_part_status, "\x00")

										if s_part_status != "1" {
											graphDot += "     <td height='140'>EBR</td>\\n"
											graphDot += "     <td height='140'>Logica<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
										} else {
											// Espacio no asignado
											graphDot += "      <td height='150'>Libre 1 <br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
										}

										// Obtengo el espacio utilizado
										s_part_next := string(extended_boot_record.Part_next[:])
										// Le quito los caracteres null
										s_part_next = strings.Trim(s_part_next, "\x00")
										i_part_next, _ := strconv.Atoi(s_part_next)

										if i_part_next == -1 {
											// Obtengo el espacio utilizado
											s_part_start := string(extended_boot_record.Part_start[:])
											// Le quito los caracteres null
											s_part_start = strings.Trim(s_part_start, "\x00")
											i_part_start, _ := strconv.Atoi(s_part_start)

											// Obtengo el espacio utilizado
											s_part_size := string(extended_boot_record.Part_size[:])
											// Le quito los caracteres null
											s_part_size = strings.Trim(s_part_size, "\x00")
											i_part_size, _ := strconv.Atoi(s_part_size)

											// Obtengo el espacio utilizado
											s_part_start_mbr := string(master_boot_record.Mbr_partition[i].Part_start[:])
											// Le quito los caracteres null
											s_part_start_mbr = strings.Trim(s_part_start_mbr, "\x00")
											i_part_start_mbr, _ := strconv.Atoi(s_part_start_mbr)

											// Obtengo el espacio utilizado
											s_part_s_mbr := string(master_boot_record.Mbr_partition[i].Part_size[:])
											// Le quito los caracteres null
											s_part_s_mbr = strings.Trim(s_part_s_mbr, "\x00")
											i_part_s_mbr, _ := strconv.Atoi(s_part_s_mbr)

											parcial = (i_part_start_mbr + i_part_s_mbr) - (i_part_size + i_part_start)
											porcentaje_real = (float64(parcial) * 100) / float64(total)

											if porcentaje_real != 0 {
												graphDot += "     <td height='150'>Libre 2<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
											}
											break

										} else {
											// Obtengo el espacio utilizado
											s_part_next := string(extended_boot_record.Part_next[:])
											// Le quito los caracteres null
											s_part_next = strings.Trim(s_part_next, "\x00")
											i_part_next, _ := strconv.Atoi(s_part_next)

											f.Seek(int64(i_part_next), io.SeekStart)
										}
									}

								}
							} else {
								graphDot += "     <td height='140'> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
							}
							graphDot += "     </tr>\\n     </table>\\n     </td>\\n"

							// Verifica que no haya espacio fragemntado
							if i != 3 {
								// Obtengo el espacio utilizado
								s_part_s = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_s = strings.Trim(s_part_s, "\x00")
								i_part_s, _ = strconv.Atoi(s_part_s)

								// Obtengo el espacio utilizado
								s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								p1 := i_part_start + i_part_s

								// Obtengo el espacio utilizado
								s_part_start = string(master_boot_record.Mbr_partition[i+1].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ = strconv.Atoi(s_part_start)

								p2 := i_part_start

								if s_part_start != "-1" {
									if (p2 - p1) != 0 {
										fragmentacion := p2 - p1
										porcentaje_real = float64(fragmentacion) * 100 / float64(total)
										porcentaje_aux = (porcentaje_real * 500) / 100

										graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Libre<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
									}
								}
							} else {
								// Obtengo el espacio utilizado
								s_part_s = string(master_boot_record.Mbr_partition[i].Part_size[:])
								// Le quito los caracteres null
								s_part_s = strings.Trim(s_part_s, "\x00")
								i_part_s, _ = strconv.Atoi(s_part_s)

								// Obtengo el espacio utilizado
								s_part_start := string(master_boot_record.Mbr_partition[i].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								p1 := i_part_start + i_part_s

								mbr_empty_byte := struct_a_bytes(mbr_empty)
								mbr_size := total + len(mbr_empty_byte)

								// Si esta libre
								if (mbr_size - p1) != 0 {
									libre := (float64(mbr_size) - float64(p1)) + float64(len(mbr_empty_byte))
									porcentaje_real = (float64(libre) * 100) / float64(total)
									porcentaje_aux = porcentaje_real * 500 / 100
									graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Libre<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
								}
							}
						}
					} else {
						graphDot += "     <td height='200' width='" + strconv.FormatFloat(porcentaje_aux, 'g', 3, 64) + "'>Libre<br/> " + strconv.FormatFloat(porcentaje_real, 'g', 3, 64) + " por ciento del Disco </td>\\n"
					}
				}
			}

			graphDot += "     </tr> \\n     </table>        \\n>];\\n\\n}"
			salida_comando += graphDot
		} else {
			salida_comando += "[ERROR] El disco no fue encontrado...\\n"
		}
	} else {
		salida_comando += "[ERROR] Disco vacio...\\n"
	}
}

// Codifica de Struct a []Bytes
func struct_a_bytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)

	// ERROR
	if err != nil && err != io.EOF {
		salida_comando += "[ERROR] Al codificar de struct a bytes \n"
	}

	return buf.Bytes()
}

// Decodifica de [] Bytes a Struct
func bytes_a_struct_mbr(s []byte) MBR {
	p := MBR{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)

	// ERROR
	if err != nil && err != io.EOF {
		salida_comando += "[ERROR] Al decodificar a MBR\\n"
	}

	return p
}

// Decodifica de [] Bytes a Struct
func bytes_a_struct_ebr(s []byte) EBR {
	p := EBR{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)

	// ERROR
	if err != nil && err != io.EOF {
		salida_comando += "[ERROR] AL decodificar a EBR\n"
	}

	return p
}
