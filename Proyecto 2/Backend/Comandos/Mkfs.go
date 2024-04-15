package Comandos

import (
	"MIA_Proyecto2_201807079/Structs"
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"strings"
	"time"
	"unsafe"
)

func ValidarDatosMKFS(context []string) {
	id := ""
	tipo := "Full"

	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "id") {
			id = tk[1]
		} else if Comparar(tk[0], "type") {
			if Comparar(tk[1], "fast") || Comparar(tk[1], "full") {
				tipo = tk[1]
			} else {
				Error("MKFS", "El comando type debe tener valores específicos")
				return
			}
		}
	}
	if id == "" {
		Error("MKFS", "EL comando requiere el parámetro id obligatoriamente")
		return
	}
	mkfs(id, tipo)
}

//  exec -path=/home/daniel/Escritorio/ArchivosPrueba/ArchivoEjemplo2.script

func mkfs(id string, t string) {
	p := ""
	particion := GetMount("MKFS", id, &p)
	n := math.Floor(float64(particion.Part_size-int64(unsafe.Sizeof(Structs.SuperBloque{}))) / float64(4+unsafe.Sizeof(Structs.Inodos{})+3*unsafe.Sizeof(Structs.BloquesArchivos{})))

	spr := Structs.NewSuperBloque()
	spr.S_magic = 0xEF53
	spr.S_inode_size = int64(unsafe.Sizeof(Structs.Inodos{}))
	spr.S_block_size = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))
	spr.S_inodes_count = int64(n)
	spr.S_free_inodes_count = int64(n)
	spr.S_blocks_count = int64(3 * n)
	spr.S_free_blocks_count = int64(3 * n)
	fecha := time.Now().String()
	copy(spr.S_mtime[:], fecha)
	spr.S_mnt_count = spr.S_mnt_count + 1
	spr.S_filesystem_type = 2
	ext2(spr, particion, int64(n), p)
}

func ext2(spr Structs.SuperBloque, p Structs.Particion, n int64, path string) {
	spr.S_bm_inode_start = p.Part_size + int64(unsafe.Sizeof(Structs.SuperBloque{}))
	spr.S_bm_block_start = spr.S_bm_inode_start + n
	spr.S_inode_start = spr.S_bm_block_start + (3 * n)
	spr.S_block_start = spr.S_bm_inode_start + (n * int64(unsafe.Sizeof(Structs.Inodos{})))

	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	//file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("MKFS", "No se ha encontrado el disco.")
		return
	}

	file.Seek(p.Part_start, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, spr)
	EscribirBytes(file, binario2.Bytes())

	zero := '0'
	file.Seek(spr.S_bm_inode_start, 0)
	for i := 0; i < int(n); i++ {
		var binarioZero bytes.Buffer
		binary.Write(&binarioZero, binary.BigEndian, zero)
		EscribirBytes(file, binarioZero.Bytes())
	}

	file.Seek(spr.S_bm_block_start, 0)
	for i := 0; i < 3*int(n); i++ {
		var binarioZero bytes.Buffer
		binary.Write(&binarioZero, binary.BigEndian, zero)
		EscribirBytes(file, binarioZero.Bytes())
	}

	inode := Structs.NewInodos()
	//INICIALIZANDO EL INODO
	inode.I_uid = -1
	inode.I_gid = -1
	inode.I_size = -1
	for i := 0; i < len(inode.I_block); i++ {
		inode.I_block[i] = -1
	}
	inode.I_type = -1
	inode.I_perm = -1

	file.Seek(spr.S_inode_start, 0)
	for i := 0; i < int(n); i++ {
		var binarioInodos bytes.Buffer
		binary.Write(&binarioInodos, binary.BigEndian, inode)
		EscribirBytes(file, binarioInodos.Bytes())
	}

	folder := Structs.NewBloquesCarpetas()

	for i := 0; i < len(folder.B_content); i++ {
		folder.B_content[i].B_inodo = -1
	}

	file.Seek(spr.S_block_start, 0)
	for i := 0; i < int(n); i++ {
		var binarioFolder bytes.Buffer
		binary.Write(&binarioFolder, binary.BigEndian, folder)
		EscribirBytes(file, binarioFolder.Bytes())
	}
	file.Close()

	recuperado := Structs.NewSuperBloque()
	//ABRIR ARCHIVO
	//file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)

	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("MKFS", "No se ha encontrado el disco.")
		return
	}

	file.Seek(p.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &recuperado)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	file.Close()

	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	fecha := time.Now().String()
	copy(inode.I_atime[:], fecha)
	copy(inode.I_ctime[:], fecha)
	copy(inode.I_mtime[:], fecha)
	inode.I_type = 0
	inode.I_perm = 664
	inode.I_block[0] = 0

	fb := Structs.NewBloquesCarpetas()
	copy(fb.B_content[0].B_name[:], ".")
	fb.B_content[0].B_inodo = 0
	copy(fb.B_content[1].B_name[:], "..")
	fb.B_content[1].B_inodo = 0
	copy(fb.B_content[2].B_name[:], "users.txt")
	fb.B_content[2].B_inodo = 1

	dataArchivo := "1,G,root\n1,U,root,root,123\n"
	inodetmp := Structs.NewInodos()
	inodetmp.I_uid = 1
	inodetmp.I_gid = 1
	inodetmp.I_size = int64(unsafe.Sizeof(dataArchivo) + unsafe.Sizeof(Structs.BloquesCarpetas{}))

	copy(inodetmp.I_atime[:], fecha)
	copy(inodetmp.I_ctime[:], fecha)
	copy(inodetmp.I_mtime[:], fecha)
	inodetmp.I_type = 1
	inodetmp.I_perm = 664
	inodetmp.I_block[0] = 1

	inode.I_size = inodetmp.I_size + int64(unsafe.Sizeof(Structs.BloquesCarpetas{})) + int64(unsafe.Sizeof(Structs.Inodos{}))

	var fileb Structs.BloquesArchivos
	copy(fileb.B_content[:], dataArchivo)

	file, err = os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	//file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("MKFS", "No se ha encontrado el disco.")
		return
	}
	file.Seek(spr.S_bm_inode_start, 0)
	caracter := '1'

	var bin1 bytes.Buffer
	binary.Write(&bin1, binary.BigEndian, caracter)
	EscribirBytes(file, bin1.Bytes())
	EscribirBytes(file, bin1.Bytes())

	file.Seek(spr.S_bm_block_start, 0)
	var bin2 bytes.Buffer
	binary.Write(&bin2, binary.BigEndian, caracter)
	EscribirBytes(file, bin2.Bytes())
	EscribirBytes(file, bin1.Bytes())

	file.Seek(spr.S_inode_start, 0)

	var bin3 bytes.Buffer
	binary.Write(&bin3, binary.BigEndian, inode)
	EscribirBytes(file, bin3.Bytes())

	file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	var bin4 bytes.Buffer
	binary.Write(&bin4, binary.BigEndian, inodetmp)
	EscribirBytes(file, bin4.Bytes())

	file.Seek(spr.S_block_start, 0)

	var bin5 bytes.Buffer
	binary.Write(&bin5, binary.BigEndian, fb)
	EscribirBytes(file, bin5.Bytes())

	//fmt.Println(spr.S_block_start + int64(unsafe.Sizeof(Structs.BloquesCarpetas{})))

	file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{})), 0)
	var bin6 bytes.Buffer
	binary.Write(&bin6, binary.BigEndian, fileb)
	EscribirBytes(file, bin6.Bytes())

	file.Close()

	nombreParticion := ""
	for i := 0; i < len(p.Part_name); i++ {
		if p.Part_name[i] != 0 {
			nombreParticion += string(p.Part_name[i])
		}
	}
	Mensaje("MKFS", "Se ha formateado la partición "+nombreParticion+" correctamente.")
}
