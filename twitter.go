package atomicgo

import (
	"encoding/json"
	"fmt"
	"net/url"

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

//TwitterAPIKeyを.jsonから入手
func TwitterAPIkeysGet(path string) (APIKeys TwitterAPIKeys, success bool) {
	// Json読み込み
	raw, ok := ReadAndCreateFileFlash(path)
	if !ok {
		return TwitterAPIKeys{}, false
	}
	// 構造体にセット
	err := json.Unmarshal(raw, &APIKeys)
	PrintError("Failed Marshal APIKeys", err)
	return APIKeys, true
}

//TwitterAPIに設定
func TwitterAPISet(APIKeys TwitterAPIKeys) (API *anaconda.TwitterApi) {
	return anaconda.NewTwitterApiWithCredentials(APIKeys.AccessToken, APIKeys.AccessTokenSecret, APIKeys.APIKey, APIKeys.APISecret)
}

func TwitterSearch(APIKeys TwitterAPIKeys, searchLimit int, keyWord string) (tweets []anaconda.Tweet) {
	// 認証
	api := TwitterAPISet(APIKeys)

	// 検索上限確認
	v := url.Values{}
	v.Set("count", fmt.Sprint(searchLimit))

	//検索
	searchResult, _ := api.GetSearch(keyWord, v)
	tweets = searchResult.Statuses
	return
}
