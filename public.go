package atomicgo

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
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

// GoのPathを入手 最後に/がつく
func GetGoDir() (goDir string) {
	_, callerFile, _, _ := runtime.Caller(0)
	goDir = filepath.Dir(callerFile) + "/"
	return
}

// 実行ディレクトリ変更
func MoveWorkDir(dirPath string) {
	os.Chdir(dirPath)
}

// プログラムの実行準備 IOを返却 io.Start()で実行
//Linuxなら/bin/bash,-c,<Command>を
//WinならC:\Windows\System32\cmd.exe,/c,<Command>
func ExecuteCommand(pront, option, command string) (io *exec.Cmd) {
	return exec.Command(pront, option, command)
}

// IOを分けて返却
func GetCommandIo(io *exec.Cmd) (inIo io.WriteCloser, outIo io.ReadCloser, errIo io.ReadCloser) {
	inIo, _ = io.StdinPipe()
	outIo, _ = io.StdoutPipe()
	errIo, _ = io.StderrPipe()
	return
}

// IO書き込み
func IoWrite(ioIn io.WriteCloser, text string) {
	io.WriteString(ioIn, text+"\n")
}

// 正規表現チェック
func StringCheck(text string, check string) (success bool) {
	return regexp.MustCompile(check).MatchString(text)
}

// 正規表現書き換え
func StringReplace(fromText string, toText string, check string) (replaced string) {
	return regexp.MustCompile(check).ReplaceAllString(fromText, toText)
}

// 乱数生成
func RandomGenerate(max int) (result int) {
	rand.Seed(time.Now().UnixNano())
	result = rand.Int() % max
	return
}

// maxまでstringを切る
func StringCut(text string, max int) (result string) {
	//文字数を制限
	textArray := strings.Split(text, "")
	if len(textArray) < max {
		return text
	}
	for i := 0; i < max; i++ {
		result = result + textArray[i]
	}
	return
}

// ファイル,フォルダーチェック
func CheckFile(filePath string) (ok bool) {
	// filePathからアクセスできるかチェック
	_, err := os.Stat(filePath)
	return err == nil
}

// ファイル作成
func CreateFile(filePath string) (success bool) {
	_, err := os.Create(filePath)
	return !PrintError("Failed Create File", err)
}

// ディレクトリ作成
func CreateDir(dirPath string, perm fs.FileMode) (success bool) {
	err := os.Mkdir(dirPath, perm)
	return !PrintError("Failed Create Directory", err)
}

// ファイル読み込み 一括
func ReadFile(filePath string) (data []byte, success bool) {
	// 読み込み
	data, err := ioutil.ReadFile(filePath)
	if PrintError("Failed Read File", err) {
		return nil, false
	}
	return data, true
}

// ファイル書き込み 一括
func WriteFileFlash(filePath string, data []byte, perm fs.FileMode) (success bool) {
	err := ioutil.WriteFile(filePath, data, perm)
	return !PrintError("Failed Write File Flash", err)
}

// ファイル書き込み バッファーあり
func WriteFileBaffer(filePath string, data []byte, perm fs.FileMode) (success bool) {
	// ファイルを開く
	file, err := os.Open(filePath)
	// 自動で閉じる
	defer file.Close()

	if PrintError("Failed Open File", err) {
		return false
	}

	// 書き込み
	fileWriter := bufio.NewWriter(file)
	_, err = fileWriter.Write(data)
	if PrintError("Failed White in Buffer", err) {
		return false
	}

	// 残りを書き込み
	err = fileWriter.Flush()
	return !PrintError("Failed White in Flash", err)
}

// ファイル一覧
func FileList(dir string) (list string, faild bool) {
	//ディレクトリ読み取り
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		PrintError("Failed read directory data", err)
		return "", false
	}

	//一覧を保存
	for _, file := range files {
		//ディレクトリなら一個下でやる
		if file.IsDir() {
			data, ok := FileList(dir + "/" + file.Name())
			if !ok {
				PrintError("Failed func fileList()", err)
				return "", false
			}
			//追加
			list = list + data
			continue
		}
		list = list + dir + "/" + file.Name() + "\n"
	}

	list = strings.ReplaceAll(list, "//", "/")
	return list, true
}

// 特定のシグナルを受けるまで終了しない
func StopWait() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Error表示
func PrintError(message string, err error) (errored bool) {
	if err != nil {
		trackBack := ""
		// 原因を特定
		for i := 1; true; i++ {
			pc, file, line, _ := runtime.Caller(i)
			trackBack += fmt.Sprintf("> %s:%d %s()\n", filepath.Base(file), line, StringReplace(runtime.FuncForPC(pc).Name(), "", "^.*/"))
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
	sync.Map
}

// 排他的Mapを入手
// 排他的MAPの型 : sync.Map
func ExMapGet() *ExMap {
	return &ExMap{}
}

// 排他的Mapに書き込み
func (m *ExMap) ExMapWrite(key string, value interface{}) {
	m.Store(key, value)
}

// 排他的Mapを読み込み value.(型名)での変換が必要
func (m *ExMap) ExMapLoad(key string) (value interface{}, ok bool) {
	value, ok = m.Load(key)
	return
}

// 排他的Mapに存在するか確認
func (m *ExMap) ExMapCheck(key string) (ok bool) {
	_, ok = m.Load(key)
	return
}

//排他的Mapの削除
func (m *ExMap) ExMapDelete(key string) {
	m.Delete(key)
}
