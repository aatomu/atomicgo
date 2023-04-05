package atomicgo

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// IOを分けて返却
func StdPipe(io *exec.Cmd) (inIo io.WriteCloser, outIo io.ReadCloser, errIo io.ReadCloser) {
	inIo, _ = io.StdinPipe()
	outIo, _ = io.StdoutPipe()
	errIo, _ = io.StderrPipe()
	return
}

// Regexp Match
func RegMatch(text string, check string) (match bool) {
	return regexp.MustCompile(check).MatchString(text)
}

// Regexp Replace
func RegReplace(fromText string, toText string, check string) (replaced string) {
	return regexp.MustCompile(check).ReplaceAllString(fromText, toText)
}

// Rand Generate
func Rand(max int) (result int) {
	result = rand.New(rand.NewSource(time.Now().UnixNano())).Int() % max
	return
}

// String Cut
func StrCut(text, suffix string, max int) (result string) {
	textArray := strings.Split(text, "")
	if len(textArray) < max {
		return text
	}
	for i := 0; i < max; i++ {
		result += textArray[i]
	}
	result += suffix
	return
}

// Listen Kill,Term,Interupt to Channel
func BreakSignal() (sc chan os.Signal) {
	sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	return
}

// Error表示
func PrintError(message string, err error) (errored bool) {
	if err != nil {
		trackBack := ""
		// 原因を特定
		for i := 1; true; i++ {
			pc, file, line, _ := runtime.Caller(i)
			trackBack += fmt.Sprintf("> %s:%d %s()\n", filepath.Base(file), line, RegReplace(runtime.FuncForPC(pc).Name(), "", "^.*/"))
			_, _, _, ok := runtime.Caller(i + 3)
			if !ok {
				break
			}
			// インデント
			for j := 0; j < i; j++ {
				trackBack += "  "
			}
		}
		//表示
		SetPrintWordColor(255, 0, 0)
		fmt.Printf("[Error] Message:\"%s\" Error:\"%s\"\n", message, err.Error())
		fmt.Printf("%s", trackBack)
		ResetPrintWordColor()
		return true
	}
	return false
}

// 文字の色を変える
func SetPrintWordColor(r int, g int, b int) {
	//文字色指定
	fmt.Print("\x1b[38;2;" + fmt.Sprint(r) + ";" + fmt.Sprint(g) + ";" + fmt.Sprint(b) + "m")
}

// 文字の色を元に戻す
func ResetPrintWordColor() {
	//文字色リセット
	fmt.Print("\x1b[39m")
}

// 背景の色を変える
func SetPrintBackColor(r int, g int, b int) {
	//背景色指定
	fmt.Print("\x1b[48;2;" + fmt.Sprint(r) + ";" + fmt.Sprint(g) + ";" + fmt.Sprint(b) + "m")
}

// 背景の色を元に戻す
func ResetPrintBackColor() {
	//背景色リセット
	fmt.Print("\x1b[49m")
}

// Byte を intに
func ConvBtoI(b []byte) int {
	n := 0
	length := len(b)
	for i := 0; i < length; i++ {
		m := 1
		for j := 0; j < length-i-1; j++ {
			m = m * 256
		}
		m = m * int(b[i])
		n = n + m
	}
	return n
}

type ExMap struct {
	Sm sync.Map
}

// 排他的Mapを入手
// 排他的MAPの型 : sync.Map
func NewExMap() *ExMap {
	return &ExMap{}
}

// 排他的Mapに書き込み
func (m *ExMap) Write(key string, value any) {
	m.Sm.Store(key, value)
}

// 排他的Mapを読み込み value.(型名)での変換が必要
func (m *ExMap) Load(key string) (value any, ok bool) {
	value, ok = m.Sm.Load(key)
	return
}

// 排他的Mapに存在するか確認
func (m *ExMap) Check(key string) (ok bool) {
	_, ok = m.Sm.Load(key)
	return
}

// 排他的Mapの削除
func (m *ExMap) Delete(key string) {
	m.Sm.Delete(key)
}
