package expressions

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseExpression(inString string, Vars map[string]float64) EquationElement {
	s1 := cleanUp(inString)
	s2 := strings.ReplaceAll(s1, "  ", " ")
	parts := strings.Split(s2, " ")
	fmt.Println("Parts: ", parts)
	exp := compileExpression(parts, Vars)
	fmt.Print("Finished Compiling: ")
	fmt.Println("expression", exp)
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

	var isFunction = false
	if strings.Contains(token, "(") {
		isFunction = true
	}
	if !isFunction {
		if len(outputQueue) > 0 {
			outputQueue, b = popE(outputQueue)
		}
		if len(outputQueue) > 0 {
			outputQueue, a = popE(outputQueue)
		}
	} else {
		a = nil
		if len(outputQueue) > 0 {
			outputQueue, b = popE(outputQueue)
		}

	}
	var res EquationElement
	res = makeOperator(token, a, b)
	return res, outputQueue
}
func makeOperator(token string, a, b EquationElement) EquationElement {
	//Function vs Operator
	fmt.Println("OPERATOR? ", token)
	if strings.Contains(token, "(") {
		//Functions
		fmt.Println("adding function with arg to output: ", b)
		fmt.Println("A: ", a, "B: ", b)
		switch token {
		case "sin(":
			return &Siner{b}
		case "cos(":
			return &Coser{b}
		default:
			return b //,err
		}
	} else {
		//Operators
		fmt.Println("Adding Operator to output. op. ", token)
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
func compileExpression(tokens []string, Vars map[string]float64) []EquationElement {
	//0 for left 1 for right
	var precidence map[string]int = map[string]int{"+": 2, "-": 2, "/": 3, "*": 3, "^": 14, "f(": 1}
	operatorStack := []string{}
	outputQueue := []EquationElement{}

	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		fmt.Println("====")
		fmt.Println("Token: ", token)
		fmt.Println("outputQueue[ ")
		for _, e := range outputQueue {
			fmt.Println(toStringNice(e))
		}
		fmt.Println("]")
		fmt.Println("operatorStack: ", operatorStack)
		fmt.Println("----")

		if f, err := strconv.ParseFloat(token, 64); err == nil {
			outputQueue = append(outputQueue, &Num{f})
		} else if len(token) == 1 && inMap(token, Vars) {
			//Variable
			outputQueue = append(outputQueue, &Variable{token})
		} else if token == "sin(" {
			//Functions
			fmt.Println("Adding Function: sin(")
			operatorStack = append(operatorStack, "sin(")
		} else if token == "cos(" {
			//Functions
			fmt.Println("Adding Function: cos(")
			operatorStack = append(operatorStack, "cos(")
		} else if token == "+" || token == "-" || token == "*" || token == "/" || token == "^" {
			//Precidence and associativity of the top on stack token
			for {
				if len(operatorStack) > 0 {
					fmt.Println("\t====")
					fmt.Println("\toutputQueue[ ")
					for _, e := range outputQueue {
						fmt.Println("\t", toStringNice(e))
					}
					fmt.Println("\t]")
					fmt.Println("\toperatorStack: ", operatorStack)
					fmt.Println("\t----")

					topOfStack := operatorStack[len(operatorStack)-1]
					fmt.Println("Top of stack: ", topOfStack)

					tOSQ := topOfStack[:] //top of stack query
					//Check if its a function
					if strings.Contains(tOSQ, "(") {
						tOSQ = "f("
					}
					fmt.Println("Top of Stack Query: ", tOSQ)
					p := precidence[topOfStack]
					thirdOption := (p%10 == precidence[token]%10 && p < 10)
					if ((p%10 > precidence[token]%10) || thirdOption) && topOfStack != "(" {
						//Something something if it gets to the sin(
						var op string
						operatorStack, op = pop(operatorStack)
						//fmt.Println("in:", outputQueue)
						var element EquationElement
						element, outputQueue = assembleOperator(op, outputQueue)
						//fmt.Println("out:", outputQueue)
						outputQueue = append(outputQueue, element)
					} else {
						break
					}
				} else {
					break
				}
			}
			fmt.Println("adding operator", token)
			operatorStack = append(operatorStack, token)
		} else if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			//TOMORROW
			//FIGURE OUT WHY IT DOESNT STOP ONCE IT DOES SIN AND DOES +
			//WHAT IT SHOULD DO IS DO SIN THEN GET THE NEXT TOKEN (/) THEN DO /, THEN DO +
			//PROBLEMS OCCUR WHEN OPERATOR STACK IS NOT EMPTY BEHIND FUNCTION
			//IN THE SECOND SET OF ATTRACTORS ON THE WEBSITE IT COMPILES FINE
			//BUT 1+SIN(X)/B still gets compiled wrong because + is still on the stack when sin gets switched to the queue
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {

				fmt.Println("\t====")
				fmt.Println("\toutputQueue[ ")
				for _, e := range outputQueue {
					fmt.Println("\t", toStringNice(e))
				}
				fmt.Println("\t]")
				fmt.Println("\toperatorStack: ", operatorStack)
				fmt.Println("\t----")

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
					fmt.Println("Ending Function: ", op)
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
		fmt.Println("\t====")
		fmt.Println("\toutputQueue[ ")
		for _, e := range outputQueue {
			fmt.Println("\t", toStringNice(e))
		}
		fmt.Println("\t]")
		fmt.Println("\toperatorStack: ", operatorStack)
		fmt.Println("\t----")

		var op string
		//fmt.Println("InEndOp: ", operatorStack)
		//fmt.Println("InEndQ: ", outputQueue)

		operatorStack, op = pop(operatorStack)
		var element EquationElement
		element, outputQueue = assembleOperator(op, outputQueue)

		//fmt.Println("OutEndOp: ", operatorStack)
		//fmt.Println("InEndQ: ", outputQueue)

		outputQueue = append(outputQueue, element)
		//fmt.Println("Out2EndQ: ", outputQueue)

	}
	fmt.Println("----@End----")
	fmt.Println("outputQueue: ")
	for _, e := range outputQueue {
		fmt.Println(toStringNice(e))
	}
	fmt.Println("operatorStack: ", operatorStack)
	fmt.Println("----")

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

//Actually just shunting car
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
	BecomeNumber(map[string]float64) float64
	BecomeString() string
}

func toStringNice(a EquationElement) string {
	if a == nil {
		return "{nil}"
	} else {
		return a.BecomeString()
	}
}

type Variable struct {
	name string
}

func (V Variable) BecomeNumber(Vars map[string]float64) float64 {
	return Vars[V.name]
}
func (V Variable) BecomeString() string {
	return V.name
}

//The Siner
type Siner struct {
	a EquationElement
}

func (S *Siner) BecomeNumber(Vars map[string]float64) float64 {
	return math.Sin(S.a.BecomeNumber(Vars))
}
func (S *Siner) BecomeString() string {
	return "sin(" + toStringNice(S.a) + ")"
}

//The Coser
type Coser struct {
	a EquationElement
}

func (C *Coser) BecomeNumber(Vars map[string]float64) float64 {
	return math.Cos(C.a.BecomeNumber(Vars))
}
func (C *Coser) BecomeString() string {
	return "cos(" + toStringNice(C.a) + ")"
}

//The Adder
type Adder struct {
	a, b EquationElement
}

func (Add *Adder) BecomeNumber(Vars map[string]float64) float64 {
	return Add.a.BecomeNumber(Vars) + Add.b.BecomeNumber(Vars)
}
func (Add *Adder) BecomeString() string {
	return "(" + toStringNice(Add.a) + "+" + toStringNice(Add.b) + ")"
}

//The Subtractor
type Subtractor struct {
	a, b EquationElement
}

func (Sub *Subtractor) BecomeNumber(Vars map[string]float64) float64 {
	return Sub.a.BecomeNumber(Vars) - Sub.b.BecomeNumber(Vars)
}
func (Sub *Subtractor) BecomeString() string {
	return "(" + toStringNice(Sub.a) + "-" + toStringNice(Sub.b) + ")"
}

//The Multiplier
type Multiplier struct {
	a, b EquationElement
}

func (Mul *Multiplier) BecomeNumber(Vars map[string]float64) float64 {
	return Mul.a.BecomeNumber(Vars) * Mul.b.BecomeNumber(Vars)
}
func (Mul *Multiplier) BecomeString() string {
	return "(" + toStringNice(Mul.a) + "*" + toStringNice(Mul.b) + ")"
}

//The Divider
type Divider struct {
	a, b EquationElement
}

func (Div *Divider) BecomeNumber(Vars map[string]float64) float64 {
	return Div.a.BecomeNumber(Vars) / Div.b.BecomeNumber(Vars)
}
func (Div *Divider) BecomeString() string {
	return "(" + toStringNice(Div.a) + "/" + toStringNice(Div.b) + ")"
}

//Base Number
type Num struct {
	n float64
}

func (n *Num) BecomeNumber(Vars map[string]float64) float64 {
	return n.n
}
func (n *Num) BecomeString() string {
	return fmt.Sprint(n.n)
}
