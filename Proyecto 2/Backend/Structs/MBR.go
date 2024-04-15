package Structs

type MBR struct {
	Mbr_tamano         int64
	Mbr_fecha_creacion [16]byte
	Mbr_dsk_signature  int64
	Dsk_fit            [1]byte
	Mbr_partition_1    Particion
	Mbr_partition_2    Particion
	Mbr_partition_3    Particion
	Mbr_partition_4    Particion
}

func NewMBR() MBR {
	var mb MBR
	return mb
}
