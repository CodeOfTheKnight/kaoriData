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
	Videos []*Video
}

type Video struct {
	Language string
	Modality string
	Quality *InfoQuality
	Server string
	StreamLink *StreamLink
}

type InfoQuality struct {
	Width int `firestore:"width"`
	Height int `firestore:"height"`
}

type StreamLink struct{
	Link string `firestore:"link"`
	Duration float64 `firestore:"duration"`
	Bitrate int `firestore:"bitrate"`
}

func NewAnime() *Anime {
	return &Anime{}
}

func NewEpisode() *Episode {
	return &Episode{}
}

func NewVideo() *Video {
	return &Video{}
}

func (a *Anime) SendToDb(c *firestore.Client, ctx context.Context) error {

	var eps []*Episode

	//Episode info
	eps = a.Episodes

	//Write season info to database
	_, err := c.Collection("Anime").
				Doc(a.Id).
				Set(ctx, map[string]string{
					"Name": a.Name,
				}, firestore.MergeAll)

	if err != nil {
		return err
	}

	//Write episodes of sesason to database
	for _, ep := range eps {

		for _, video :=  range ep.Videos {

			q := strconv.Itoa(video.Quality.Height) + "p"
			if q == "0p" {
				q = "undefined"
			}

			//Send streamLinks and create all collections
			_, err = c.Collection("Anime").
				Doc(a.Id).
				Collection("Episodes").
				Doc(ep.Number).
				Collection("Languages").
				Doc(video.Language).
				Collection("Quality").
				Doc(q).
				Collection("Servers").
				Doc(video.Server).
				Set(ctx, structs.Map(video.StreamLink), firestore.MergeAll)

			if err != nil {
				return err
			}

			//Send language info
			_, err = c.Collection("Anime").
				Doc(a.Id).
				Collection("Episodes").
				Doc(ep.Number).
				Collection("Languages").
				Doc(video.Language).Set(ctx, map[string]string{
					"Modality": video.Modality,
				}, firestore.MergeAll)

			if err != nil {
				return err
			}

			//Send quality info
			_, err = c.Collection("Anime").
				Doc(a.Id).
				Collection("Episodes").
				Doc(ep.Number).
				Collection("Languages").
				Doc(video.Language).
				Set(ctx, structs.Map(video.Quality), firestore.MergeAll)

		}

		//Send episode data
		_, err = c.Collection("Anime").
										Doc(a.Id).
										Collection("Episodes").
										Doc(ep.Number).
										Set(ctx, map[string]string{
											"Title": ep.Title,
										}, firestore.MergeAll)

		if err != nil {
			return err
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

func (v *Video) GetQuality(link string)  error {

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
		v.Quality.Height = 0
		v.Quality.Width = 0
		return err
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Split(line, "=")
		switch(fields[0]){
		case "width":
			num, _ := strconv.Atoi(fields[1])
			v.Quality.Width = num
		case "height":
			num, _ := strconv.Atoi(fields[1])
			v.Quality.Height = num
		case "duration":
			num, _ := strconv.ParseFloat(fields[1], 64)
			v.StreamLink.Duration = num
		case "bit_rate":
			num, _ := strconv.Atoi(fields[1])
			v.StreamLink.Bitrate = num
		}
	}

	return nil
}