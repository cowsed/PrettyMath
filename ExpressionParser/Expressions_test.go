package expressions_test

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"./."
)

func TestParseExpression1Result(t *testing.T) {
	exp := "cos(2-2)"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, float64(1), got.BecomeNumber(map[string]float64{}))
}
func TestParseExpression1Eq(t *testing.T) {
	exp := "cos(2-2)"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, "cos((2-2))", got.BecomeString())
}

func TestParseExpression2Result(t *testing.T) {
	exp := "1+3/6"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, float64(1.5), got.BecomeNumber(map[string]float64{}))
}
func TestParseExpression2Eq(t *testing.T) {
	exp := "1+3/6"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, "(1+(3/6))", got.BecomeString())
}

func TestParseExpression3Result(t *testing.T) {
	exp := "a+b"
	vs := map[string]float64{"a": 2, "b": 3}
	got := expressions.ParseExpression(exp, vs)
	equals(t, float64(5), got.BecomeNumber(vs))
}
func TestParseExpression3Eq(t *testing.T) {
	exp := "a+b"
	got := expressions.ParseExpression(exp, map[string]float64{"p": 3.1415})
	equals(t, "(a+b)", got.BecomeString())
}

func TestParseExpression4Result(t *testing.T) {
	exp := "x+sin(y)/b"
	vs := map[string]float64{"x": 2, "y": 0, "b": 3}
	got := expressions.ParseExpression(exp, vs)
	equals(t, float64(2), got.BecomeNumber(vs))
}
func TestParseExpression4Eq(t *testing.T) {
	exp := "x+sin(y)/b"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, "(x+(sin(y)/b))", got.BecomeString())
}

func TestParseExpression5Result(t *testing.T) {
	exp := "1+sin(2)/3"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, float64(1+math.Sin(2)/3), got.BecomeNumber(map[string]float64{}))
}
func TestParseExpression5Eq(t *testing.T) {
	exp := "1+sin(2)/3"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, "(1+(sin(2)/3))", got.BecomeString())
}
func TestParseExpression6Result(t *testing.T) {
	exp := "1+(2)/3"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, fmt.Sprintf("%.3f", float64(1+(2/3.0))), fmt.Sprintf("%.3f", got.BecomeNumber(map[string]float64{})))
}
func TestParseExpression6Eq(t *testing.T) {
	exp := "1+(2)/3"
	got := expressions.ParseExpression(exp, map[string]float64{})
	equals(t, "(1+(2/3))", got.BecomeString())
}

/*
func TestShuntingCar1(t *testing.T){
	exp:=[]string{"2","2","+"}
	res:=expressions.ShuntingCar([]string{"2", "+", "2"})
	equals(t, exp,res)
}

func TestShuntingCar2(t *testing.T){
	fmt.Println("Starting Shunting car on {\"sin(\" ,\"2\" , \")\"}")
	exp:=[]string{"sin(","2",")"}
	res:=expressions.ShuntingCar([]string{"2", "sin"})
	equals(t, exp,res)
}
*/

//Testing Helpers

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
