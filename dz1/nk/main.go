package main

import (
	"fmt"
	"math"
	"strconv"
)

type Var struct {
	s string
	Number int
	Value bool
}

// Функция преобразования переменной в удобочитаемый вид
func (v Var) PrettyString() string {
	formatted := fmt.Sprintf("x%d", v.Number)
	if !v.Value {
		return "!" + formatted
	}
	return formatted
}

func (v Var) String() string {
	if v.s == "" {
		v.s = fmt.Sprintf("(%d,%t)", v.Number, v.Value)
	}
	return v.s
}

// Структура представляет собой неопределенный коэффициент
type K []Var

func (a K) PrettyString() string {
	var prettyString string
	for _, bit := range a {
		prettyString = bit.PrettyString() + prettyString
	}
	return prettyString
}

func (a K) KString() string {
	var foramtted string
	foramtted += "K_("
	for _, v := range a {
		foramtted += strconv.Itoa(v.Number)
	}
	foramtted += ")^("
	for _, v := range a {
		if v.Value {
			foramtted += "1"
		} else {
			foramtted += "0"
		}
	}
	foramtted += ")"
	return foramtted
}

func (a K) String() string {
	var stringed string
	for _, v := range a {
		stringed += v.String()
	}
	return stringed
}

func (a K) Equal(b K) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Number != b[i].Number || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func (a K) Less(b K) bool {
	return len(a) < len(b)
}

func (a K) AppendVar(v Var) K {
	newK := make(K, len(a))
	copy(newK, a)
	newK = append(newK, v)
	return newK
}

func (a K) IsCovers(b K) bool {
	for i := range a {
		var found = false
		for j := range b {
			if a[i].Number == b[j].Number {
				if a[i].Value != b[j].Value {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type Term []bool

func (t Term) MakeAllSubsets() []K {
	allSubsets := make([]K, 0, int(math.Pow(2, float64(len(t)))))
	for number, value := range t {
		v := Var{
			Number: number,
			Value:  value,
		}
		for _, subset := range allSubsets {
			allSubsets = append(allSubsets, subset.AppendVar(v))
		}
		allSubsets = append(allSubsets, K{v})
	}
	return allSubsets
}

type Equation struct {
	Coefficients []K
	Value bool
}

func (e Equation) KString() string {
	var formatted string
	for i, k := range e.Coefficients {
		if i != 0 {
			formatted += " + "
		}
		formatted += k.KString()
	}
	formatted += " = "
	if e.Value {
		formatted += "1"
	} else {
		formatted += "0"
	}
	return formatted
}

func (e Equation) Include(q Equation) bool {
	for _, qk := range q.Coefficients {
		var found = false
		for _, ek := range e.Coefficients {
			if qk.Equal(ek) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (e *Equation) ExcludeCoefficient(k K) {
	newCoefficients := make([]K, 0, len(e.Coefficients))
	for _, old := range e.Coefficients {
		if !old.Equal(k) {
			newCoefficients = append(newCoefficients, old)
		}
	}
	e.Coefficients = newCoefficients
}

// Функция возвращает СДНФ от ФАЛ
func MakeSystemOfEquations(f []int) []Equation {
	variableNumber := int(math.Log2(float64(len(f))))
	format := "%0" + strconv.Itoa(variableNumber) + "b"

	equations := make([]Equation, 0, len(f))
	for i := range f {
		term := make(Term, 0, variableNumber)
		binaryString := fmt.Sprintf(format, i)
		for _, char := range binaryString {
			switch char {
			case '0':
				term = append(term, false)
			case '1':
				term = append(term, true)
			default:
				panic(fmt.Sprintf("unexpected char: %s", string(char)))
			}
		}

		var value bool
		switch f[i] {
		case 0:
			value = false
		case 1:
			value = true
		default:
			panic(fmt.Sprintf("unexpected boolean value: %d", f[i]))
		}

		equations = append(equations, Equation{
			Coefficients: term.MakeAllSubsets(),
			Value:        value,
		})
	}
	return equations
}

func ExcludeZeroCoefficients(system []Equation) []Equation {
	newSystem := make([]Equation, 0, len(system))
	for _, equationToExclude := range system {
		if equationToExclude.Value == true {
			for _, zeroEquation := range system {
				if zeroEquation.Value == false {
					for _, zeroCoefficient := range zeroEquation.Coefficients {
						equationToExclude.ExcludeCoefficient(zeroCoefficient)
					}
				}
			}
			newSystem = append(newSystem, equationToExclude)
		}
	}
	return newSystem
}

type KS []K

func (ks KS) Complexity() int {
	complexity := 0
	for _, r := range ks {
		complexity += len(r)
	}
	return complexity
}

func GetMostRepeated(system []Equation) ([]K, [][]Equation) {
	type repetition struct {
		Count int
		K
	}

	set := make(map[string]repetition)
	for _, equation := range system {
		for _, coefficient := range equation.Coefficients {
			if r, found := set[coefficient.String()]; found {
				r.Count++
				set[coefficient.String()] = r
			} else {
				set[coefficient.String()] = repetition{
					Count: 1,
					K:     coefficient,
				}
			}
		}
	}
	maxRepeat := 0
	for _, rep := range set {
		if rep.Count > maxRepeat {
			maxRepeat = rep.Count
		}
	}

	minSize := math.MaxInt32
	for _, rep := range set {
		if rep.Count == maxRepeat {
			if len(rep.K) < minSize {
				minSize = len(rep.K)
			}
		}
	}

	var withMaxRepeats []K
	for _, rep := range set {
		if rep.Count == maxRepeat && len(rep.K) == minSize {
			withMaxRepeats = append(withMaxRepeats, rep.K)
		}
	}

	var newSystems [][]Equation
	for _, repeat := range withMaxRepeats {
		var newSystem []Equation
		for _, equation := range system {
			var isSolved = false
			for _, coefficient := range equation.Coefficients {
				if coefficient.Equal(repeat) {
					isSolved = true
					break
				}
			}
			if !isSolved {
				newSystem = append(newSystem, equation)
			}
		}
		newSystems = append(newSystems, newSystem)
	}
	return withMaxRepeats, newSystems
}

func GetMinimalVariant(system []Equation, result []K) []K {
	if len(system) == 0 {
		return result
	}
	for {
		mostRepeateds, newSystems := GetMostRepeated(system)
		if len(mostRepeateds) == 1 {
			mostRepeated, newSystem := mostRepeateds[0], newSystems[0]

			result = append(result, mostRepeated)
			system = newSystem

			combSet := make(map[string]struct{})
			for _, k := range result {
				combSet[k.String()] = struct{}{}
			}

			var isSystemSolved = true
			for _, eq := range system {
				var isEquationSolved = false
				for _, eqK := range eq.Coefficients {
					if _, found := combSet[eqK.String()]; found {
						isEquationSolved = true
						break
					}
				}
				if !isEquationSolved {
					isSystemSolved = false
					break
				}
			}
			if isSystemSolved {
				return result
			}
		} else {
			var possibleResults [][]K
			for i := range mostRepeateds {
				possibleResult := GetMinimalVariant(newSystems[i], append(result, mostRepeateds[i]))
				possibleResults = append(possibleResults, possibleResult)
			}

			minComplexity := math.MaxInt32
			var withMinComplexity []K
			for i := range possibleResults {
				complexity := KS(possibleResults[i]).Complexity()
				if complexity < minComplexity {
					minComplexity = complexity
					withMinComplexity = possibleResults[i]
				}
			}
			return withMinComplexity
		}
	}
}

func ExcludeOther(system []Equation, included []K) []Equation {
	includedSet := make(map[string]struct{}, len(included))
	for _, k := range included {
		includedSet[k.String()] = struct{}{}
	}

	var newSystem []Equation
	for _, equation := range system {
		newEquation := Equation{
			Value: equation.Value,
		}
		for _, k := range equation.Coefficients {
			if _, found := includedSet[k.String()]; found {
				newEquation.Coefficients = append(newEquation.Coefficients, k)
			}
		}
		newSystem = append(newSystem, newEquation)
	}
	return newSystem
}


func Format(ks []K) string {
	var result string
	for i, k := range ks {
		if i != 0 {
			result += " + "
		}
		result += k.PrettyString()
	}
	return result
}

func main() {
	f := []int{
		0, 0, 0, 1, 0, 0, 1, 0, 0, 1, // 00-09
		0, 1, 0, 1, 1, 1, 1, 0, 1, 1, // 10-19
		1, 0, 0, 0, 0, 0, 0, 1, 0, 0, // 20-29
		1, 1, 0, 1, 1, 1, 0, 0, 1, 0, // 30-39
		0, 1, 0, 1, 1, 1, 0, 1, 1, 0, // 40-49
		1, 1, 1, 0, 1, 1, 0, 1, 0, 0, // 50-59
		0, 1, 0, 1,                   // 60-63*/
	}
	system := MakeSystemOfEquations(f)
	system = ExcludeZeroCoefficients(system)
	fmt.Println("system with ")
	for _, eq := range system {
		fmt.Println(eq.KString())
	}
	result := GetMinimalVariant(system, nil)

	fmt.Println("result system:")
	system = ExcludeOther(system, result)
	for _, eq := range system {
		fmt.Println(eq.KString())
	}

	complexity := KS(result).Complexity()
	fmt.Printf("result size: %d\n", len(result))
	fmt.Printf("result complexity: %d\n", complexity)
	fmt.Printf("result: %s\n", Format(result))
}
