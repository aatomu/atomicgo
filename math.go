package atomicgo

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func Eval(in string) (ans string, ok bool) {
	in = strings.ReplaceAll(in, " ", "")
	// かっこの数の確認
	if len(strings.Split(in, "(")) != len(strings.Split(in, ")")) {
		return "", false
	}
	// 0-9 +-*/以外のが入ってたらerror
	if !regexp.MustCompile(`^.*[0-9\. +\-*/()]$`).MatchString(in) {
		return "", false
	}
	// 演算用変数 a^b a*b a/b a+b a-b
	var saveEquation string  // 式
	var editEquation string  // 式
	var bracketsDeep int = 0 // かっこ数
	var valueA string        // 変数A
	var valueB string        // 変数B
	var shouldMath bool      // 演算するか

	// 括弧の処理
	for _, s := range strings.Split(in, "") {
		switch {
		case s == "(" && bracketsDeep == 0:
			bracketsDeep++
		case s == "(" && bracketsDeep != 0:
			valueA += s
			bracketsDeep++
		case s == ")" && bracketsDeep == 1:
			result, ok := Eval(valueA)
			if !ok {
				return "", false
			}
			saveEquation += result
			bracketsDeep--
			valueA = ""
		case s == ")" && bracketsDeep != 0:
			valueA += s
			bracketsDeep--
		case bracketsDeep != 0:
			valueA += s
		default:
			saveEquation += s
		}
	}
	// リセット
	valueA, valueB = "", ""
	// いろんな演算
	type CalcData struct {
		Func   func(string, string) string
		Str    string
		Ignore string
	}
	mathTypes := []CalcData{
		{Func: StrPow, Str: "^", Ignore: `[\-+/*]`}, // 累乗
		{Func: StrMul, Str: "*", Ignore: `[\-+/ ]`}, // 乗算
		{Func: StrDiv, Str: "/", Ignore: `[\-+  ]`}, // 除算
		{Func: StrSum, Str: "+", Ignore: `[\-   ]`}, // 和算
		{Func: StrDec, Str: "-", Ignore: `[     ]`}, // 減算
	}
	// 演算
	for _, mathData := range mathTypes {
		for _, s := range strings.Split(saveEquation, "") {
			switch shouldMath {
			case false:
				switch {
				case regexp.MustCompile(mathData.Ignore).MatchString(s):
					// 保存
					editEquation += valueA + s
					// リセット
					valueA = ""
					continue
				case s == mathData.Str:
					// valueAの処理に移行
					shouldMath = true
					continue
				default:
					valueA += s
				}
			case true:
				switch {
				case regexp.MustCompile(mathData.Ignore).MatchString(s):
					// 演算 & 保存
					editEquation += mathData.Func(valueA, valueB) + s
					// リセット
					valueA, valueB = "", ""
					// valueAの処理に移行
					shouldMath = false
					continue
				case s == mathData.Str:
					// 演算
					valueA = mathData.Func(valueA, valueB)
					// リセット
					valueB = ""
					continue
				default:
					valueB += s
				}
			}
		}
		// 特殊処理
		switch {
		case valueA != "" && valueB != "":
			// 演算 & 保存
			editEquation += mathData.Func(valueA, valueB)
		case valueA != "":
			editEquation += valueA
		}
		// 移行
		saveEquation = editEquation
		// リセット
		valueA, valueB, shouldMath, editEquation = "", "", false, ""
	}
	//return ans, true
	return saveEquation, true
}

// 配列の最後を参照
func ArrLast(arr []string) int {
	return len(arr) - 1
}

// 累乗
func StrPow(a, b string) string {
	powA, _ := strconv.ParseFloat(a, 64)
	powB, _ := strconv.ParseFloat(b, 64)
	return fmt.Sprintf("%.15f", math.Pow(powA, powB))
}

// 乗算
func StrMul(a, b string) string {
	powA, _ := strconv.ParseFloat(a, 64)
	powB, _ := strconv.ParseFloat(b, 64)
	return fmt.Sprintf("%.15f", powA*powB)
}

// 除算
func StrDiv(a, b string) string {
	powA, _ := strconv.ParseFloat(a, 64)
	powB, _ := strconv.ParseFloat(b, 64)
	return fmt.Sprintf("%.15f", powA/powB)
}

// 和算
func StrSum(a, b string) string {
	powA, _ := strconv.ParseFloat(a, 64)
	powB, _ := strconv.ParseFloat(b, 64)
	return fmt.Sprintf("%.15f", powA+powB)
}

// 除算
func StrDec(a, b string) string {
	powA, _ := strconv.ParseFloat(a, 64)
	powB, _ := strconv.ParseFloat(b, 64)
	return fmt.Sprintf("%.15f", powA-powB)
}
