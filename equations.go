package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var vars = map[string]float64{"x": 1.0, "y": 2.0, "p": math.Pi}

func parseExpression(inString string) EquationElement {
	//s1:=inString//Replace Variables
	fmt.Println("Vars: ", vars)
	s1 := cleanUp(inString)
	s2 := strings.ReplaceAll(s1, "  ", " ")
	parts := strings.Split(s2, " ")
	fmt.Println("Parts: ", parts)
	exp := CompileExpression(parts)
	fmt.Print("Finished Compiling: ")
	fmt.Println(exp[0])
	fmt.Println(exp[0].BecomeString())
	return exp[0]
}

func inMap(key string, m map[string]float64) bool {
	_, in := m[key]
	return in
}
func assembleOperator(token string, outputQueue []EquationElement) (EquationElement, []EquationElement) {
	//Stick together elements into biger element
	var a EquationElement = nil
	var b EquationElement = nil

	if len(outputQueue) > 0 {
		outputQueue, b = popE(outputQueue)
	}
	if len(outputQueue) > 0 {
		outputQueue, a = popE(outputQueue)
	}
	var res EquationElement
	res = makeOperator(token, a, b)
	return res, outputQueue
}
func makeOperator(token string, a, b EquationElement) EquationElement {
	//Function vs Operator
	if a == nil {
		//Functions
		fmt.Println("adding function with arg: ", b)
		switch token {
		case "sin":
			return &Siner{b}
		case "cos":
			return &Coser{b}
		default:
			return b //,err
		}
	} else {
		//Operators
		fmt.Println("Adding Operator")
		switch token {
		case "+":
			return &Adder{a, b}
		case "-":
			return &Subtractor{a, b}
		case "*":
			return &Multiplier{a, b}
		case "/":
			return &Divider{a, b}
		}
		return a //,err
	}

}

//A Variant of shunting car
func CompileExpression(tokens []string) []EquationElement {
	//0 for left 1 for right
	var precidence map[string]int = map[string]int{"+": 2, "-": 2, "/": 3, "*": 3, "^": 15}
	operatorStack := []string{}
	outputQueue := []EquationElement{}

	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		fmt.Println("Token: ", token)
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			outputQueue = append(outputQueue, &Num{f})
		} else if len(token) == 1 && inMap(token, vars) {
			//Variable
			outputQueue = append(outputQueue, &Variable{token})
		} else if token == "sin(" {
			//Functions
			operatorStack = append(operatorStack, "sin")
		} else if token == "cos(" {
			//Functions
			operatorStack = append(operatorStack, "cos")
		} else if token == "+" || token == "-" || token == "*" || token == "/" || token == "^" {
			//Precidence and associativity of the top on stack token
			for {
				if len(operatorStack) > 0 {
					topOfStack := operatorStack[len(operatorStack)-1]
					fmt.Println("Top of stack: ", topOfStack)
					p := precidence[topOfStack]
					thirdOption := (p%10 == precidence[token]%10 && p < 10) && topOfStack != "("
					if (p%10 > precidence[token]%10) || thirdOption {
						var op string
						operatorStack, op = pop(operatorStack)
						fmt.Println("in:", outputQueue)
						var element EquationElement
						element, outputQueue = assembleOperator(op, outputQueue)
						fmt.Println("out:", outputQueue)
						outputQueue = append(outputQueue, element)
					} else {
						break
					}
				} else {
					break
				}
			}
			fmt.Println("adding token", token)
			operatorStack = append(operatorStack, token)
		} else if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				var op string
				operatorStack, op = pop(operatorStack)
				var element EquationElement
				element, outputQueue = assembleOperator(op, outputQueue)
				outputQueue = append(outputQueue, element)

			}
			//while the operator at the top of the operator stack is not a left parenthesis:
			/* If the stack runs out without finding a left parenthesis, then there are mismatched parentheses. */
			if len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] == "(" {
				operatorStack, _ = pop(operatorStack)
			}

			if len(operatorStack) > 0 {
				op := operatorStack[len(operatorStack)-1]
				if op == "sin(" || op == "cos(" {
					var op string
					operatorStack, op = pop(operatorStack)
					var element EquationElement
					element, outputQueue = assembleOperator(op, outputQueue)
					outputQueue = append(outputQueue, element)
				}
			}
			//if there is a function token at the top of the operator stack, then:
			//    pop the function from the operator stack onto the output queue.
		}

	}
	for len(operatorStack) > 0 {

		var op string
		fmt.Println("InEndOp: ", operatorStack)
		fmt.Println("InEndQ: ", outputQueue)

		operatorStack, op = pop(operatorStack)
		var element EquationElement
		element, outputQueue = assembleOperator(op, outputQueue)

		fmt.Println("OutEndOp: ", operatorStack)
		fmt.Println("InEndQ: ", outputQueue)

		outputQueue = append(outputQueue, element)
		fmt.Println("Out2EndQ: ", outputQueue)

	}

	return outputQueue
}
func popE(sl []EquationElement) ([]EquationElement, EquationElement) {
	if len(sl) > 0 {
		res := sl[len(sl)-1]
		sl = sl[:len(sl)-1]
		return sl, res
	}
	return sl, nil

}

func ShuntingCar(tokens []string) []string {
	//0 for left 1 for right
	var precidence map[string]int = map[string]int{"+": 2, "-": 2, "/": 3, "*": 3, "^": 15}
	operatorStack := []string{}
	outputQueue := []string{}

	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			outputQueue = append(outputQueue, token) // &Num{f})
		} else if token == "sin(" {
			//Functions
			operatorStack = append(operatorStack, "sin(")
		} else if token == "cos(" {
			//Functions
			operatorStack = append(operatorStack, "sin")
		} else if token == "+" || token == "-" || token == "*" || token == "/" || token == "^" {
			//Precidence and associativity of the top on stack token
			for {
				if len(operatorStack) > 0 {
					topOfStack := operatorStack[len(operatorStack)-1]
					p := precidence[topOfStack]
					thirdOption := (p%10 == precidence[token]%10 && p < 10) && topOfStack != "("
					if (p%10 > precidence[token]%10) || thirdOption {
						var op string
						operatorStack, op = pop(operatorStack)
						outputQueue = append(outputQueue, op)
					} else {
						break
					}
				} else {
					break
				}
			}
			fmt.Println("adding token", token)
			operatorStack = append(operatorStack, token)
		} else if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				var op string
				operatorStack, op = pop(operatorStack)
				outputQueue = append(outputQueue, op)

			}
			//while the operator at the top of the operator stack is not a left parenthesis:
			/* If the stack runs out without finding a left parenthesis, then there are mismatched parentheses. */
			if len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] == "(" {
				operatorStack, _ = pop(operatorStack)
			}

			if len(operatorStack) > 0 {
				op := operatorStack[len(operatorStack)-1]
				if op == "sin(" || op == "cos(" {
					var op string
					operatorStack, op = pop(operatorStack)
					outputQueue = append(outputQueue, op)
				}
			}
			//if there is a function token at the top of the operator stack, then:
			//    pop the function from the operator stack onto the output queue.
		}

	}
	for len(operatorStack) > 0 {

		var op string
		operatorStack, op = pop(operatorStack)
		fmt.Println(operatorStack)
		outputQueue = append(outputQueue, op)

	}

	return outputQueue
}
func pop(sl []string) ([]string, string) {
	if len(sl) > 0 {
		res := sl[len(sl)-1]
		sl = sl[:len(sl)-1]
		return sl, res
	}
	return sl, ""

}

func findBetweenParen(s []string) ([]string, int) {
	paren := 1
	for i, element := range s {

		if element == "(" {
			paren++
		} else if element == ")" {
			paren--
		}
		fmt.Println("when paren searching enc. ", element, " #p: ", paren)
		if paren == 0 {
			fmt.Println("Found: ", s[0:i], "end: ", i)
			return s[0:i], i
		}
	}
	return []string{"AHH EROOOR"}, -1
}

func cleanUp(start string) string {
	end := ""
	for i, r := range start {
		c := string(r)
		notIntoFunc := true
		if i > 2 {
			notIntoFunc = (start[i-3:i+1] != "sin(" && start[i-3:i+1] != "cos(")
		}
		if c == "(" || c == ")" || c == "+" || c == "-" || c == "*" || c == "/" || c == "^" {

			if (i > 0 && string(start[i-1]) != " ") && notIntoFunc {
				end += " "
			}
			end += c
			if i < len(start)-1 && string(start[i+1]) != " " {
				end += " "
			}
		} else {
			end += c
		}
	}
	return end
}

type EquationElement interface {
	BecomeNumber() float64
	BecomeString() string
	AcceptsSecond() bool
	AddSecond(EquationElement)
}
type Variable struct {
	name string
}

func (V Variable) BecomeNumber() float64 {
	return vars[V.name]
}
func (V Variable) BecomeString() string {
	return V.name
}

func (V *Variable) AcceptsSecond() bool              { return false }
func (V *Variable) AddSecond(second EquationElement) {}

//The Siner
type Siner struct {
	a EquationElement
}

func (S *Siner) BecomeNumber() float64 {
	return math.Sin(S.a.BecomeNumber())
}
func (S *Siner) BecomeString() string {
	return "sin(" + S.a.BecomeString() + ")"
}

func (S *Siner) AcceptsSecond() bool              { return true }
func (S *Siner) AddSecond(second EquationElement) { S.a = second }

//The Coser
type Coser struct {
	a EquationElement
}

func (C *Coser) BecomeNumber() float64 {
	return math.Cos(C.a.BecomeNumber())
}
func (C *Coser) BecomeString() string {
	return "cos" + C.a.BecomeString() + ")"
}
func (C *Coser) AcceptsSecond() bool              { return true }
func (C *Coser) AddSecond(second EquationElement) { C.a = second }

//The Adder
type Adder struct {
	a, b EquationElement
}

func (Add *Adder) BecomeNumber() float64 {
	return Add.a.BecomeNumber() + Add.b.BecomeNumber()
}
func (Add *Adder) BecomeString() string {
	return "(" + Add.a.BecomeString() + "+" + Add.b.BecomeString() + ")"
}

func (Add *Adder) AcceptsSecond() bool              { return true }
func (Add *Adder) AddSecond(second EquationElement) { Add.b = second }

//The Subtractor
type Subtractor struct {
	a, b EquationElement
}

func (Sub *Subtractor) BecomeNumber() float64 {
	return Sub.a.BecomeNumber() - Sub.b.BecomeNumber()
}
func (Sub *Subtractor) BecomeString() string {
	return "(" + Sub.a.BecomeString() + "-" + Sub.b.BecomeString() + ")"
}
func (Sub *Subtractor) AcceptsSecond() bool              { return true }
func (Sub *Subtractor) AddSecond(second EquationElement) { Sub.b = second }

//The Multiplier
type Multiplier struct {
	a, b EquationElement
}

func (Mul *Multiplier) BecomeNumber() float64 {
	//fmt.Println(Mul.a.BecomeNumber()," and* ",Mul.b.BecomeNumber())
	return Mul.a.BecomeNumber() * Mul.b.BecomeNumber()
}
func (Mul *Multiplier) BecomeString() string {
	return "(" + Mul.a.BecomeString() + "*" + Mul.b.BecomeString() + ")"
}

func (Mul *Multiplier) AcceptsSecond() bool              { return true }
func (Mul *Multiplier) AddSecond(second EquationElement) { Mul.b = second }

//The Divider
type Divider struct {
	a, b EquationElement
}

func (Div *Divider) BecomeNumber() float64 {
	return Div.a.BecomeNumber() / Div.b.BecomeNumber()
}
func (Div *Divider) BecomeString() string {
	return "(" + Div.a.BecomeString() + "/" + Div.b.BecomeString() + ")"
}
func (Div *Divider) AcceptsSecond() bool              { return true }
func (Div *Divider) AddSecond(second EquationElement) { Div.b = second }

//Base Number
type Num struct {
	n float64
}

func (n *Num) BecomeNumber() float64 {
	return n.n
}
func (n *Num) BecomeString() string {
	return fmt.Sprint(n.n)
}

func (n *Num) AcceptsSecond() bool              { return false }
func (n *Num) AddSecond(second EquationElement) {}
