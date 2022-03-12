package atomicgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"math"
	"strconv"
)

func Calc(in string) (ans float64, err error) {
	// 文字列チェック
	if !StringCheck(in, `^[0-9.+\-*/^\(\)]+$`) {
		return 0, fmt.Errorf("contains unknown characters")
	}
	// 式を木に
	expr, err := parser.ParseExpr(in)
	if err != nil {
		return 0, err
	}

	// 演算
	ans = calc(expr)
	return
}

func calc(expr ast.Expr) (result float64) {
	// 分岐
	switch e := expr.(type) {
	// 数値
	case *ast.BasicLit:
		fmt.Println("Number    :", e.Value)
		return Str2Float(e.Value)
	// 数式
	case *ast.BinaryExpr:
		// 変数
		var x, y float64
		// X
		x = calc(e.X)
		// Y
		y = calc(e.Y)
		type op struct {
			f func(float64, float64) float64
		}
		Operation := map[string]op{}
		Operation["+"] = op{f: func(a, b float64) float64 { return a + b }}
		Operation["-"] = op{f: func(a, b float64) float64 { return a - b }}
		Operation["*"] = op{f: func(a, b float64) float64 { return a * b }}
		Operation["/"] = op{f: func(a, b float64) float64 { return a / b }}
		Operation["^"] = op{f: func(a, b float64) float64 { return math.Pow(a, b) }}
		fmt.Println("Elucation :", x, e.Op, y, "=", Operation[e.Op.String()].f(x, y))
		return Operation[e.Op.String()].f(x, y)
	// 括弧内の式
	case *ast.ParenExpr:
		fmt.Println("Children  :", e.X)
		return calc(e.X)
	// 不明
	default:
		fmt.Printf("Unknown :%T\n", e)
	}
	return
}

//変換
func Str2Float(in string) (float float64) {
	float, _ = strconv.ParseFloat(in, 64)
	return
}
