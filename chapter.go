package kaoriData

type Chapter struct {
	Number string
	Title string
	Pages []*Page
}

func NewChapter() *Chapter {
	return &Chapter{}
}
