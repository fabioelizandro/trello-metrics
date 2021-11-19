package kanban

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CachedBoard struct {
	underlyingImpl Board
	cacheDir       string
}

func CreateCachedBoard(underlyingImpl Board, cacheDir string) (*CachedBoard, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	appCacheDir := filepath.Join(userCacheDir, cacheDir)

	if _, err := os.Stat(appCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(appCacheDir, 0754)
		if err != nil {
			return nil, err
		}
	}

	return &CachedBoard{underlyingImpl: underlyingImpl, cacheDir: appCacheDir}, nil
}

func (t *CachedBoard) DoneCards() ([]*Card, error) {
	cacheKey := filepath.Join(
		t.cacheDir,
		"done-cards.json",
	)

	cache, err := ioutil.ReadFile(cacheKey)
	if err != nil {
		cards, err := t.underlyingImpl.DoneCards()
		if err != nil {
			return cards, err
		}

		cache, err = json.Marshal(cards)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(cacheKey, cache, 0754)
		if err != nil {
			return nil, err
		}
	}

	cards := []*Card{}
	err = json.Unmarshal(cache, &cards)
	if err != nil {
		return nil, err
	}

	return cards, nil
}
