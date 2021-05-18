package kaoriData

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/fatih/structs"
)

type Manga struct {
	Id string
	Name string
	ChaptersNumber int
	Chapters []*Chapter
}

func NewManga() *Manga {
	return &Manga{}
}

func (m *Manga) SendToKaori(kaoriServer, token string) error {
	return sendToKaori(m, kaoriServer, token)
}

func (m *Manga) SendToDatabase(c *firestore.Client, ctx context.Context) error {

	ma := structs.Map(m)
	delete(ma, "Chapters")

	//Send manga data
	mangaDoc := c.Collection("Manga").Doc(m.Id)
	_, err := mangaDoc.Set(ctx, ma, firestore.MergeAll)
	if err != nil {
		return err
	}

	for _, ch := range m.Chapters {

		for _, p := range ch.Pages {

			mc := structs.Map(ch)
			delete(mc, "Pages")

			//Send language data
			_, err := mangaDoc.Collection("Languages").
				Doc(p.Language).Set(ctx, map[string]string{
					"Fill": "hcbhdbchd",
			}, firestore.MergeAll)
			if err != nil {
				return err
			}

			//Send chapters data
			chapterDoc := mangaDoc.Collection("Languages").
										Doc(p.Language).
										Collection("Chapters").
										Doc(ch.Number)

			_, err = chapterDoc.Set(ctx, mc, firestore.MergeAll)
			if err != nil {
				return err
			}

			//Send pages
			pagesDoc := chapterDoc.Collection("Pages").
									Doc(p.Number).
									Collection("Servers").
									Doc(p.Server)

			_, err = chapterDoc.Collection("Pages").
				Doc(p.Number).Set(ctx, map[string]string{
					"Fill": "hfsbv hdbhdb",
			}, firestore.MergeAll)
			if err != nil {
				return err
			}

			_, err = pagesDoc.Set(ctx, map[string]string{
				"Link": p.Link,
			}, firestore.MergeAll)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func (m *Manga) AppendFile(filePath string) error {
	return appendFile(m, filePath)
}
