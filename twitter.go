package atomicgo

import (
	"encoding/json"
	"fmt"
	"log"
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
func TwitterAPIkeysGet(path string) (APIKeys TwitterAPIKeys) {
	// Json読み込み
	raw, err := ReadAndCreateFileFlash(path)
	if err != nil {
		log.Println("failed read APIKeys")
	}
	// 構造体にセット
	json.Unmarshal(raw, &APIKeys)
	return
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
