package Structs

type Content struct {
	B_name  [12]byte
	B_inodo int64
}

func NewContent() Content {
	var cont Content
	cont.B_inodo = -1
	return cont
}
