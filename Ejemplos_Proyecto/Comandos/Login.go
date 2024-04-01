package Comandos

import (
	"MIA_Proyecto2_201807079/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

type UsuarioActivo struct {
	User     string
	Password string
	Id       string
	Uid      int
	Gid      int
}

var Logged UsuarioActivo

func ValidarDatosLOGIN(context []string) bool {
	id := ""
	user := ""
	pass := ""

	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "id") {
			id = tk[1]
		} else if Comparar(tk[0], "usuario") {
			user = tk[1]
		} else if Comparar(tk[0], "password") {
			pass = tk[1]
		}
	}
	if id == "" || user == "" || pass == "" {
		Error("LOGIN", "Se necesitan parámetros obligatorios para el comando LOGIN.")
		return false
	}
	return sesionActiva(user, pass, id)
}

func sesionActiva(u string, p string, id string) bool {
	var path string
	partition := GetMount("LOGIN", id, &path)
	if string(partition.Part_status) == "0" {
		Error("LOGIN", "No se encontró la partición montada con el id: "+id)
		return false
	}
	//file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("LOGIN", "No se ha encontrado el disco.")
		return false
	}

	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("LOGIN", "Error al leer el archivo")
		return false
	}
	inode := Structs.NewInodos()
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("LOGIN", "Error al leer el archivo")
		return false
	}

	var fb Structs.BloquesArchivos
	txt := ""
	for bloque := 1; bloque < 16; bloque++ {
		if inode.I_block[bloque-1] == -1 {
			break
		}
		file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(bloque-1), 0)

		data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)

		if err_ != nil {
			Error("LOGIN", "Error al leer el archivo")
			return false
		}

		for i := 0; i < len(fb.B_content); i++ {
			if fb.B_content[i] != 0 {
				txt += string(fb.B_content[i])
			}
		}
	}

	vctr := strings.Split(txt, "\n")
	for i := 0; i < len(vctr)-1; i++ {
		linea := vctr[i]
		if linea[2] == 'U' || linea[2] == 'u' {
			in := strings.Split(linea, ",")
			if Comparar(in[3], u) && Comparar(in[4], p) && in[0] != "0" {
				idGrupo := "0"
				existe := false
				for j := 0; j < len(vctr)-1; j++ {
					line := vctr[j]
					if (line[2] == 'G' || line[2] == 'g') && line[0] != '0' {
						inG := strings.Split(line, ",")
						if inG[2] == in[2] {
							idGrupo = inG[0]
							existe = true
							break
						}
					}
				}
				if !existe {
					Error("Login", "No se encontró el grupo \""+in[2]+"\".")
					return false
				}

				Mensaje("LOGIN", "logueado correctamente")
				fmt.Println("\t\t¡BIENVENIDO " + u + "! :D")
				Logged.Id = id
				Logged.User = u
				Logged.Password = p
				Logged.Uid, _ = strconv.Atoi(in[0])
				Logged.Gid, _ = strconv.Atoi(idGrupo)
				return true
			}
		}
	}
	Error("LOGIN", "No se encontró el usuario "+u)
	return false
}

func CerrarSesion() bool {
	Mensaje("LOGOUT", "¡Adiós "+Logged.User+", espero volver a verte! :D")
	Logged = UsuarioActivo{}
	return false
}
