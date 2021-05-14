package kaoriData

type Manga struct {
	Id string
	Name string
	ChaptersNumber int
	Chapters []*Chapter
}

type Chapter struct {
	Number string
	Title string
	Pages []*Page
}

type Page struct {
	Number string
	Language string
	Server string
}

func (m *Manga) SendToKaori(kaoriServer, token string) error {
	return sendToKaori(m, kaoriServer, token)
}

func (m *Manga) AppendFile(filePath string) error {
	return appendFile(m, filePath)
}
