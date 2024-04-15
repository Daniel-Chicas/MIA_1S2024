package Comandos

import (
	"MIA_Proyecto2_201807079/Structs"
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var ExisteDirectorio bool = true

func ValidarDatosMKFILE(context []string, particion Structs.Particion, pth string) {
	ExisteDirectorio = true
	path := ""
	p := false
	size := ""
	cont := ""
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = tk[1]
		} else if Comparar(tk[0], "r") {
			p = true
		} else if Comparar(tk[0], "size") {
			size = tk[1]
		} else if Comparar(tk[0], "cont") {
			cont = tk[1]
		}
	}
	if size != "" {
		tam, err := strconv.Atoi(size)
		if err != nil {
			Error("MKFILE", "Se esperaba un entero para el parámetro size")
			return
		} else if tam < 0 {
			Error("MKFILE", "Se esperaba un número mayor a 0 para el parámetro size")
			return
		}
	}

	if path == "" {
		Error("MKFILE", "Se necesitan parametros obligatorio para crear un directorio.")
		return
	}
	tmp := GetPath(path)
	mkfile(tmp, p, particion, pth)
	if ExisteDirectorio {
		setDataFile(tmp, p, size, cont, particion, pth)
	}
}

func mkfile(path []string, p bool, particion Structs.Particion, pth string) {
	copia := path
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	//file, err := os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKFILE", "No se ha encontrado el disco.")
		return
	}
	file.Seek(particion.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("MKFILE", "Error al leer el archivo")
		return
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("MKFILE", "Error al leer el archivo")
		return
	}

	var newf string
	if len(path) == 0 {
		Error("MKFILE", "No se ha brindado una path válida")
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
						Error("MKFILE", "Error al leer el archivo")
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
								Error("MKFILE", "Error al leer el archivo")
								return
							}
							if inode.I_uid != int64(Logged.Uid) {
								Error("MKFILE", "No tiene permisos para crear carpetas en este directorio.")
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
					Error("MKFILE", "Error al leer el archivo")
					return
				}
				if v == len(path)-2 {
					stack += "/" + path[v+1]

					mkfile(GetPath(stack), false, particion, pth)
					file.Seek(spr.S_inode_start, 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &inode)
					if err_ != nil {
						Error("MKFILE", "Error al leer el archivo")
						return
					}
					return

				}
			} else {
				direccion := ""
				for i := 0; i < len(path); i++ {
					direccion += "/" + path[i]
				}
				Error("MKFILE", "No se pudo crear el directorio: "+direccion+", no existen directorios.")
				ExisteDirectorio = false
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
					Error("MKFILE", "Error al leer el archivo")
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
							Error("MKFILE", "No se ha podido crear el directorio")
							return
						}
						bb = GetFree(spr, pth, "BB")
						if bb == -1 {
							Error("MKFILE", "No se ha podido crear el directorio")
							return
						}

						inodetmp.I_uid = int64(Logged.Uid)
						inodetmp.I_gid = int64(Logged.Gid)
						inodetmp.I_size = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))

						fecha := time.Now().String()
						copy(inodetmp.I_atime[:], spr.S_mtime[:])
						copy(inodetmp.I_ctime[:], fecha)
						copy(inodetmp.I_mtime[:], fecha)
						inodetmp.I_type = 1
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
						Error("MKFILE", "No se ha podido crear el directorio")
						return
					}
					past = GetFree(spr, pth, "BB")
					if past == -1 {
						Error("MKFILE", "No se ha podido crear el directorio")
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
					inodetmp.I_type = 1
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
						Error("MKFILE", "No se ha encontrado el disco.")
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
		Error("MKFILE", "No se ha encontrado el disco.")
		return
	}

	file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*bi, 0)
	var binInodeTemp bytes.Buffer
	binary.Write(&binInodeTemp, binary.BigEndian, inodetmp)
	EscribirBytes(file, binInodeTemp.Bytes())

	file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*bb+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*bb, 0)
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
	Mensaje("MKFILE", "Se ha creado el archivo "+ruta)
	file.Close()
}

func getDataFile(path []string, particion Structs.Particion, pth string) string {
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return ""
	}
	file.Seek(particion.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return ""
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return ""
	}

	if len(path) == 0 {
		Error("REP", "No se ha brindado una path válida")
		return ""
	}

	var aux []string
	for i := 0; i < len(path); i++ {
		aux = append(aux, path[i])
	}
	path = aux

	for v := 0; v < len(path); v++ {
		for i := 0; i < 16; i++ {
			if i < 16 {
				if inode.I_block[i] != -1 {
					folder = Structs.NewBloquesCarpetas()
					file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &folder)
					if err_ != nil {
						Error("REP", "Error al leer el archivo")
						return ""
					}

					for j := 0; j < 4; j++ {
						nombreFIle := ""
						for nam := 0; nam < len(folder.B_content[j].B_name); nam++ {
							if folder.B_content[j].B_name[nam] == 0 {
								continue
							}
							nombreFIle += string(folder.B_content[j].B_name[nam])
						}
						if Comparar(nombreFIle, path[v]) {
							inodeAux := inode
							inode = Structs.NewInodos()
							file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*folder.B_content[j].B_inodo, 0)

							data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &inode)
							if err_ != nil {
								Error("REP", "Error al leer el archivo")
								return ""
							}

							if inode.I_type == 1 && nombreFIle == path[len(path)-1] {
								if j == 2 {
									archivo := Structs.BloquesArchivos{}
									contenido := ""
									for k := 0; k < 16; k++ {
										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(k), 0)
										if inode.I_block[k] == -1 {
											break
										}
										data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
										buffer = bytes.NewBuffer(data)
										err_ = binary.Read(buffer, binary.BigEndian, &archivo)

										for l := 0; l < len(archivo.B_content); l++ {
											if archivo.B_content[l] != 0 {
												contenido += string(archivo.B_content[l])
											}
										}
									}

									if nombreFIle == path[len(path)-1] {
										return contenido
									}
								} else if j == 3 {
									archivo := Structs.BloquesArchivos{}
									contenido := ""
									for k := 0; k < 16; k++ {
										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(16+k), 0)
										if inode.I_block[k] == -1 {
											break
										}
										data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
										buffer = bytes.NewBuffer(data)
										err_ = binary.Read(buffer, binary.BigEndian, &archivo)

										for l := 0; l < len(archivo.B_content); l++ {
											if archivo.B_content[l] != 0 {
												contenido += string(archivo.B_content[l])
											}
										}
									}

									if nombreFIle == path[len(path)-1] {
										return contenido
									}
								}

							}

							break

						}
					}

				} else {
					break
				}
			}
		}
	}
	return ""
}

func setDataFile(path []string, p bool, s string, cont string, particion Structs.Particion, pth string) {
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	file.Seek(particion.Part_start, 0)
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

	if len(path) == 0 {
		Error("REP", "No se ha brindado una path válida")
		return
	}

	var aux []string
	for i := 0; i < len(path); i++ {
		aux = append(aux, path[i])
	}
	path = aux

	for v := 0; v < len(path); v++ {
		for i := 0; i < 16; i++ {
			if i < 16 {
				if inode.I_block[i] != -1 {
					folder = Structs.NewBloquesCarpetas()
					file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &folder)
					if err_ != nil {
						Error("REP", "Error al leer el archivo")
						return
					}

					for j := 0; j < 4; j++ {
						nombreFIle := ""
						for nam := 0; nam < len(folder.B_content[j].B_name); nam++ {
							if folder.B_content[j].B_name[nam] == 0 {
								continue
							}
							nombreFIle += string(folder.B_content[j].B_name[nam])
						}
						if Comparar(nombreFIle, path[v]) {
							inodeAux := inode
							inode = Structs.NewInodos()
							file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*folder.B_content[j].B_inodo, 0)

							data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &inode)

							agregado := false

							if inode.I_type == 1 && path[v] == path[len(path)-1] {
								archivo := Structs.BloquesArchivos{}
								file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{})), 0)

								tam := 0
								if s != "" {
									tam, err = strconv.Atoi(s)
									if err != nil {
										Error("MKFILE", "Se esperaba un entero para el parámetro size")
										return
									} else if tam < 0 {
										Error("MKFILE", "Se esperaba un número mayor a 0 para el parámetro size")
										return
									}
								}

								if inode.I_block[1] != -1 {
									if !Confirmar("Desea sobreescribir el archivo " + nombreFIle) {
										Mensaje("MKFILE", "No se ha sobreescrito el archivo "+nombreFIle)
										return
									}
								}
								agregado = true
								txt := ""
								if cont != "" {
									datosComoBytes, err := ioutil.ReadFile(cont)
									if err != nil {
										Error("MKFILE", "No se pudo leer el archivo "+cont)
									}
									datosComoString := string(datosComoBytes)
									txt += datosComoString
								} else if s != "" {
									contador := 0
									for contenido := 0; contenido < tam; contenido++ {
										txt += strconv.Itoa(contador)
										if contador == 9 {
											contador = -1
										}
										contador++
									}
								}

								tam = len(txt)
								var cadenasS []string
								if tam > 64 {
									for tam > 64 {
										auxtxt := ""
										for k := 0; k < 64; k++ {
											auxtxt += string(txt[k])
										}
										cadenasS = append(cadenasS, auxtxt)
										txt = cortarCadena(txt)
										tam = len(txt)
									}
									if tam < 64 && tam != 0 {
										cadenasS = append(cadenasS, txt)
									}
								} else {
									cadenasS = append(cadenasS, txt)
								}
								if len(cadenasS) > 16 {
									Error("MKFILE", "Se ha llenado la cantidad de archivos posibles y no se pueden generar más.")
									var auxCadenasS []string
									for k := 0; k < 16; k++ {
										auxCadenasS = append(auxCadenasS, cadenasS[k])
									}
									cadenasS = auxCadenasS
								}
								folderaux := Structs.NewBloquesCarpetas()
								file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i], 0)

								data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
								buffer = bytes.NewBuffer(data)
								err_ = binary.Read(buffer, binary.BigEndian, &folderaux)
								if err_ != nil {
									Error("REP", "Error al leer el archivo")
									return
								}

								file.Close()

								file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
								if err != nil {
									Error("MKFILE", "No se ha encontrado el disco.")
									return
								}
								if j == 2 {
									for k := 0; k < 16; k++ {
										if k == len(cadenasS) {
											break
										}
										var fbAux Structs.BloquesArchivos
										if inode.I_block[k] == -1 {
											file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(k), 0)
											var binAux bytes.Buffer
											binary.Write(&binAux, binary.BigEndian, fbAux)
											EscribirBytes(file, binAux.Bytes())
										} else {
											fbAux = archivo
										}

										copy(fbAux.B_content[:], cadenasS[k])

										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(k), 0)
										var bin6 bytes.Buffer
										binary.Write(&bin6, binary.BigEndian, fbAux)
										EscribirBytes(file, bin6.Bytes())

									}
									principal := inode.I_block[0]
									for k := 0; k < len(cadenasS); k++ {

										inode.I_block[k] = principal
									}
								} else if j == 3 {
									for k := 0; k < 16; k++ {
										if k == len(cadenasS) {
											break
										}
										var fbAux Structs.BloquesArchivos
										if inode.I_block[k] == -1 {
											file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*(int64(16+k)), 0)
											var binAux bytes.Buffer
											binary.Write(&binAux, binary.BigEndian, fbAux)
											EscribirBytes(file, binAux.Bytes())
										} else {
											fbAux = archivo
										}

										copy(fbAux.B_content[:], cadenasS[k])

										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*(int64(16+k)), 0)
										var bin6 bytes.Buffer
										binary.Write(&bin6, binary.BigEndian, fbAux)
										EscribirBytes(file, bin6.Bytes())

									}
									principal := inode.I_block[0]
									for k := 0; k < len(cadenasS); k++ {
										inode.I_block[k] = principal
									}
								}
							}

							if agregado {
								file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*folder.B_content[j].B_inodo, 0)
								var inodos bytes.Buffer
								binary.Write(&inodos, binary.BigEndian, inode)
								EscribirBytes(file, inodos.Bytes())

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
	}
	return
}

func cortarCadena(txt string) string {
	aux := ""
	for i := 64; i < len(txt); i++ {
		aux += string(txt[i])
	}

	return aux
}
