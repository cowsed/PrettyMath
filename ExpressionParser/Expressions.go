package expressions

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

//ParseExpression takes in a string, cleans it, tokenizes it and converts it to a calculatable structure of ExpressionElements
func ParseExpression(inString string, Vars map[string]float64) ExpressionElement {
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
func assembleOperator(token string, outputQueue []ExpressionElement) (ExpressionElement, []ExpressionElement) {
	//Stick together elements into biger element
	var a ExpressionElement = nil
	var b ExpressionElement = nil

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
	var res ExpressionElement
	res = makeOperator(token, a, b)
	return res, outputQueue
}
func makeOperator(token string, a, b ExpressionElement) ExpressionElement {
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
func compileExpression(tokens []string, Vars map[string]float64) []ExpressionElement {
	//0 for left 1 for right
	var precidence map[string]int = map[string]int{"+": 2, "-": 2, "/": 3, "*": 3, "^": 14, "f(": 1}
	operatorStack := []string{}
	outputQueue := []ExpressionElement{}

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
						var element ExpressionElement
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
				var element ExpressionElement
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
					var element ExpressionElement
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
		var element ExpressionElement
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

//Helper popping functionns to make writing a bit easier - needs reslicing which is apparently bad but whatever
func popE(sl []ExpressionElement) ([]ExpressionElement, ExpressionElement) {
	if len(sl) > 0 {
		res := sl[len(sl)-1]
		sl = sl[:len(sl)-1]
		return sl, res
	}
	return sl, nil

}
func pop(sl []string) ([]string, string) {
	if len(sl) > 0 {
		res := sl[len(sl)-1]
		sl = sl[:len(sl)-1]
		return sl, res
	}
	return sl, ""

}

//Cleans the string and separates it into an easily tokenizable form
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

//ExpressionElement is an interface for part of a expression
//In the future also have a better way to register this one of these into the parser
type ExpressionElement interface {
	BecomeNumber(map[string]float64) float64
	BecomeString() string
}

//becomes a string but doesnt error if its nli - mostly for debugging
func toStringNice(a ExpressionElement) string {
	if a == nil {
		return "{nil}"
	}
	return a.BecomeString()

}

//Variable is an ExpressionElement that references a variable map when called upon to calculate
type Variable struct {
	name string
}

//BecomeNumber causes this to calculate itself
func (V Variable) BecomeNumber(Vars map[string]float64) float64 {
	return Vars[V.name]
}

//BecomeString creates a string representation of this element and its children
func (V Variable) BecomeString() string {
	return V.name
}

//Siner is an ExpressionElement that returns the sine of its sub element
type Siner struct {
	a ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (S *Siner) BecomeNumber(Vars map[string]float64) float64 {
	return math.Sin(S.a.BecomeNumber(Vars))
}

//BecomeString creates a string representation of this element and its children
func (S *Siner) BecomeString() string {
	return "sin(" + toStringNice(S.a) + ")"
}

//Coser is an ExpressionElement that returns the cosine of its sub element
type Coser struct {
	a ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (C *Coser) BecomeNumber(Vars map[string]float64) float64 {
	return math.Cos(C.a.BecomeNumber(Vars))
}

//BecomeString creates a string representation of this element and its children
func (C *Coser) BecomeString() string {
	return "cos(" + toStringNice(C.a) + ")"
}

//Adder adds its two sub elements
type Adder struct {
	a, b ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (Add *Adder) BecomeNumber(Vars map[string]float64) float64 {
	return Add.a.BecomeNumber(Vars) + Add.b.BecomeNumber(Vars)
}

//BecomeString creates a string representation of this element and its children
func (Add *Adder) BecomeString() string {
	return "(" + toStringNice(Add.a) + "+" + toStringNice(Add.b) + ")"
}

//Subtractor subtracts its two sub elements
type Subtractor struct {
	a, b ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (Sub *Subtractor) BecomeNumber(Vars map[string]float64) float64 {
	return Sub.a.BecomeNumber(Vars) - Sub.b.BecomeNumber(Vars)
}

//BecomeString creates a string representation of this element and its children
func (Sub *Subtractor) BecomeString() string {
	return "(" + toStringNice(Sub.a) + "-" + toStringNice(Sub.b) + ")"
}

//Multiplier multiplies its two sub elements
type Multiplier struct {
	a, b ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (Mul *Multiplier) BecomeNumber(Vars map[string]float64) float64 {
	return Mul.a.BecomeNumber(Vars) * Mul.b.BecomeNumber(Vars)
}

//BecomeString creates a string representation of this element and its children
func (Mul *Multiplier) BecomeString() string {
	return "(" + toStringNice(Mul.a) + "*" + toStringNice(Mul.b) + ")"
}

//Divider divides its two sub elements
type Divider struct {
	a, b ExpressionElement
}

//BecomeNumber causes this to calculate itself
func (Div *Divider) BecomeNumber(Vars map[string]float64) float64 {
	return Div.a.BecomeNumber(Vars) / Div.b.BecomeNumber(Vars)
}

//BecomeString creates a string representation of this element and its children
func (Div *Divider) BecomeString() string {
	return "(" + toStringNice(Div.a) + "/" + toStringNice(Div.b) + ")"
}

//Num is just a constant number
type Num struct {
	n float64
}

//BecomeNumber causes this to calculate itself
func (n *Num) BecomeNumber(Vars map[string]float64) float64 {
	return n.n
}

//BecomeString creates a string representation of this element and its children
func (n *Num) BecomeString() string {
	return fmt.Sprint(n.n)
}
