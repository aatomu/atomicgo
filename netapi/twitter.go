package netapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

// TwitterAccount はTwitterの認証用の情報
type TwitterAPIKeys struct {
	AccessToken       string `json:"accessToken"`
	AccessTokenSecret string `json:"accessSecret"`
	APIKey            string `json:"APIKey"`
	APISecret         string `json:"APISecret"`
	Token             string `json:"Token"`
}

// TwitterAPIKeyを.jsonから入手
func TwitterAPIkeysGet(path string) (APIKeys TwitterAPIKeys, err error) {
	// Json読み込み
	if _, stat := os.Stat(path); stat != nil {
		return TwitterAPIKeys{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return TwitterAPIKeys{}, err
	}

	// 構造体にセット
	err = json.Unmarshal(raw, &APIKeys)
	return APIKeys, err
}

// TwitterAPIに設定
func (api *TwitterAPIKeys) TwitterAPISet() *anaconda.TwitterApi {
	return anaconda.NewTwitterApiWithCredentials(api.AccessToken, api.AccessTokenSecret, api.APIKey, api.APISecret)
}

func (api *TwitterAPIKeys) TwitterSearch(searchLimit int, keyWord string) (anaconda.SearchResponse, error) {
	// 認証
	twitterApi := api.TwitterAPISet()

	// 検索上限確認
	v := url.Values{}
	v.Set("count", fmt.Sprint(searchLimit))

	//検索
	return twitterApi.GetSearch(keyWord, v)
}
