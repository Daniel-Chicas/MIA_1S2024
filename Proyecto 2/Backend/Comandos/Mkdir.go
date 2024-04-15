package Comandos

import (
	"MIA_Proyecto2_201807079/Structs"
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"time"
	"unsafe"
)

func ValidarDatosMKDIR(context []string, particion Structs.Particion, pth string) {
	path := ""
	p := false
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = tk[1]
		} else if Comparar(tk[0], "p") {
			p = true
		}
	}

	if path == "" {
		Error("MKDIR", "Se necesitan parametros obligatorio para crear un directorio.")
		return
	}
	tmp := GetPath(path)
	mkdir(tmp, p, particion, pth)
}

func mkdir(path []string, p bool, particion Structs.Particion, pth string) {
	copia := path
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	//file, err := os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
		return
	}
	file.Seek(particion.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}

	var newf string
	if len(path) == 0 {
		Error("MKDIR", "No se ha brindado una path válida")
		return
	}
	var past int64
	var bi int64
	var bb int64
	fnd := false
	inodetmp := Structs.NewInodos()
	foldertmp := Structs.NewBloquesCarpetas()

	newf = path[len(path)-1]
	var father int64

	var aux []string
	for i := 0; i < len(path); i++ {
		aux = append(aux, path[i])
	}
	path = aux
	var stack string

	for v := 0; v < len(path)-1; v++ {
		fnd = false
		for i := 0; i < 16; i++ {
			if i < 16 {
				if inode.I_block[i] != -1 {
					folder = Structs.NewBloquesCarpetas()
					file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &folder)
					if err_ != nil {
						Error("MKDIR", "Error al leer el archivo")
						return
					}

					for j := 0; j < 4; j++ {
						nombreCarpeta := ""
						for nam := 0; nam < len(folder.B_content[j].B_name); nam++ {
							if folder.B_content[j].B_name[nam] == 0 {
								continue
							}
							nombreCarpeta += string(folder.B_content[j].B_name[nam])
						}
						if Comparar(nombreCarpeta, path[v]) {
							stack += "/" + path[v]
							fnd = true
							father = folder.B_content[j].B_inodo
							inode = Structs.NewInodos()
							file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*folder.B_content[j].B_inodo, 0)

							data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &inode)

							if err_ != nil {
								Error("MKDIR", "Error al leer el archivo")
								return
							}
							if inode.I_uid != int64(Logged.Uid) {
								Error("MKDIR", "No tiene permisos para crear carpetas en este directorio.")
								return
							}

							break

						}
					}

				} else {
					break
				}
			}
		}
		if !fnd {
			if p {
				stack += "/" + path[v]
				mkdir(GetPath(stack), false, particion, pth)
				file.Seek(spr.S_inode_start, 0)

				data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
				buffer = bytes.NewBuffer(data)
				err_ = binary.Read(buffer, binary.BigEndian, &inode)

				if err_ != nil {
					Error("MKDIR", "Error al leer el archivo")
					return
				}
				if v == len(path)-2 {
					stack += "/" + path[v+1]

					mkdir(GetPath(stack), false, particion, pth)
					file.Seek(spr.S_inode_start, 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &inode)
					if err_ != nil {
						Error("MKDIR", "Error al leer el archivo")
						return
					}
					return

				}
			} else {
				direccion := ""
				for i := 0; i < len(path); i++ {
					direccion += "/" + path[i]
				}
				Error("MKDIR", "No se pudo crear el directorio: "+direccion+", no existen directorios.")
				return
			}
		}
	}

	fnd = false
	for i := 0; i < 16; i++ {
		if inode.I_block[i] != -1 {

			if i < 16 {
				folderAux := folder
				file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0)
				data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
				buffer = bytes.NewBuffer(data)
				err_ = binary.Read(buffer, binary.BigEndian, &folder)
				if err_ != nil {
					Error("MKDIR", "Error al leer el archivo")
					return
				}
				nameAux1 := ""
				for nam := 0; nam < len(folder.B_content[2].B_name); nam++ {
					if folder.B_content[2].B_name[nam] == 0 {
						continue
					}
					nameAux1 += string(folder.B_content[2].B_name[nam])
				}
				nameAux2 := ""
				for nam := 0; nam < len(folderAux.B_content[2].B_name); nam++ {
					if folderAux.B_content[2].B_name[nam] == 0 {
						continue
					}
					nameAux2 += string(folderAux.B_content[2].B_name[nam])
				}
				padre := ""
				for k := 0; k < len(path); k++ {
					if k >= 1 {
						padre = path[k-1]
					}
				}

				if padre == nameAux1 {
					continue
				}
				for j := 0; j < 4; j++ {

					if folder.B_content[j].B_inodo == -1 {
						past = inode.I_block[i]
						bi = GetFree(spr, pth, "BI")
						if bi == -1 {
							Error("MKDIR", "No se ha podido crear el directorio")
							return
						}
						bb = GetFree(spr, pth, "BB")
						if bb == -1 {
							Error("MKDIR", "No se ha podido crear el directorio")
							return
						}

						inodetmp.I_uid = int64(Logged.Uid)
						inodetmp.I_gid = int64(Logged.Gid)
						inodetmp.I_size = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))

						fecha := time.Now().String()
						copy(inodetmp.I_atime[:], spr.S_mtime[:])
						copy(inodetmp.I_ctime[:], fecha)
						copy(inodetmp.I_mtime[:], fecha)
						inodetmp.I_type = 0
						inodetmp.I_perm = 664
						inodetmp.I_block[0] = bb

						copy(foldertmp.B_content[0].B_name[:], ".")
						foldertmp.B_content[0].B_inodo = bi
						copy(foldertmp.B_content[1].B_name[:], "..")
						foldertmp.B_content[1].B_inodo = father
						copy(foldertmp.B_content[2].B_name[:], "-")
						copy(foldertmp.B_content[3].B_name[:], "-")

						folder.B_content[j].B_inodo = bi
						copy(folder.B_content[j].B_name[:], newf)
						fnd = true
						i = 20
						break
					}
				}
			}
		} else {
			break
		}
	}

	if !fnd {
		for i := 0; i < 16; i++ {
			if inode.I_block[i] == -1 {
				if i < 16 {
					bi = GetFree(spr, pth, "BI")
					if bi == -1 {
						Error("MKDIR", "No se ha podido crear el directorio")
						return
					}
					past = GetFree(spr, pth, "BB")
					if past == -1 {
						Error("MKDIR", "No se ha podido crear el directorio")
						return
					}

					bb = GetFree(spr, pth, "BB")

					folder = Structs.NewBloquesCarpetas()
					copy(folder.B_content[0].B_name[:], ".")
					folder.B_content[0].B_inodo = bi
					copy(folder.B_content[1].B_name[:], "..")
					folder.B_content[1].B_inodo = father
					folder.B_content[2].B_inodo = bi
					copy(folder.B_content[2].B_name[:], newf)
					copy(folder.B_content[3].B_name[:], "-")

					inodetmp.I_uid = int64(Logged.Uid)
					inodetmp.I_gid = int64(Logged.Gid)
					inodetmp.I_size = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))

					fecha := time.Now().String()
					copy(inodetmp.I_atime[:], spr.S_mtime[:])
					copy(inodetmp.I_ctime[:], fecha)
					copy(inodetmp.I_mtime[:], fecha)
					inodetmp.I_type = 0
					inodetmp.I_perm = 664
					inodetmp.I_block[0] = bb

					copy(foldertmp.B_content[0].B_name[:], ".")
					foldertmp.B_content[0].B_inodo = bi
					copy(foldertmp.B_content[1].B_name[:], "..")
					foldertmp.B_content[1].B_inodo = father
					copy(foldertmp.B_content[2].B_name[:], "-")
					copy(foldertmp.B_content[3].B_name[:], "-")
					file.Close()

					copy(folder.B_content[2].B_name[:], newf)

					inode.I_block[i] = past
					file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
					//file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))
					if err != nil {
						Error("MKDIR", "No se ha encontrado el disco.")
						return
					}

					file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*father, 0)
					var binInodo bytes.Buffer
					binary.Write(&binInodo, binary.BigEndian, inode)
					EscribirBytes(file, binInodo.Bytes())
					file.Close()
					break
				}
			}
		}
	}

	file.Close()

	file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
	//file, err = os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
		return
	}

	file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*bi, 0)
	var binInodeTemp bytes.Buffer
	binary.Write(&binInodeTemp, binary.BigEndian, inodetmp)
	EscribirBytes(file, binInodeTemp.Bytes())

	file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*bb+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*bb, 0) //TODO AGREGARLE LOS 16 ARCHIVOS
	var binFolderTmp bytes.Buffer
	binary.Write(&binFolderTmp, binary.BigEndian, foldertmp)
	EscribirBytes(file, binFolderTmp.Bytes())

	file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*past+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*past, 0)
	var binFolder bytes.Buffer
	binary.Write(&binFolder, binary.BigEndian, folder)
	EscribirBytes(file, binFolder.Bytes())

	updatebm(spr, pth, "BI")
	updatebm(spr, pth, "BB")

	ruta := ""
	for i := 0; i < len(copia); i++ {
		ruta += "/" + copia[i]
	}
	Mensaje("MKDIR", "Se ha creado el directorio "+ruta)
	file.Close()
}

func GetPath(path string) []string {
	var result []string
	if path == "" {
		return result
	}
	aux := strings.Split(path, "/")
	for i := 1; i < len(aux); i++ {
		result = append(result, aux[i])
	}

	return result
}

func GetFree(spr Structs.SuperBloque, pth string, t string) int64 {
	ch := '2'
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
		return -1
	}
	if t == "BI" {
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < int(spr.S_inodes_count); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return -1
			}
			if ch == '0' {
				file.Close()
				return int64(i)
			}
		}
	} else {
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i < int(spr.S_blocks_count); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return -1
			}
			if ch == '0' {
				file.Close()
				return int64(i)
			}
		}
	}
	file.Close()
	return -1
}

func updatebm(spr Structs.SuperBloque, pth string, t string) {
	ch := 'x'
	var num int
	//file, err := os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
		return
	}

	if t == "BI" {
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < int(spr.S_inodes_count); i++ {

			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return
			}

			if ch == '0' {
				num = i
				break
			}
		}
		file.Close()

		file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
		//file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

		if err != nil {
			Error("MKDIR", "No se ha encontrado el disco.")
			return
		}
		zero := '1'
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < num+1; i++ {
			var binarioZero bytes.Buffer
			binary.Write(&binarioZero, binary.BigEndian, zero)
			EscribirBytes(file, binarioZero.Bytes())
		}
		file.Close()

	} else {
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i < int(spr.S_block_start); i++ {

			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return
			}

			if ch == '0' {
				num = i
				break
			}
		}
		file.Close()

		file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
		//file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

		if err != nil {
			Error("MKDIR", "No se ha encontrado el disco.")
			return
		}
		zero := '1'
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i < num+1; i++ {
			var binarioZero bytes.Buffer
			binary.Write(&binarioZero, binary.BigEndian, zero)
			EscribirBytes(file, binarioZero.Bytes())
		}
		file.Close()
	}
	file.Close()
}
