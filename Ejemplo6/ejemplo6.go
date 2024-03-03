package main

/*--------------------------Import--------------------------*/

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*--------------------------/Import--------------------------*/

/*--------------------------Structs--------------------------*/

// Master Boot Record
type mbr = struct {
	Mbr_tamano         [100]byte
	Mbr_fecha_creacion [100]byte
	Mbr_dsk_signature  [100]byte
	Dsk_fit            [100]byte
	Mbr_partition      [4]partition
}

// Partition
type partition = struct {
	Part_status [100]byte
	Part_type   [100]byte
	Part_fit    [100]byte
	Part_start  [100]byte
	Part_size   [100]byte
	Part_name   [100]byte
}

// Extended Boot Record
type ebr = struct {
	Part_status [100]byte
	Part_fit    [100]byte
	Part_start  [100]byte
	Part_size   [100]byte
	Part_next   [100]byte
	Part_name   [100]byte
}

// Super Bloque
type super_bloque = struct {
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
type inodo = struct {
	I_uid   [100]byte
	I_gid   [100]byte
	I_size  [100]byte
	I_atime [100]byte
	I_ctime [100]byte
	I_mtime [100]byte
	I_block [100]byte
	I_type  [100]byte
	I_perm  [100]byte
}

// Bloques de Carpetas
type bloque_carpeta = struct {
	B_content [4]cotent
}

// Content
type cotent = struct {
	B_name  [100]byte
	B_inodo [100]byte
}

// Bloques de Archivos
type bloque_archivo = struct {
	B_content [100]byte
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
	fmt.Println("Analizador en GO")
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
	} else if data == "rmdisk" {
		/* RMDISK */
		rmdisk(commandArray)
	} else if data == "fdisk" {
		/* FDISK */
		fdisk(commandArray)
	} else if data == "pause" {
		/* PAUSE */
		pause()
	} else {
		/* ERROR */
		fmt.Println("[ERROR] El comando no fue reconocido...")
	}
}

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
				band_error = true
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

			if val_unit == "k" || val_unit == "m" {
				// Activo la bandera del parametro
				band_unit = true
			} else {
				// Parametro no valido
				fmt.Println("[ERROR] El Valor del parametro -unit no es valido...")
				band_error = true
				break
			}
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
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
					msg_error(err)
				}

				// Se escriben los datos en disco

				// Apertura del archivo
				f, err := os.OpenFile(val_path, os.O_RDWR, 0660)

				// ERROR
				if err != nil {
					msg_error(err)
				} else {
					// Conversion de struct a bytes
					mbr_byte := struct_a_bytes(master_boot_record)

					// Escribo el mbr desde el inicio del archivos
					f.Seek(0, os.SEEK_SET)
					f.Write(mbr_byte)
					f.Close()

					fmt.Println("[SUCCES] El disco fue creado con exito!")

					// Mostrar contenido guardado
					//mostrar_mkdisk(val_path)
				}
			}
		}
	}

	fmt.Println("[MENSAJE] El comando MKDISK aqui finaliza")
}

/* RMDISK */
func rmdisk(commandArray []string) {
	fmt.Println("[MENSAJE] El comando RMDISK aqui inicia")

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
		for band_path {

			// Si existe el archivo binario
			_, e := os.Stat(val_path)

			if e != nil {
				// Si no existe
				if os.IsNotExist(e) {
					fmt.Println("[ERROR] No existe el disco que desea eliminar...")
					band_path = false
				}
			} else {
				// Si existe
				fmt.Print("[MENSAJE] ¿Desea eliminar el disco [S/N]?: ")

				// Obtengo la opcion ingresada por el usuario
				var opcion string
				fmt.Scanln(&opcion)

				// Verificando entrada
				if opcion == "s" || opcion == "S" {

					// Elimina el archivo
					cmd := exec.Command("/bin/sh", "-c", "rm \""+val_path+"\"")
					cmd.Dir = "/"
					_, err := cmd.Output()

					// ERROR
					if err != nil {
						msg_error(err)
					} else {
						fmt.Println("[SUCCES] El Disco fue eliminado!")
					}

					band_path = false
				} else if opcion == "n" || opcion == "N" {
					fmt.Println("[MENSAJE] EL disco no fue eliminado")
					band_path = false
				} else {
					fmt.Println("[ERROR] Opcion no valida intentalo de nuevo...")
				}
			}
		}
	}

	fmt.Println("[MENSAJE] El comando RMDISK aqui finaliza")
}

/* FDISK */
func fdisk(commandArray []string) {
	fmt.Println("[MENSAJE] El comando FDISK aqui inicia")

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
				band_error = true
			}

			// Valido que el tamaño sea positivo
			if val_size < 0 {
				band_error = true
				fmt.Println("[ERROR] El parametro -size es negativo...")
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

			if val_unit == "b" || val_unit == "k" || val_unit == "m" {
				// Activo la bandera del parametro
				band_unit = true
			} else {
				// Parametro no valido
				fmt.Println("[ERROR] El Valor del parametro -unit no es valido...")
				band_error = true
				break
			}
		/* PARAMETRO OBLIGATORIO -> PATH */
		case strings.Contains(data, "path="):
			if band_path {
				fmt.Println("[ERROR] El parametro -path ya fue ingresado...")
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
				fmt.Println("[ERROR] El parametro -type ya fue ingresado...")
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
				fmt.Println("[ERROR] El Valor del parametro -type no es valido...")
				band_error = true
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
				fmt.Println("[ERROR] El Valor del parametro -fit no es valido...")
				band_error = true
				break
			}
		/* PARAMETRO OBLIGATORIO -> NAME */
		case strings.Contains(data, "name="):
			// Valido si el parametro ya fue ingresado
			if band_name {
				fmt.Println("[ERROR] El parametro -name ya fue ingresado...")
				band_error = true
				break
			}

			// Activo la bandera del parametro
			band_name = true

			// Reemplaza comillas
			val_name = strings.Replace(val_data, "\"", "", 2)
		/* PARAMETRO NO VALIDO */
		default:
			fmt.Println("[ERROR] Parametro no valido...")
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
					fmt.Println("[ERROR] El parametro -name no fue ingresado")
				}
			} else {
				fmt.Println("[ERROR] El parametro -path no fue ingresado")
			}
		} else {
			fmt.Println("[ERROR] El parametro -size no fue ingresado")
		}
	}

	fmt.Println("[MENSAJE] El comando FDISK aqui finaliza")
}

/* PAUSE */
func pause() {
	fmt.Print("[MENSAJE] Presiona enter para continuar...")
	fmt.Scanln()
}

/*--------------------------/Comandos--------------------------*/

// Verifica o crea la ruta para el disco duro
func crear_disco(ruta string) {
	aux, err := filepath.Abs(ruta)

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Crea el directiorio de forma recursiva
	cmd1 := exec.Command("/bin/sh", "-c", "sudo mkdir -p '"+filepath.Dir(aux)+"'")
	cmd1.Dir = "/"
	_, err = cmd1.Output()

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Da los permisos al directorio
	cmd2 := exec.Command("/bin/sh", "-c", "sudo chmod -R 777 '"+filepath.Dir(aux)+"'")
	cmd2.Dir = "/"
	_, err = cmd2.Output()

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Verifica si existe la ruta para el archivo
	if _, err := os.Stat(filepath.Dir(aux)); errors.Is(err, os.ErrNotExist) {
		if err != nil {
			fmt.Println("[FAILURE] No se pudo crear el disco...")
		}
	}
}

// Crea la Particion Primaria
func crear_particion_primaria(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := mbr{}
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
		fmt.Println("[ERROR] No existe un disco duro con ese nombre...")
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
		f.Seek(0, os.SEEK_SET)
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

				fmt.Println("[ESPACIO DISPONIBLE] ", espacio_disponible, " Bytes")
				fmt.Println("[ESPACIO NECESARIO] ", size_bytes, " Bytes")

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
								copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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

								copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(i_part_start_ant+i_part_size_ant+1))
							}

							copy(master_boot_record.Mbr_partition[num_particion].Part_size[:], strconv.Itoa(size_bytes))
							copy(master_boot_record.Mbr_partition[num_particion].Part_status[:], "0")
							copy(master_boot_record.Mbr_partition[num_particion].Part_name[:], nombre)

							// Se guarda de nuevo el MBR

							// Conversion de struct a bytes
							mbr_byte := struct_a_bytes(master_boot_record)

							// Se posiciona al inicio del archivo para guardar la informacion del disco
							f.Seek(0, os.SEEK_SET)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion
							s_part_start = string(master_boot_record.Mbr_partition[num_particion].Part_start[:])
							// Le quito los caracteres null
							s_part_start = strings.Trim(s_part_start, "\x00")
							i_part_start, _ := strconv.Atoi(s_part_start)

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start), os.SEEK_SET)

							// Lo llena de unos
							for i := 1; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							fmt.Println("[SUCCES] La Particion primaria fue creada con exito!")
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
								copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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
							f.Seek(0, os.SEEK_SET)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion best
							s_part_start_best = string(master_boot_record.Mbr_partition[best_index].Part_start[:])
							// Le quito los caracteres null
							s_part_start_best = strings.Trim(s_part_start_best, "\x00")
							i_part_start_best, _ = strconv.Atoi(s_part_start_best)

							// Conversion de struct a bytes

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start_best), os.SEEK_SET)

							// Lo llena de unos
							for i := 1; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							fmt.Println("[SUCCES] La Particion primaria fue creada con exito!")
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
								copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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
							f.Seek(0, os.SEEK_SET)
							f.Write(mbr_byte)

							// Obtengo el inicio de la particion best
							s_part_start_worst = string(master_boot_record.Mbr_partition[worst_index].Part_start[:])
							// Le quito los caracteres null
							s_part_start_worst = strings.Trim(s_part_start_worst, "\x00")
							i_part_start_worst, _ = strconv.Atoi(s_part_start_worst)

							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_start_worst), os.SEEK_SET)

							// Lo llena de unos
							for i := 1; i < size_bytes; i++ {
								f.Write([]byte{1})
							}

							fmt.Println("[SUCCES] La Particion primaria fue creada con exito!")
						}
					} else {
						fmt.Println("[ERROR] Ya existe una particion creada con ese nombre...")
					}
				} else {
					fmt.Println("[ERROR] La particion que desea crear excede el espacio disponible...")
				}
			} else {
				fmt.Println("[ERROR] La suma de particiones primarias y extendidas no debe exceder de 4 particiones...")
				fmt.Println("[MENSAJE] Se recomienda eliminar alguna particion para poder crear otra particion primaria o extendida")
			}
		} else {
			fmt.Println("[ERROR] el disco se encuentra vacio...")
		}

		f.Close()
	}
}

// Crea la Particion Extendida
func crear_particion_extendia(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := mbr{}
	ebr_empty := ebr{}
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
		msg_error(err)
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
		f.Seek(0, os.SEEK_SET)
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

			// Si aun no ha creado la extendida
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

					fmt.Println("[ESPACIO DISPONIBLE] ", espacio_disponible, " Bytes")
					fmt.Println("[ESPACIO NECESARIO] ", size_bytes, " Bytes")

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
									copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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

									copy(master_boot_record.Mbr_partition[num_particion].Part_start[:], strconv.Itoa(i_part_start_ant+i_part_size_ant+1))
								}

								copy(master_boot_record.Mbr_partition[num_particion].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[num_particion].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[num_particion].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Escribe en la posicion inicial del archivo
								f.Seek(0, os.SEEK_SET)
								f.Write(mbr_byte)

								// Obtengo el tamaño de la particion
								s_part_start = string(master_boot_record.Mbr_partition[num_particion].Part_start[:])
								// Le quito los caracteres null
								s_part_start = strings.Trim(s_part_start, "\x00")
								i_part_start, _ := strconv.Atoi(s_part_start)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start), os.SEEK_SET)

								extended_boot_record := ebr{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								ebr_empty_byte := struct_a_bytes(ebr_empty)

								// Lo corro una posicion de donde se encuentra
								pos_actual, _ := f.Seek(0, os.SEEK_CUR)
								f.Seek(int64(pos_actual+1), os.SEEK_SET)

								// Lo llena de unos
								for i := 1; i < (size_bytes - int(binary.Size(ebr_empty_byte))); i++ {
									f.Write([]byte{1})
								}

								fmt.Println("[SUCCES] La Particion extendida fue creada con exito!")
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
									copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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

									copy(master_boot_record.Mbr_partition[best_index].Part_start[:], strconv.Itoa(i_part_start_best_ant+i_part_size_best_ant+1))
								}

								copy(master_boot_record.Mbr_partition[best_index].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[best_index].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[best_index].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Se escribe al inicio del archivo
								f.Seek(0, os.SEEK_SET)
								f.Write(mbr_byte)

								// Obtengo el inicio de la particion best
								s_part_start_best = string(master_boot_record.Mbr_partition[best_index].Part_start[:])
								// Le quito los caracteres null
								s_part_start_best = strings.Trim(s_part_start_best, "\x00")
								i_part_start_best, _ = strconv.Atoi(s_part_start_best)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start_best), os.SEEK_SET)

								extended_boot_record := ebr{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start_best)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								// Lo corro una posicion de donde se encuentra
								pos_actual, _ := f.Seek(0, os.SEEK_CUR)
								f.Seek(int64(pos_actual+1), os.SEEK_SET)

								ebr_empty_byte := struct_a_bytes(mbr_empty)

								// Lo llena de unos
								for i := 1; i < (size_bytes - int(binary.Size(ebr_empty_byte))); i++ {
									f.Write([]byte{1})
								}

								fmt.Println("[SUCCES] La Particion extendida fue creada con exito!")
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
									copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(int(binary.Size(mbr_empty_byte))+1))
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

									copy(master_boot_record.Mbr_partition[worst_index].Part_start[:], strconv.Itoa(i_part_start_worst_ant+i_part_size_worst_ant+1))
								}

								copy(master_boot_record.Mbr_partition[worst_index].Part_size[:], strconv.Itoa(size_bytes))
								copy(master_boot_record.Mbr_partition[worst_index].Part_status[:], "0")
								copy(master_boot_record.Mbr_partition[worst_index].Part_name[:], nombre)

								// Se guarda de nuevo el MBR

								// Conversion de struct a bytes
								mbr_byte := struct_a_bytes(master_boot_record)

								// Se escribe desde el inicio del archivo
								f.Seek(0, os.SEEK_SET)
								f.Write(mbr_byte)

								// Obtengo el inicio de la particion best
								s_part_start_worst = string(master_boot_record.Mbr_partition[worst_index].Part_start[:])
								// Le quito los caracteres null
								s_part_start_worst = strings.Trim(s_part_start_worst, "\x00")
								i_part_start_worst, _ = strconv.Atoi(s_part_start_worst)

								// Se posiciona en el inicio de la particion
								f.Seek(int64(i_part_start_worst), os.SEEK_SET)

								extended_boot_record := ebr{}
								copy(extended_boot_record.Part_fit[:], aux_fit)
								copy(extended_boot_record.Part_status[:], "0")
								copy(extended_boot_record.Part_start[:], s_part_start_worst)
								copy(extended_boot_record.Part_size[:], "0")
								copy(extended_boot_record.Part_next[:], "-1")
								copy(extended_boot_record.Part_name[:], "")
								ebr_byte := struct_a_bytes(extended_boot_record)
								f.Write(ebr_byte)

								// Lo corro una posicion de donde se encuentra
								pos_actual, _ := f.Seek(0, os.SEEK_CUR)
								f.Seek(int64(pos_actual+1), os.SEEK_SET)

								ebr_empty_byte := struct_a_bytes(mbr_empty)

								// Lo llena de unos
								for i := 1; i < (size_bytes - int(binary.Size(ebr_empty_byte))); i++ {
									f.Write([]byte{1})
								}

								fmt.Println("[SUCCES] La Particion extendida fue creada con exito!")
							}
						} else {
							fmt.Println("[ERROR] Ya existe una particion creada con ese nombre...")
						}
					} else {
						fmt.Println("[ERROR] La particion que desea crear excede el espacio disponible...")
					}
				} else {
					fmt.Println("[ERROR] La suma de particiones primarias y extendidas no debe exceder de 4 particiones...")
					fmt.Println("[MENSAJE] Se recomienda eliminar alguna particion para poder crear otra particion primaria o extendida")
				}
			} else {
				fmt.Println("[ERROR] Solo puede haber una particion extendida por disco...")
			}
		} else {
			fmt.Println("[ERROR] el disco se encuentra vacio...")
		}
		f.Close()
	}
}

// Crea la Particion Logica
func crear_particion_logica(direccion string, nombre string, size int, fit string, unit string) {
	aux_fit := ""
	aux_unit := ""
	size_bytes := 1024

	mbr_empty := mbr{}
	ebr_empty := ebr{}
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
		fmt.Println("[ERROR] No existe el disco duro con ese nombre...")
	} else {
		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, os.SEEK_SET)
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
					f.Seek(int64(cont), os.SEEK_SET)

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

						fmt.Println("[ESPACIO DISPONIBLE] ", i_part_size, " Bytes")
						fmt.Println("[ESPACIO NECESARIO] ", size_bytes, " Bytes")

						// Si excede el tamaño de la extendida
						if i_part_size < size_bytes {
							fmt.Println("[ERROR] La particion logica a crear excede el espacio disponible de la particion extendida...")
						} else {
							copy(extended_boot_record.Part_status[:], "0")
							copy(extended_boot_record.Part_fit[:], aux_fit)

							// Posicion actual en el archivo
							pos_actual, _ := f.Seek(0, os.SEEK_CUR)

							copy(extended_boot_record.Part_start[:], strconv.Itoa(int(pos_actual)-int(binary.Size(ebr_empty))+1))
							copy(extended_boot_record.Part_size[:], strconv.Itoa(size_bytes))
							copy(extended_boot_record.Part_next[:], "-1")
							copy(extended_boot_record.Part_name[:], nombre)

							// Obtencion de datos
							s_part_start := string(master_boot_record.Mbr_partition[num_extendida].Part_start[:])
							s_part_start = strings.Trim(s_part_start, "\x00")
							i_part_start, _ := strconv.Atoi(s_part_start)

							// Se posiciona en el inicio de la particion
							ebr_byte := struct_a_bytes(extended_boot_record)
							f.Seek(int64(i_part_start), os.SEEK_SET)
							f.Write(ebr_byte)

							fmt.Println("[SUCCES] La Particion logica fue creada con exito!")
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

						espacio_disponible := i_part_size + i_part_start

						fmt.Println("[ESPACIO DISPONIBLE] ", espacio_disponible, " Bytes")
						fmt.Println("[ESPACIO NECESARIO] ", size_bytes, " Bytes")

						// Obtencion de datos
						s_part_next := string(extended_boot_record.Part_next[:])
						s_part_next = strings.Trim(s_part_next, "\x00")
						i_part_next, _ := strconv.Atoi(s_part_next)

						pos_actual, _ := f.Seek(0, os.SEEK_CUR)

						for (i_part_next != -1) && (int(pos_actual) < (i_part_size + i_part_start)) {
							// Se posiciona en el inicio de la particion
							f.Seek(int64(i_part_next), os.SEEK_SET)

							// Calculo del tamaño de struct en bytes
							ebr2 := struct_a_bytes(ebr_empty)
							sstruct := len(ebr2)

							// Lectrura del archivo binario desde el inicio
							lectura := make([]byte, sstruct)
							f.Read(lectura)

							// Posicion actual
							pos_actual, _ = f.Seek(0, os.SEEK_CUR)

							// Conversion de bytes a struct
							extended_boot_record = bytes_a_struct_ebr(lectura)

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

							// Posicion actual del archivo
							pos_actual, _ = f.Seek(0, os.SEEK_CUR)

							// Escribo el next del ultimo EBR
							f.Seek(int64(int(pos_actual)-int(binary.Size(ebr_empty))), os.SEEK_SET)
							ebr_byte := struct_a_bytes(extended_boot_record)
							f.Write(ebr_byte)

							// Escribo el nuevo EBR
							f.Seek(int64(i_part_start_ext+i_part_size_ext), os.SEEK_SET)
							copy(extended_boot_record.Part_status[:], "0")
							copy(extended_boot_record.Part_fit[:], aux_fit)

							// Posicion actual del archivo
							pos_actual, _ = f.Seek(0, os.SEEK_CUR)

							copy(extended_boot_record.Part_start[:], strconv.Itoa(int(pos_actual)))
							copy(extended_boot_record.Part_size[:], strconv.Itoa(size_bytes))
							copy(extended_boot_record.Part_next[:], "-1")
							copy(extended_boot_record.Part_name[:], nombre)

							ebr_byte = struct_a_bytes(extended_boot_record)
							f.Write(ebr_byte)
							fmt.Println("[SUCCES] La Particion logica fue creada con exito!")
						} else {
							fmt.Println("[ERROR] La particion logica a crear excede el espacio disponible de la particion extendida...")
						}
					}
				} else {
					fmt.Println("[ERROR] No se puede crear una particion logica si no hay una extendida...")
				}
			} else {
				fmt.Println("[ERROR] Ya existe una particion con ese nombre...")
			}
		} else {
			fmt.Println("[ERROR] el disco se encuentra vacio...")
		}
		f.Close()
	}
}

// Verifica si el nombre de la particion esta disponible
func existe_particion(direccion string, nombre string) bool {
	extendida := -1
	mbr_empty := mbr{}
	ebr_empty := ebr{}
	var empty [100]byte
	fin := false

	// Abro el archivo para lectura con opcion a modificar
	f, err := os.OpenFile(direccion, os.O_RDWR, 0660)

	// ERROR
	if err == nil {
		// Procedo a leer el archivo

		// Calculo del tamaño de struct en bytes
		mbr2 := struct_a_bytes(mbr_empty)
		sstruct := len(mbr2)

		// Lectrura del archivo binario desde el inicio
		lectura := make([]byte, sstruct)
		f.Seek(0, os.SEEK_SET)
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

				// Posicion actual en el archivo
				f.Seek(int64(i_part_start), os.SEEK_SET)

				// Calculo del tamaño de struct en bytes
				ebr2 := struct_a_bytes(ebr_empty)
				sstruct := len(ebr2)

				// Lectrura de conjunto de bytes desde el inicio de la particion
				for !fin {
					// Lectrura de conjunto de bytes en archivo binario
					lectura := make([]byte, sstruct)
					// Lee a partir del inicio de la particion
					n_leidos, _ := f.Read(lectura)

					// Posicion actual en el archivo
					pos_actual, _ := f.Seek(0, os.SEEK_CUR)

					// Verifico si no ha leido nada y se ha pasado del tamaño de la particion
					if n_leidos == 0 && (pos_actual >= int64(i_part_size+i_part_start)) {
						fin = true
						break
					}

					// Conversion de bytes a struct
					extended_boot_record := bytes_a_struct_ebr(lectura)
					sstruct = len(lectura)

					if extended_boot_record.Part_size == empty {
						fin = true
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
			fmt.Println("[ERROR] el disco se encuentra vacio...")
		}
	}

	f.Close()
	return false
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

// Decodifica de [] Bytes a Struct
func bytes_a_struct_ebr(s []byte) ebr {
	p := ebr{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)

	// ERROR
	if err != nil && err != io.EOF {
		msg_error(err)
	}

	return p
}

// Muestra los datos en el disco
func mostrar_mkdisk(ruta string) {
	var empty [100]byte
	mbr_empty := mbr{}

	fmt.Println("COMANDO MOSTRAR")

	// Apertura de archivo
	f, err := os.OpenFile(ruta, os.O_RDWR, 0660)

	// ERROR
	if err != nil {
		msg_error(err)
	}

	// Calculo del tamaño de struct en bytes
	mbr2 := struct_a_bytes(mbr_empty)
	sstruct := len(mbr2)

	// Lectrura del archivo binario desde el inicio
	lectura := make([]byte, sstruct)
	f.Seek(0, os.SEEK_SET)
	f.Read(lectura)

	// Conversion de bytes a struct
	mbr := bytes_a_struct_mbr(lectura)

	if mbr.Mbr_tamano != empty {
		fmt.Print("Tamaño: ")
		fmt.Println(string(mbr.Mbr_tamano[:]))
		fmt.Print("Fecha: ")
		fmt.Println(string(mbr.Mbr_fecha_creacion[:]))
		fmt.Print("Signature: ")
		fmt.Println(string(mbr.Mbr_dsk_signature[:]))
		fmt.Print("Fit: ")
		fmt.Println(string(mbr.Dsk_fit[:]))
		fmt.Println("Particion 1")
		fmt.Print("Type: ")
		fmt.Println(string(mbr.Mbr_partition[1].Part_type[:]))
	}
	f.Close()
}

/*--------------------------/Metodos o Funciones--------------------------*/
