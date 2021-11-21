package trellometrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/adlio/trello"
)

type CachedCardActions struct {
	cacheDir string
}

func CreateCachedCardActions(cacheDir string) (*CachedCardActions, error) {
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

	return &CachedCardActions{cacheDir: appCacheDir}, nil
}

func (a *CachedCardActions) ListChangeActions(card *trello.Card) (trello.ActionCollection, error) {
	cacheKey := filepath.Join(
		a.cacheDir,
		fmt.Sprintf("card-actions-%s.json", card.ID),
	)

	cache, err := ioutil.ReadFile(cacheKey)
	if err != nil {
		actions, err := card.GetListChangeActions()
		if err != nil {
			return actions, err
		}

		cache, err = json.Marshal(actions)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(cacheKey, cache, 0754)
		if err != nil {
			return nil, err
		}
	}

	actions := trello.ActionCollection{}
	err = json.Unmarshal(cache, &actions)
	if err != nil {
		return nil, err
	}

	return actions, nil
}
