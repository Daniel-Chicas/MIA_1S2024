package Comandos

import (
	"MIA_Proyecto2_201807079/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

var contadorBloques int
var contadorArchivos int
var bloquesUsados []int64

func ValidarDatosREP(context []string) {
	contadorBloques = 0
	contadorArchivos = 0
	bloquesUsados = []int64{}
	name := ""
	path := ""
	id := ""
	ruta := ""
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = strings.ReplaceAll(tk[1], "\"", "")
		} else if Comparar(tk[0], "name") {
			name = tk[1]
		} else if Comparar(tk[0], "id") {
			id = tk[1]
		} else if Comparar(tk[0], "ruta") {
			ruta = tk[1]
		}
	}

	if name == "" || path == "" || id == "" {
		Error("REP", "Se esperan parámetros obligatorios.")
		return
	}

	if Comparar(name, "DISK") {
		dks(path, id)
	} else if Comparar(name, "TREE") {
		tree(path, id)
	} else if Comparar(name, "FILE") {
		if ruta == "" {
			Error("REP", "Se espera el parámetro ruta.")
			return
		}
		fileR(path, id, ruta)
	} else {
		Error("REP", name+", no es un reporte válido.")
		return
	}
}

func dks(p string, id string) {
	var pth string
	GetMount("REP", id, &pth)

	//file, err := os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	var disk Structs.MBR
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}
	file.Close()

	aux := strings.Split(p, ".")
	if len(aux) > 2 {
		Error("REP", "No se admiten nombres de archivos que contengan punto (.)")
		return
	}
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	partitions := GetParticiones(disk)
	var extended Structs.Particion
	ext := false
	for i := 0; i < 4; i++ {
		if partitions[i].Part_status == '1' {
			if partitions[i].Part_type == "E"[0] || partitions[i].Part_type == "e"[0] {
				ext = true
				extended = partitions[i]
			}
		}
	}

	content := ""
	content = "digraph G{\n rankdir=TB;\n forcelabels= true;\n graph [ dpi = \"600\" ]; \n node [shape = plaintext];\n nodo1 [label = <<table>\n <tr>\n"
	var positions [5]int64
	var positionsii [5]int64
	positions[0] = disk.Mbr_partition_1.Part_start - (1 + int64(unsafe.Sizeof(Structs.MBR{})))
	positions[1] = disk.Mbr_partition_2.Part_start - disk.Mbr_partition_1.Part_start + disk.Mbr_partition_1.Part_size
	positions[2] = disk.Mbr_partition_3.Part_start - disk.Mbr_partition_2.Part_start + disk.Mbr_partition_2.Part_size
	positions[3] = disk.Mbr_partition_4.Part_start - disk.Mbr_partition_3.Part_start + disk.Mbr_partition_3.Part_size
	positions[4] = disk.Mbr_tamano + 1 - disk.Mbr_partition_4.Part_start + disk.Mbr_partition_4.Part_size

	copy(positionsii[:], positions[:])

	logic := 0
	tmplogic := ""
	if ext {
		tmplogic = "<tr>\n"
		auxEbr := Structs.NewEBR()
		//file, err := os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
		file, err = os.Open(strings.ReplaceAll(pth, "\"", ""))

		if err != nil {
			Error("REP", "No se ha encontrado el disco.")
			return
		}

		file.Seek(extended.Part_start, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}
		file.Close()
		var tamGen int64 = 0
		for auxEbr.Part_next != -1 {
			tamGen += auxEbr.Part_size
			res := float64(auxEbr.Part_size) / float64(disk.Mbr_tamano)
			res = res * 100
			tmplogic += "<td>\"EBR\"</td>"
			s := fmt.Sprintf("%.2f", res)
			tmplogic += "<td>\"Logica\n " + s + "% de la partición extendida\"</td>\n"

			resta := float64(auxEbr.Part_next) - (float64(auxEbr.Part_start) + float64(auxEbr.Part_size))
			resta = resta / float64(disk.Mbr_tamano)
			resta = resta * 10000.00
			resta = math.Round(resta) / 100.00
			if resta != 0 {
				s = fmt.Sprintf("%f", resta)
				tmplogic += "<td>\"Logica\n " + s + "% libre de la partición extendida\"</td>\n"
				logic++
			}
			logic += 2
			file, err = os.Open(strings.ReplaceAll(pth, "\"", ""))

			if err != nil {
				Error("REP", "No se ha encontrado el disco.")
				return
			}

			file.Seek(auxEbr.Part_next, 0)
			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
			if err_ != nil {
				Error("REP", "Error al leer el archivo")
				return
			}
			file.Close()
		}
		resta := float64(extended.Part_size) - float64(tamGen)
		resta = resta / float64(disk.Mbr_tamano)
		resta = math.Round(resta * 100)
		if resta != 0 {
			s := fmt.Sprintf("%.2f", resta)
			tmplogic += "<td>\"Libre \n " + s + "% de la partición extendida.\"</td>\n"
			logic++
		}
		tmplogic += "</tr>\n\n"
		logic += 2

	}
	var tamPrim int64
	for i := 0; i < 4; i++ {
		if partitions[i].Part_type == 'E' {
			tamPrim += partitions[i].Part_size
			res := float64(partitions[i].Part_size) / float64(disk.Mbr_tamano)
			res = math.Round(res*10000.00) / 100.00
			s := fmt.Sprintf("%.2f", res)
			content += "<td COLSPAN='" + strconv.Itoa(logic) + "'> Extendida \n" + s + "% del disco</td>\n"
		} else if partitions[i].Part_start != -1 {
			tamPrim += partitions[i].Part_size
			res := float64(partitions[i].Part_size) / float64(disk.Mbr_tamano)
			res = math.Round(res*10000.00) / 100.00
			s := fmt.Sprintf("%.2f", res)
			content += "<td ROWSPAN='2'> Primaria \n" + s + "% del disco</td>\n"
		}
	}

	if tamPrim != 0 {
		libre := disk.Mbr_tamano - tamPrim
		res := float64(libre) / float64(disk.Mbr_tamano)
		res = math.Round(res * 100)
		s := fmt.Sprintf("%.2f", res)
		content += "<td ROWSPAN='2'> Libre \n" + s + "% del disco</td>"

	}
	content += "</tr>\n\n"
	content += tmplogic
	content += "</table>>];\n}\n"

	//CREAR IMAGEN
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	mode := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(mode))
	disco := strings.Split(pth, "/")
	Mensaje("REP", "Reporte tipo DISK del disco "+disco[len(disco)-1]+", creado correctamente.")
}

func tree(p string, id string) {
	var pth string
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	partition := GetMount("REP", id, &pth)

	if partition.Part_start == -1 {
		return
	}

	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}

	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}

	freeI := GetFree(spr, pth, "BI")
	aux := strings.Split(strings.ReplaceAll(p, "\"", ""), ".")
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	content := "digraph G{\n rankdir=LR;\n graph [ dpi = \"600\" ]; \n forcelabels= true;\n node [shape = plaintext];\n"
	for i := 0; i < int(freeI); i++ {
		atime := arregloString(inode.I_atime)
		ctime := arregloString(inode.I_ctime)
		mtime := arregloString(inode.I_mtime)
		content += "inode" + strconv.Itoa(i) + "  [label = <<table>\n" +
			"<tr><td COLSPAN = '2' BGCOLOR=\"#000080\">" +
			"<font color=\"white\">INODO " + strconv.Itoa(i) + "</font>" +
			"</td></tr>\n " +
			"<tr><td BGCOLOR=\"#87CEFA\">NOMBRE</td><td BGCOLOR=\"#87CEFA\" >VALOR</td></tr>\n" +
			"<tr><td>i_uid</td><td>" + strconv.Itoa(int(inode.I_uid)) + "</td></tr>\n" +
			"<tr><td>i_gid</td><td>" + strconv.Itoa(int(inode.I_gid)) + "</td></tr>\n" +
			"<tr><td>i_size</td><td>" + strconv.Itoa(int(inode.I_size)) + "</td></tr>\n" +
			"<tr><td>i_atime</td><td>" + atime + "</td></tr>\n" +
			"<tr><td>i_ctime</td><td>" + ctime + "</td></tr>\n" +
			"<tr><td>i_mtime</td><td>" + mtime + "</td></tr>\n"
		for j := 0; j < 16; j++ {
			content += "<tr>\n<td>i_block_" + strconv.Itoa(j+1) + "</td><td port=\"" + strconv.Itoa(j) + "\">" + strconv.Itoa(int(inode.I_block[j])) + "</td></tr>\n"
		}
		content += "<tr><td>i_type</td><td>" + strconv.Itoa(int(inode.I_type)) + "</td></tr>\n" +
			"<tr><td>i_perm</td><td>" + strconv.Itoa(int(inode.I_perm)) + "</td></tr></table>>];\n"

		if inode.I_type == 0 {

			for j := 0; j < 16; j++ {
				if inode.I_block[j] != -1 {
					bloquesUsados = append(bloquesUsados, inode.I_block[j])
					contadorBloques++
					if existeEnArreglo(bloquesUsados, inode.I_block[j]) == 1 {
						foldertmp := Structs.BloquesCarpetas{}

						file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[j]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[j], 0)
						data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &foldertmp)
						if err_ != nil {
							Error("REP", "Error al leer el archivo")
							return
						}

						if foldertmp.B_content[2].B_inodo == -1 {
							continue
						}
						content += "inode" + strconv.Itoa(i) + ":" + strconv.Itoa(j) + "-> BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inode.I_block[j])) + "\n"

						content += "BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inode.I_block[j])) +
							" [label = <<table><tr><td COLSPAN = '2' BGCOLOR=\"#145A32\">" +
							"<font color=\"white\">BLOCK " + strconv.Itoa(int(inode.I_block[j])) + "</font>" +
							"</td></tr><tr><td BGCOLOR=\"#90EE90\">B_NAME</td>" +
							"<td BGCOLOR=\"#90EE90\" >B_INODO</td></tr>\n"

						for k := 0; k < 4; k++ {
							ctmp := ""
							for name := 0; name < len(foldertmp.B_content[k].B_name); name++ {
								if foldertmp.B_content[k].B_name[name] != 0 {
									ctmp += string(foldertmp.B_content[k].B_name[name])
								}
							}
							content += "<tr>\n<td>" + ctmp + "</td>\n<td port=\"" + strconv.Itoa(k) + "\">" + strconv.Itoa(int(foldertmp.B_content[k].B_inodo)) + "</td>\n</tr>\n"
						}

						content += "</table>>];\n"

						for b := 0; b < 4; b++ {
							if foldertmp.B_content[b].B_inodo != -1 {
								ctmp := ""
								for name := 0; name < len(foldertmp.B_content[b].B_name); name++ {
									if foldertmp.B_content[b].B_name[name] != 0 {
										ctmp += string(foldertmp.B_content[b].B_name[name])
									}
								}
								if ctmp != "." && ctmp != ".." {
									content += "BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inode.I_block[j])) + ":" + strconv.Itoa(b) + " -> inode" + strconv.Itoa(int(foldertmp.B_content[b].B_inodo)) + ";\n"
								}
							}
						}
					}
				}
			}

		} else {
			for j := 0; j < 16; j++ {
				if inode.I_block[j] != -1 {
					if j < 16 {
						var contador int64 = 0
						var posicion int
						for {
							foldertmp := Structs.NewBloquesCarpetas()
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador, 0)
							data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &foldertmp)

							if err_ != nil {
								Error("REP", "Error al leer el archivo")
								return
							}
							salir := false
							for l := 0; l < 4; l++ {
								if foldertmp.B_content[l].B_inodo == inode.I_block[0] {
									posicion = l
									salir = true
									break
								}
							}
							if salir {
								break
							}
							contador++
						}
						if posicion == 2 || posicion == 0 {
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador+int64(unsafe.Sizeof(Structs.BloquesCarpetas{})), 0)
							for k := 0; k < 16; k++ {
								contadorArchivos++
								if inode.I_block[k] == -1 {
									break
								}
								content += "inode" + strconv.Itoa(i) + ":" + strconv.Itoa(k) + "-> FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inode.I_block[k])) + "\n"
								filetmp := Structs.BloquesArchivos{}
								data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
								buffer = bytes.NewBuffer(data)
								err_ = binary.Read(buffer, binary.BigEndian, &filetmp)
								if err_ != nil {
									Error("REP", "Error al leer el archivo")
									return
								}

								contenido := ""
								for arch := 0; arch < len(filetmp.B_content); arch++ {
									if filetmp.B_content[arch] != 0 {
										contenido += string(filetmp.B_content[arch])
									}
								}
								content += "FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inode.I_block[k])) + " [label = <<table >\n<tr><td COLSPAN = '2' BGCOLOR=\"#CCCC00\">FILE " + strconv.Itoa(int(inode.I_block[k])) +
									"</td></tr>\n <tr><td COLSPAN = '2'>" + contenido + "</td></tr>\n</table>>];\n"
							}
						} else if posicion == 3 {
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(16), 0)
							for k := 0; k < 16; k++ {
								contadorArchivos++
								if inode.I_block[k] == -1 {
									break
								}
								content += "inode" + strconv.Itoa(i) + ":" + strconv.Itoa(k) + "-> FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inode.I_block[k])) + "\n"
								filetmp := Structs.BloquesArchivos{}
								data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
								buffer = bytes.NewBuffer(data)
								err_ = binary.Read(buffer, binary.BigEndian, &filetmp)
								if err_ != nil {
									Error("REP", "Error al leer el archivo")
									return
								}

								contenido := ""
								for arch := 0; arch < len(filetmp.B_content); arch++ {
									if filetmp.B_content[arch] != 0 {
										contenido += string(filetmp.B_content[arch])
									}
								}
								content += "FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inode.I_block[k])) + " [label = <<table >\n<tr><td COLSPAN = '2' BGCOLOR=\"#CCCC00\">FILE " + strconv.Itoa(int(inode.I_block[k])) +
									"</td></tr>\n <tr><td COLSPAN = '2'>" + contenido + "</td></tr>\n</table>>];\n"
							}
						}
						break
					}
				} else {
					break
				}
			}
		}
		file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*int64(i+1), 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inode)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}
	}
	file.Close()
	content += "\n\n}\n"
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	mode := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(mode))
	disco := strings.Split(pth, "/")
	Mensaje("REP", "Reporte tipo TREE del disco "+disco[len(disco)-1]+", creado correctamente.")
}

func fileR(p string, id string, ruta string) {

	carpeta := ""
	direccion := strings.Split(p, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(p, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(p)
	} else {
		fileaux.Close()
	}

	var path string
	particion := GetMount("MKDIR", id, &path)
	tmp := GetPath(ruta)
	data := getDataFile(tmp, particion, path)
	b := []byte(data)
	err_ := ioutil.WriteFile(p, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	archivo := strings.Split(ruta, "/")
	Mensaje("REP", "Reporte tipo FILE del archivo  "+archivo[len(archivo)-1]+", creado correctamente.")
}

func arregloString(arreglo [16]byte) string {
	reg := ""
	for i := 0; i < 16; i++ {
		if arreglo[i] != 0 {
			reg += string(arreglo[i])
		}
	}
	return reg
}

func existeEnArreglo(arreglo []int64, busqueda int64) int {
	regresa := 0
	for _, numero := range arreglo {
		if numero == busqueda {
			regresa++
		}
	}
	return regresa
}
