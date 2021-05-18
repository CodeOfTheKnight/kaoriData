package kaoriData

type Page struct {
	Number string
	Language string
	Server string
	Link string
}

func NewPage() *Page {
	return &Page{}
}

