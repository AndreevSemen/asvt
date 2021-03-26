package main

import (
	"fmt"
	"math"
	"strconv"
)

// Представляет одну переменную в импликанте
type Bit int

// Переменная может:
const (
	Tilde Bit = -1 // Отсутствовать в импиканте
	False Bit = 0 // Быть инвертированной
	True  Bit = 1 // Быть прямой
)

// Функция преобразования переменной в удобочитаемый вид
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

// Представляет любую импликанту (терм) в виде набора переменных
type Term []Bit

// Функция преобразования импликанты в удобочитаемый вид
// Переводим каждую переменную в строку и конкатенируем их
// Стоит отметить, что переменные импликанты печатаются в обратном порядке:
// 01~~ -> "x1!x0"
func (a Term) PrettyString() string {
	var prettyString string
	for i, bit := range a {
		prettyString = bit.PrettyString(i) + prettyString
	}
	return prettyString
}

// Функция сравнения импликант на равенство
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

// Функция нахождения расстояния между двумя импликантами
func (a Term) Distance(b Term) int {
	// Если импликанты зависят от разного количества
	// переменных, то считаем, что они не сравнимы
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

// Функция расчета веса импликанты
func (a Term) Weight() int {
	count := 0
	for _, bit := range a {
		if bit == True {
			count++
		}
	}
	return count
}


// Функция которая возвращает номер первой переменной,
// в которой импликанты различаются
func (a Term) DifferentBitIndex(b Term) int {
	// Если импликанты зависят от разного количества
	// переменных, то считаем, что они не сравнимы
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

// Функция проверки вхождения одной импликанты в другую
// К примеру импликанты 10~0 и 1~10 входят в импликанту 1~~0
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

// На первом шаге алгоритма мы разбиваем импликанты на группы по весу
// Данная структура представляет собой импликанту с отметкой о том,
// была ли она задействована при образовании новой склеенной импликанты
type GroupItem struct {
	Term
	IsGlued bool
}

// Вспомогательная функция для создания элемента по умолчанию
func NewGroupItem(term Term) GroupItem {
	return GroupItem{
		Term:    term,
		IsGlued: false,
	}
}

// Представляет собой отображение веса на группу элементов с данным весом
type Groups map[int][]GroupItem

// Функция, которая создает группы, объединенные по весам
func GroupByWeight(terms []Term) Groups {
	if len(terms) == 0 {
		return nil
	}
	groups := make(map[int][]GroupItem, len(terms[0]))
	for _, term := range terms {
		// Считаем вес импликанты и добавляем ее в соответствующую группу
		weight := term.Weight()
		groups[weight] = append(groups[weight], NewGroupItem(term))
	}
	return groups
}

// Функцию реализует процесс склейки соседних по весу групп
// Результатом функции является набор импликант, образовавшихся при склеивании
func GlueGroups(a, b []GroupItem) (newTerms []Term) {
	// Перебираем все возможные пары элементов двух групп
	for i := range a {
		for j := range b {
			// Если он отличаются в одной позиции, то производим склейку
			if a[i].Distance(b[j].Term) == 1 {
				// Импликанты, которые участвовали в склеивании помечаются
				// поднятым флагом IsGlued для того, чтобы далее их можно было исключить
				a[i].IsGlued = true
				b[j].IsGlued = true
				// Создаем новую импликанту
				newTerm := make(Term, len(a[i].Term))
				copy(newTerm, a[i].Term)
				newTerm[a[i].DifferentBitIndex(b[j].Term)] = Tilde
				newTerms = append(newTerms, newTerm)
			}
		}
	}
	return newTerms
}

// Функция создает новый набор на основе входного, но без повторяющихся элементов
// К пр.: 10~1, 10~1, 010~, ~1~~ -> 10~1, 010~, ~1~~
func MakeUniqueSet(terms []Term) []Term {
	// Множество уникальных элементов
	uniqueSet := make([]Term, 0)
	for _, term := range terms {
		// Ищем элемент term во множестве уже найденных уникальных элементов
		found := false
		for _, termInSet := range uniqueSet {
			if termInSet.Equals(term) {
				found = true
				break
			}
		}
		// Если элемент еще не присутствует в уникальных, то добавляем его
		if !found {
			uniqueSet = append(uniqueSet, term)
		}
	}
	return uniqueSet
}

// Функцию реализует первый шаг алгоритма - склейка импликант
// Функция возвращает набор импликант, которые больше невозможно склеить
func Step1(impls []Term) []Term {
	// Формируем весовые группы
	groups := GroupByWeight(impls)
	// Склеиваем каждую весовую группу с предыдущей по весу, если такая имеется
	// Склеенные импликанты сохраняем
	glued := make([]Term, 0)
	for weight, groupB := range groups {
		if groupA, found := groups[weight-1]; found {
			newGlued := GlueGroups(groupA, groupB)
			glued = append(glued, newGlued...)
		}
	}
	// Если не произошло ни одного склеивания, то возвращаем входной набор импликант
	if len(glued) == 0 {
		return impls
	}
	// Ищем те импликанты, которые не были склеены
	unaffectedTerms := make([]Term, 0)
	for _, group := range groups {
		for _, term := range group {
			if !term.IsGlued {
				unaffectedTerms = append(unaffectedTerms, term.Term)
			}
		}
	}
	// К новым полученным импликантам добавляем те, что не были склеены
	impls = MakeUniqueSet(append(unaffectedTerms, glued...))
	// Запускаем следующий шаг рекурсии
	return Step1(impls)
}

// Линия представляет собой описание строки либо столбца таблицы для шагов 2, 3 и 4
// Каждой линии соответствует импликанта и служебная метка
type Line struct {
	Term Term
	IsMarked bool
}

// Таблица для шагов 2, 3 и 4
// Содержит описания столбцов и строк, а так же таблицу меток о том, что
// импликанта строки покрывает импликанту столбца
type Table struct {
	Columns []Line
	Rows []Line
	Marks [][]bool
}

// Создает новую таблицу из наборов простых и исходных импликант
func NewTable(prime, source []Term) Table {
	t := Table{}
	// Каждой строке ставим соответствие простую импликанту
	for _, row := range prime {
		t.Rows = append(t.Rows, Line{
			Term:     row,
			IsMarked: false,
		})
	}
	// Каждому столбцу ставим в соответствие исходную импликанту
	for _, column := range source {
		t.Columns = append(t.Columns, Line{
			Term:     column,
			IsMarked: false,
		})
	}
	// Заполняем таблицу меток
	t.Marks = make([][]bool, len(t.Rows))
	for i := range t.Rows {
		t.Marks[i] = make([]bool, len(t.Columns))
		for j := range t.Columns {
			// Если строка покрывает столбец, то ставим отметку
			if t.Rows[i].Term.Covers(t.Columns[j].Term) {
				t.Marks[i][j] = true
			}
		}
	}
	return t
}

// Данная функция проверяет является ли набор строк под номерами
// из rowsTakeOff избыточными
func (t Table) IsRowsExcess(rowsTakeOff map[int]struct{}) bool {
	// Создаем массив меток о том, что столбцы были покрыты
	coveredColumns := make([]bool, len(t.Columns))
	// Проверяем каждую строку
	for i := range t.Rows {
		// Если номер строки содержится в rowsTakeOff, то переходим к следующей строке
		if _, found := rowsTakeOff[i]; found {
			continue
		}
		// Ставим метки о том, какие столбцы были покрыты
		for j := range t.Columns {
			if t.Marks[i][j] == true {
				coveredColumns[j] = true
			}
		}
	}
	// Если хоть один из столбцов не был покрыт, то набор оставшихся
	// строк не покрывает функцию полностью
	for _, covered := range coveredColumns {
		if !covered {
			return false
		}
	}
	return true
}

// Реализует 2, 3 и 4 шаги алгоритма
// Возвращает таблицу и набор существенных строк
func Steps2and3and4(prime, source []Term) (Table, map[int]struct{}) {
	// Создаем таблицу
	t := NewTable(prime, source)
	// Ищем существенные строки и столбцы
	for j := range t.Columns {
		marksInColumn := 0
		rowWithMark := 0
		// Считаем количество отметок в столбце
		for i := range t.Rows {
			if t.Marks[i][j] {
				marksInColumn++
				rowWithMark = i
			}
		}
		// Если в столбце только одна отметка, то строка и столбец,
		// соответствующие этой отметке существенны
		if marksInColumn == 1 {
			t.Columns[j].IsMarked = true
			t.Rows[rowWithMark].IsMarked = true
		}
	}
	// Составляем набор существенных строк
	essentials := make(map[int]struct{}, 0)
	for i := range t.Rows {
		if t.Rows[i].IsMarked {
			essentials[i] = struct{}{}
		}
	}
	return t, essentials
}

// Функция реализует 5 шаг алгоритма
func Step5(t Table, essentialRows, excessRows map[int]struct{}) []Term {
	// possibleResults - все доступные наборы импликант, покрывающие ФАЛ
	possibleResults := make([][]Term, 0)
	// Добавляем тривиальный случай - все строки в таблице
	trivialResult := make([]Term, 0, len(t.Rows))
	for i, row := range t.Rows {
		if _, found := excessRows[i]; found {
			continue
		}
		trivialResult = append(trivialResult, row.Term)
	}
	possibleResults = append(possibleResults, trivialResult)

	// Проходим по всем строкам. Исключаем те, которые являются избыточными и
	// рекурсивно вызываем шаг 5 алгоритма для нового набора строк
	// Если строка существенная (essential), то не исключаем ее в любом случае
	for i := range t.Rows {
		// Проверяем не является ли строка существенной
		if _, found := essentialRows[i]; found {
			continue
		}
		// Проверяем не была ли ранее исключена данная строка
		if _, found := excessRows[i]; found {
			continue
		}
		// Предполагаем, что строка i избыточная
		excessRows[i] = struct{}{}
		// Если строка под номером i является избыточной, то отбрасываем ее и
		// к доступному набору импликант, покрывающих ФАЛ, прибавляем результат
		// рекурсивного вызова шага 5 без строки i
		if t.IsRowsExcess(excessRows) {
			possibleResults = append(possibleResults, Step5(t, essentialRows, excessRows))
		}
		delete(excessRows, i)
	}

	// Ищем из доступных наборов импликант, покрывающих ФАЛ, тот, что является минимальным
	var minResult []Term
	minResultSize := math.MaxInt64
	for _, result := range possibleResults {
		// Считаем количество литералов в наборе
		resultSize := 0
		for _, term := range result {
			for _, bit := range term {
				if bit != Tilde {
					resultSize++
				}
			}
		}
		// Если количество литералов меньше, того, что было найдено ранее,
		// то запоминаем этот набор
		if resultSize < minResultSize {
			minResult = result
			minResultSize = resultSize
		}
	}
	return minResult
}

// Функция возвращает СДНФ от ФАЛ
func MakeSDNF(f []int) []Term {
	variableNumber := int(math.Log2(float64(len(f))))
	sdnf := make([]Term, 0, len(f))
	format := "%0" + strconv.Itoa(variableNumber) + "b"
	for i := range f {
		if f[i] == 1 {
			term := make(Term, 0, variableNumber)
			binaryString := fmt.Sprintf(format, i)
			for _, char := range binaryString {
				switch char {
				case '0':
					term = append(term, False)
				case '1':
					term = append(term, True)
				default:
					panic(fmt.Sprintf("unexpected char: %s", string(char)))
				}
			}
			sdnf = append(sdnf, term)
		}
	}
	return sdnf
}

// Функция форматирует импликанты в строку
// К пр.: [10~1, 010~, ~1~~] -> "x3!x1x0 + !x2x1!x0 + x3!x1x0"
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
	f := []int{
		0, 0, 0, 1, 0, 0, 1, 0, 0, 1, // 00-09
		0, 1, 0, 1, 1, 1, 1, 0, 1, 1, // 10-19
		1, 0, 0, 0, 0, 0, 0, 1, 0, 0, // 20-29
		1, 1, 0, 1, 1, 1, 0, 0, 1, 0, // 30-39
		0, 1, 0, 1, 1, 1, 0, 1, 1, 0, // 40-49
		1, 1, 1, 0, 1, 1, 0, 1, 0, 0, // 50-59
		0, 1, 0, 1,                   // 60-63
	}
	impls := MakeSDNF(f)
	primeImpls := Step1(impls)
	table, essentialRows := Steps2and3and4(primeImpls, impls)
	excessRows := make(map[int]struct{})
	result := Step5(table, essentialRows, excessRows)
	// Сделаем проверку на то, что все исходные импликанты покрыты
	covered := 0
	total := 0
	for _, sourceImpl := range impls {
		isCover := false
		for _, minImpl := range result {
			if minImpl.Covers(sourceImpl) {
				isCover = true
			}
		}
		if !isCover {
			fmt.Print("E")
		} else {
			covered++
			fmt.Print(".")
		}
		total++
	}
	fmt.Println()
	fmt.Printf("total coverage: %3.4f%%\n", float32(covered)/float32(total)*100)
	fmt.Printf("result: %s\n", Format(result))
}
