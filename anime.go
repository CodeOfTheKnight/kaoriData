package kaoriData

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type EpLanguage string
type EpQuality string

type Anime struct {
	Id string `firestore:"-"`
	Name string `firestore:"name"`
	Episodes []*Episode `firestore:"episodes"`
}

type Episode struct {
	Number string `firestore:"number"`
	Title string
	Links map[EpLanguage]map[EpQuality]map[string]StreamLink `firestore:"links"`
}

type StreamLink struct{
	Link string `firestore:"link"`
	Width int `firestore:"width"`
	Height int `firestore:"height"`
	Duration float64 `firestore:"duration"`
	Bitrate int `firestore:"bitrate"`
}

func NewAnime() *Anime {
	var a Anime
	return &a
}

func NewEpisode() *Episode {
	var ep Episode
	ep.Links = make(map[EpLanguage]map[EpQuality]map[string]StreamLink)
	return &ep
}

func (a *Anime) SendToDb(c *firestore.Client, ctx context.Context) error {

	var eps []*Episode

	//Season info
	m := structs.Map(a)
	delete(m, "Episodes")
	fmt.Println("MAPPA:", m)

	//Episode info
	eps = a.Episodes

	//Write season info to database
	_, err := c.Collection("Anime").Doc(a.Id).Set(ctx, m, firestore.MergeAll)
	if err != nil {
		return err
	}

	//Write episodes of sesason to database
	for _, ep := range eps {

		//Check languages
		for lang, _ := range ep.Links {
			for quality, _ := range ep.Links[lang] {
				for server, streamLinks := range ep.Links[lang][quality] {

					streamLinksMap := structs.Map(streamLinks)

					_, err = c.Collection("Anime").
						Doc(a.Id).
						Collection("Episodes").
						Doc(ep.Number).
						Collection("Languages").
						Doc(string(lang)).
						Collection("Quality").
						Doc(string(quality)).
						Collection("Servers").
						Doc(server).
						Set(ctx, streamLinksMap, firestore.MergeAll)

					if err != nil {
						return err
					}

					fmt.Println(len(ep.Links))
				}
			}

			//Write episode data
			_, err := c.Collection("Anime").Doc(a.Id).
												Collection("Episodes").
												Doc(ep.Number).
												Set(ctx, map[string]string{"Title": ep.Title}, firestore.MergeAll)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (a *Anime) SendToKaori(kaoriUrl, token string) error {

	//Create JSON
	data, err := json.MarshalIndent(a, " ", "\t")
	if err != nil {
		return errors.New("Error to create JSON: " + err.Error())
	}

	fmt.Println("DATA:", string(data))

	//Create client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	//Create request
	req, err := http.NewRequest("POST", kaoriUrl, bytes.NewReader(data))
	if err != nil {
		return errors.New("Error to create request: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer " + token)

	//Do request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Error to send data, status code: " + resp.Status)
	}

	return nil
}

func (sl *StreamLink) GetQuality(link string) error {

	command := fmt.Sprintf("ffprobe -v error -select_streams v:0 -show_entries stream=width,height,duration,bit_rate -of default=noprint_wrappers=1 %s", link)

	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Split(line, "=")
		switch(fields[0]){
		case "width":
			num, _ := strconv.Atoi(fields[1])
			sl.Width = num
		case "height":
			num, _ := strconv.Atoi(fields[1])
			sl.Height = num
		case "duration":
			num, _ := strconv.ParseFloat(fields[1], 64)
			sl.Duration = num
		case "bit_rate":
			num, _ := strconv.Atoi(fields[1])
			sl.Bitrate = num
		}
	}

	return nil
}