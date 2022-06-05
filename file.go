package atomicgo

import (
	"bufio"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// GoのPathを入手 最後に/がつく
func GetGoDir() (goDir string) {
	_, file, _, _ := runtime.Caller(1)
	goDir = filepath.Dir(file) + "/"
	return
}

// 実行ディレクトリ変更
func MoveWorkDir(dirPath string) {
	os.Chdir(dirPath)
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
	file, err := os.Create(filePath)

	if PrintError("Failed Open File", err) {
		return false
	}
	// 自動で閉じる
	defer file.Close()

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
func FileList(dir string) (list []string, success bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}, false
	}

	for _, file := range files {
		if file.IsDir() {
			result, ok := FileList(filepath.Join(dir, file.Name()))
			list = append(list, result...)
			if !ok {
				return list, false
			}
			continue
		}
		list = append(list, filepath.Join(dir, file.Name()))
	}

	return list, true
}
