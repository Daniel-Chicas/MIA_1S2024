package Structs

type EBR struct {
	Part_status byte
	Part_fit    byte
	Part_start  int64
	Part_size   int64
	Part_next   int64
	Part_name   [16]byte
}

func NewEBR() EBR {
	var eb EBR
	eb.Part_status = '0'
	eb.Part_size = 0
	eb.Part_next = -1
	return eb
}
