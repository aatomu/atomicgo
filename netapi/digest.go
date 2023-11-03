package netapi

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

type DigestAuth struct {
	realm string
	nonce map[string]bool
}

func DigestAuthNew(Realm string) *DigestAuth {
	return &DigestAuth{
		realm: Realm,
		nonce: map[string]bool{},
	}
}

// Send "Digest Auth" To Client
func (d *DigestAuth) Require(w http.ResponseWriter, lifetime time.Duration) {
	nonceText := fmt.Sprintf("%x", rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	d.nonce[nonceText] = false

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Digest realm="%s",nonce="%s",algorithm=MD5,qop="auth"`, d.realm, nonceText))
	w.WriteHeader(http.StatusUnauthorized)

	go func() {
		time.Sleep(lifetime)
		delete(d.nonce, nonceText)
	}()
}

func (d *DigestAuth) GetUsername(r *http.Request) (username string) {
	Auth := r.Header.Values("Authorization")
	if len(Auth) != 1 {
		return
	}
	value := regexp.MustCompile(`username=\"(.+)\"`).FindStringSubmatch(Auth[0])
	if len(value) != 2 {
		return
	}
	username = value[1]
	return
}

// Digest Auth CheckSum
func (d DigestAuth) Checksum(user string, r *http.Request) (ok bool, err error) {
	Auth := r.Header.Values("Authorization")
	if len(Auth) != 1 {
		return false, errors.New("authorization required")
	}
	// response = MD5(A1):nonce:nc:cnonce:qop:MD5(A2)
	// A1(user) = username:realm:password
	// A2 = Method:uri

	// Get nc Value
	var nc string
	value := regexp.MustCompile(`nc=([0-9]+)`).FindStringSubmatch(Auth[0])
	if len(value) != 2 {
		return false, fmt.Errorf("invaild nc value")
	}
	nc = value[1]

	// Get cnonce Value
	var cnonce string
	value = regexp.MustCompile(`cnonce=\"(\w+?)\"`).FindStringSubmatch(Auth[0])
	if len(value) != 2 {
		return false, fmt.Errorf("invaild cnonce value")
	}
	cnonce = value[1]

	// Get A2 Value
	A2 := fmt.Sprintf("%s:%s", r.Method, r.RequestURI)
	md5Hash := md5.Sum([]byte(A2))
	A2 = fmt.Sprintf("%x", md5Hash)

	// Get response Value
	var response string
	value = regexp.MustCompile("response=\"(.+?)\"").FindStringSubmatch(Auth[0])
	if len(value) != 2 {
		return false, fmt.Errorf("invaild response value")
	}
	response = value[1]

	for nonce := range d.nonce {
		// Create Checksum Hash
		checksumHash := fmt.Sprintf("%s:%s:%s:%s:auth:%s", user, nonce, nc, cnonce, A2)
		md5Hash = md5.Sum([]byte(checksumHash))
		checksumHash = fmt.Sprintf("%x", md5Hash)
		if response == checksumHash {
			return true, nil
		}
	}
	return false, nil
}

func (d DigestAuth) NewUser(username, password string) (user string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, d.realm, password))))
}
