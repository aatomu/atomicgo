package netapi

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DigestAuth struct {
	realm      string
	nonce      map[string]string // nonce="timestamp Hash(timestamp+randText)", nonce = {timestamp:"nonce"}
	activeTime time.Duration
	deleteTime time.Duration
}

func DigestAuthNew(Realm string, activeTime, deleteTime time.Duration) *DigestAuth {
	return &DigestAuth{
		realm:      Realm,
		nonce:      map[string]string{},
		activeTime: activeTime,
		deleteTime: deleteTime,
	}
}

// Send "Digest Auth" To Client
func (d *DigestAuth) Require(r *http.Request, w http.ResponseWriter) {
	// New Nonce
	timestamp := fmt.Sprintf("%d", time.Now().UnixMicro())
	randText := fmt.Sprintf("%x", rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	d.nonce[timestamp] = randText

	if d.shouldUserInput(r) {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Digest realm="%s",nonce="%s",algorithm=MD5,qop="auth"`, d.realm, newNonce(timestamp, randText)))
	} else {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Digest realm="%s",nonce="%s",stale=true,algorithm=MD5,qop="auth"`, d.realm, newNonce(timestamp, randText)))
	}
	w.WriteHeader(http.StatusUnauthorized)

	// Delete Nonce
	go func() {
		time.Sleep(d.activeTime + d.deleteTime)
		delete(d.nonce, timestamp)
	}()
}

// Bypass User Input (if nonce is stale)
func (d *DigestAuth) shouldUserInput(r *http.Request) bool {
	// Exist Authorization Header
	Auth := r.Header.Values("Authorization")
	if len(Auth) != 1 {
		return true
	}

	value := regexp.MustCompile(`nonce=\"(.+?)\"`).FindStringSubmatch(Auth[0])
	// Exist Nonce Tag
	if len(value) != 2 {
		return true
	}

	content := strings.Split(value[1], " ")
	// Available Tag
	if len(content) != 2 {
		return true
	}

	t, _ := strconv.ParseInt(content[0], 10, 64)
	timestamp := time.UnixMicro(t)
	// Time check
	// timestamp > time.Now() - lifetime
	isOld := time.Since(timestamp) > d.activeTime
	return isOld
}

func (d *DigestAuth) GetUsername(r *http.Request) (username string) {
	Auth := r.Header.Values("Authorization")
	if len(Auth) != 1 {
		return
	}
	value := regexp.MustCompile(`username=\"(.+?)\"`).FindStringSubmatch(Auth[0])
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
	value := regexp.MustCompile(`nc=([0-9a-f]+)`).FindStringSubmatch(Auth[0])
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

	for timestamp, randText := range d.nonce {
		// Create Checksum Hash
		checksumHash := fmt.Sprintf("%s:%s:%s:%s:auth:%s", user, newNonce(timestamp, randText), nc, cnonce, A2)
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

// nonce="timestamp Hash(timestamp+randText)"
func newNonce(timestamp, randText string) (nonce string) {
	return fmt.Sprintf("%s %x", timestamp, sha256.Sum256([]byte(timestamp+randText)))
}
