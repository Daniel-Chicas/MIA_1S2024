package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"
)

type Control struct {
	Sigue int32
}

// Profesor representa la estructura del profesor.
type Profesor struct {
	IDProfesor int32
	CUI        [13]byte
	Nombre     [25]byte
	Curso      [25]byte
}

// Estudiante representa la estructura del estudiante.
type Estudiante struct {
	IDEstudiante int32
	CUI          [13]byte
	Nombre       [25]byte
	Carnet       [10]byte
}

var contadorEstudiante = int32(1)
var contadorProfesor = int32(1)

func main() {
	crearArchivo()
	for {
		fmt.Println("********************** MENU PRINCIPAL **********************")
		fmt.Println("\t 1.- Registro Profesor.")
		fmt.Println("\t 2.- Registro Estudiante.")
		fmt.Println("\t 3.- Ver Registros.")
		fmt.Println("\t 4.- Salir.")
		fmt.Println("Elija una opción:")
		var opcion int
		fmt.Print("\t")
		fmt.Scanln(&opcion)
		switch opcion {
		case 1:
			agregarProfesor()
		case 2:
			agregarEstudiante()
		case 3:
			leerEstructuras()
		case 4:
			os.Exit(0)
		default:
			fmt.Println("Entrada incorrecta")
		}
	}
}

// delay espera la cantidad de segundos especificada.
func delay(secs int) {
	for i := (int32(time.Now().Unix()) + int32(secs)); int32(time.Now().Unix()) != i; time.Sleep(time.Second) {
	}
}

// crearArchivo verifica la existencia del archivo binario y lo crea si no existe.
func crearArchivo() {
	fmt.Println("****************** CREAR ARCHIVO BINARIO ******************")
	fmt.Println("Creando archivo...")

	_, err := os.Stat("estructura.bin")

	if os.IsNotExist(err) {
		file, err := os.Create("estructura.bin")
		if err != nil {
			fmt.Println("Error al crear el archivo:", err)
			os.Exit(1)
		}
		file.Close()
		delay(2)
	}
}

// agregarEstudiante permite al usuario ingresar datos para registrar un estudiante en el archivo binario.
func agregarEstudiante() {
	// Agregar código para estudiante
}

// agregarProfesor permite al usuario ingresar datos para registrar un profesor en el archivo binario.
func agregarProfesor() {
	fmt.Println("****************** INSERTAR PROFESOR ******************")
	file, err := os.OpenFile("estructura.bin", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("El archivo no se pudo abrir correctamente.")
		os.Exit(1)
	}
	defer file.Close()

	var prof Profesor
	var tempCUI string
	var tempName string
	var tempCurso string

	fmt.Print("Ingrese CUI del Profesor: ")
	fmt.Scan(&tempCUI)
	fmt.Print("Ingrese Nombre del Profesor: ")
	fmt.Scan(&tempName)
	fmt.Print("Ingrese Curso del Profesor: ")
	fmt.Scan(&tempCurso)
	prof.IDProfesor = contadorProfesor
	contadorProfesor++

	copy(prof.CUI[:], tempCUI)
	copy(prof.Nombre[:], tempName)
	copy(prof.Curso[:], tempCurso)

	cont := Control{1}

	binary.Write(file, binary.BigEndian, &cont)
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, &prof)
	EscribirBytes(file, binario.Bytes())
}

// leerEstructuras lee y muestra en consola los registros almacenados en el archivo binario.
func leerEstructuras() {
	fmt.Println("****************** DATOS GUARDADOS ******************\n")

	file, err := os.Open("estructura.bin")
	if err != nil {
		fmt.Println("El archivo no se pudo abrir correctamente.")
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		var miControl Control
		err := binary.Read(reader, binary.BigEndian, &miControl)
		if err != nil {
			break
		}

		if miControl.Sigue == 0 {
			var est Estudiante
			binary.Read(reader, binary.BigEndian, &est)
			fmt.Println("TIPO: >>>>> ESTUDIANTE <<<<<")
			fmt.Printf("\tID: %d\n", est.IDEstudiante)
			fmt.Printf("\tCUI: %s\n", string(est.CUI[:]))
			fmt.Printf("\tNOMBRE: %s\n", string(est.Nombre[:]))
			fmt.Printf("\tCARNET: %s\n", string(est.Carnet[:]))
		} else if miControl.Sigue == 1 {
			var profe Profesor
			binary.Read(reader, binary.BigEndian, &profe)
			fmt.Println("TIPO: >>>>> PROFESOR <<<<<")
			fmt.Printf("\tID: %d\n", profe.IDProfesor)
			fmt.Printf("\tCUI: %s\n", string(profe.CUI[:]))
			fmt.Printf("\tNOMBRE: %s\n", string(profe.Nombre[:]))
			fmt.Printf("\tCURSO: %s\n", string(profe.Curso[:]))
		}
		fmt.Println("\n")
	}
	var salir string
	fmt.Scanln(&salir)
}

// EscribirBytes es un ayudante para escribir un conjunto de bytes en el archivo.
func EscribirBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}
