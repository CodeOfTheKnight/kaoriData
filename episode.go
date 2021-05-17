package kaoriData

import (
	"errors"
	"strconv"
)

type Episode struct {
	Number string `firestore:"number"`
	Title string
	Videos []*Video
}

func NewEpisode() *Episode {
	return &Episode{}
}

func (ep *Episode) CheckEpisode() error {

	if err := ep.checkNumber(); err != nil {
		return err
	}

	//TODO: Get title if it hasn't been set

	for i, _ := range ep.Videos {
		if err := ep.Videos[i].CheckVideo(); err != nil {
			return err
		}
	}

	return nil
}

func (ep *Episode) checkNumber() error {

	if ep.Number == "" {
		return errors.New("Number of episode not setted")
	}

	if _, err := strconv.Atoi(ep.Number); err != nil {
		return errors.New("Number of episode not valid")
	}

	return nil
}
