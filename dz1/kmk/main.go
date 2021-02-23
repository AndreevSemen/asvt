package main

import (
	"fmt"
	"math"
	"strconv"
)

type Bit int

const (
	Tilde Bit = -1
	False Bit = 0
	True  Bit = 1
)

func (b Bit) String() string {
	switch b {
	case Tilde:
		return "~"
	case False:
		return "0"
	case True:
		return "1"
	default:
		panic(fmt.Sprintf("bad bit value: %d", b))
	}
}

func (b Bit) PrettyString(index int) string {
	switch b {
	case Tilde:
		return ""
	case False:
		return "!x" + strconv.Itoa(index)
	case True:
		return "x" + strconv.Itoa(index)
	default:
		panic(fmt.Sprintf("bad bit value: %d", b))
	}
}

type Term []Bit

func (a Term) String() string {
	var stringify string
	for _, bit := range a {
		stringify += bit.String()
	}
	return stringify
}

func (a Term) PrettyString() string {
	var prettyString string
	for i, bit := range a {
		prettyString = bit.PrettyString(i) + prettyString
	}
	return prettyString
}

func (a Term) Equals(b Term) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (a Term) Distance(b Term) int {
	if len(a) != len(b) {
		panic("expected terms lengths are equal")
	}
	dist := 0
	for i := range a {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}

func (a Term) Weight() int {
	count := 0
	for _, bit := range a {
		if bit == True {
			count++
		}
	}
	return count
}

func (a Term) DifferentBitIndex(b Term) int {
	if len(a) != len(b) {
		panic("expected terms lengths are equal")
	}
	for i := range a {
		if a[i] != b[i] {
			return i
		}
	}
	return -1
}

func (a Term) Covers(b Term) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != Tilde {
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

type GroupItem struct {
	Term
	IsGlued bool
}

func NewGroupItem(term Term) GroupItem {
	return GroupItem{
		Term:    term,
		IsGlued: false,
	}
}

type Groups map[int][]GroupItem

func GroupByWeight(terms []Term) Groups {
	if len(terms) == 0 {
		return nil
	}
	groups := make(map[int][]GroupItem, len(terms[0]))
	for _, term := range terms {
		groups[term.Weight()] = append(groups[term.Weight()], NewGroupItem(term))
	}
	return groups
}

func GlueGroups(a, b []GroupItem) (newTerms []Term) {
	for i := range a {
		for j := range b {
			if a[i].Distance(b[j].Term) == 1 {
				a[i].IsGlued = true
				b[j].IsGlued = true
				newTerm := make(Term, len(a[i].Term))
				copy(newTerm, a[i].Term)
				newTerm[a[i].DifferentBitIndex(b[j].Term)] = Tilde
				newTerms = append(newTerms, newTerm)
			}
		}
	}
	return newTerms
}

func MakeUniqueSet(terms []Term) []Term {
	uniqueSet := make([]Term, 0)
	for _, term := range terms {
		found := false
		for _, termInSet := range uniqueSet {
			if termInSet.Equals(term) {
				found = true
				break
			}
		}
		if !found {
			uniqueSet = append(uniqueSet, term)
		}
	}
	return uniqueSet
}

func Step1(impls []Term) []Term {
	groups := GroupByWeight(impls)
	glued := make([]Term, 0)
	for weight, groupB := range groups {
		if groupA, found := groups[weight-1]; found {
			newGlued := GlueGroups(groupA, groupB)
			glued = append(glued, newGlued...)
		}
	}
	if len(glued) == 0 {
		return impls
	}
	unaffectedTerms := make([]Term, 0)
	for _, group := range groups {
		for _, term := range group {
			if !term.IsGlued {
				unaffectedTerms = append(unaffectedTerms, term.Term)
			}
		}
	}
	return Step1(append(MakeUniqueSet(unaffectedTerms), MakeUniqueSet(glued)...))
}

type Line struct {
	Term Term
	IsMarked bool
}

type Table struct {
	Columns []Line
	Rows []Line
	Marks [][]bool
}

func NewTable(prime, source []Term) Table {
	t := Table{}
	for _, row := range prime {
		t.Rows = append(t.Rows, Line{
			Term:     row,
			IsMarked: false,
		})
	}
	for _, column := range source {
		t.Columns = append(t.Columns, Line{
			Term:     column,
			IsMarked: false,
		})
	}
	t.Marks = make([][]bool, len(t.Rows))
	for i := range t.Rows {
		t.Marks[i] = make([]bool, len(t.Columns))
	}
	return t
}

func (t Table) IsRowsExcess(rowsTakeOff []int) bool {
	Loop: for i := range t.Rows {
		for _, row := range rowsTakeOff {
			if i == row {
				continue Loop
			}
		}
		for j := range t.Columns {
			if !t.Rows[i].Term.Covers(t.Columns[j].Term) {
				return false
			}
		}
	}
	return true
}

func Steps2and3and4(prime, source []Term) Table {
	t := NewTable(prime, source)
	for i := range t.Rows {
		for j := range t.Columns {
			if t.Rows[i].Term.Covers(t.Columns[j].Term) {
				t.Marks[i][j] = true
			}
		}
	}
	for j := range t.Columns {
		marksInColumn := 0
		rowWithMark := 0
		for i := range t.Rows {
			if t.Marks[i][j] {
				marksInColumn++
				rowWithMark = i
			}
		}
		if marksInColumn == 1 {
			t.Columns[j].IsMarked = true
			t.Rows[rowWithMark].IsMarked = true
		}
	}
	newTable := Table{}
	for j := range t.Columns {
		newTable.Columns = append(newTable.Columns, t.Columns[j])
	}
	for i := range t.Rows {
		if t.Rows[i].IsMarked {
			newTable.Rows = append(newTable.Rows, t.Rows[i])
			newTable.Marks = append(newTable.Marks, t.Marks[i])
		}
	}
	return newTable
}

func Step5(t Table, excessRows []int) []Term {
	possibleResults := make([][]Term, 0)
	trivialResult := make([]Term, 0, len(t.Rows))
	for _, row := range t.Rows {
		trivialResult = append(trivialResult, row.Term)
	}
	possibleResults = append(possibleResults, trivialResult)

	Loop: for i := range t.Rows {
		for _, row := range excessRows {
			if i == row {
				continue Loop
			}
		}
		if t.IsRowsExcess(append(excessRows, i)) {
			possibleResults = append(possibleResults, Step5(t, append(excessRows, i)))
		}
	}

	var minResult []Term
	minResultSize := math.MaxInt64
	for _, result := range possibleResults {
		resultSize := 0
		for _, term := range result {
			for _, bit := range term {
				if bit != Tilde {
					resultSize++
				}
			}
		}
		if resultSize < minResultSize {
			minResult = result
			minResultSize = resultSize
		}
	}
	return minResult
}

func Format(impls []Term) string {
	var result string
	for i, impl := range impls {
		if i != 0 {
			result += " + "
		}
		result += impl.PrettyString()
	}
	return result
}

func main() {
	impls := []Term{
		/*{0, 1, 0, 0},
		{0, 0, 1, 1},
		{0, 1, 0, 1},
		{1, 0, 0, 1},
		{0, 1, 1, 1},
		{1, 1, 0, 1},
		{1, 1, 1, 0},
		{1, 1, 1, 1},*/
		{0, 0, 0, 0},
		{0, 0, 0, 1},
		{0, 0, 1, 0},
		{0, 1, 0, 0},
		{0, 1, 0, 1},
		{0, 1, 1, 0},
		{1, 0, 0, 1},
		{1, 1, 0, 0},
		{1, 1, 0, 1},
		{1, 1, 1, 0},

	}
	result := Step5(Steps2and3and4(Step1(impls), impls), nil)
	fmt.Println(Format(result))
}
