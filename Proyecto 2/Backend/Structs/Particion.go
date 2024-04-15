package Structs

type Particion struct {
	Part_status byte
	Part_type   byte
	Part_fit    byte
	Part_start  int64
	Part_size   int64
	Part_name   [16]byte
}

func NewParticion() Particion {
	var Part Particion
	Part.Part_status = '0'
	Part.Part_type = 'P'
	Part.Part_fit = 'F'
	Part.Part_start = -1
	Part.Part_size = 0
	Part.Part_name = [16]byte{}
	return Part
}
