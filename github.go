package atomicgo

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type gitHubSession struct {
	Token map[string]string
	Rate  struct {
		Used      int       //使用回数
		Limit     int       //使用上限
		Remaining int       //残り回数
		Reset     time.Time //リセット
	}
	Version []string
}

func NewGithubAPI(token string) (s gitHubSession) {
	s.Token = map[string]string{}
	s.Token["Authorization"] = "token " + token
	return
}

// endPoints : https://docs.github.com/ja/rest/overview/endpoints-available-for-github-apps
// example : Get(/users/atomu21263)
func (s *gitHubSession) Get(endPoint string) (data []byte, err error) {
	// リクエスト送信
	resp, err := HttpReqest(Get, "https://api.github.com"+endPoint, "", s.Token)
	// エラーチェック
	if err != nil {
		return
	}
	// ヘッダー処理
	s.parseHeader(resp.Header)
	defer resp.Body.Close()
	// 読み取り
	return ioutil.ReadAll(resp.Body)
}

func (s *gitHubSession) parseHeader(h http.Header) {
	// Used
	s.Rate.Used, _ = strconv.Atoi(h["X-Ratelimit-Used"][0])
	// Limit
	s.Rate.Limit, _ = strconv.Atoi(h["X-Ratelimit-Limit"][0])
	// Remaining
	s.Rate.Remaining, _ = strconv.Atoi(h["X-Ratelimit-Remaining"][0])
	// Reset
	t, _ := strconv.Atoi(h["X-Ratelimit-Reset"][0])
	s.Rate.Reset = time.Unix(int64(t), 0)
	// Version
	s.Version = h["X-Github-Media-Type"]
}
