package Structs

type BloquesCarpetas struct {
	B_content [4]Content
}

func NewBloquesCarpetas() BloquesCarpetas {
	var bl BloquesCarpetas
	for i := 0; i < len(bl.B_content); i++ {
		bl.B_content[i] = NewContent()
	}
	return bl
}
