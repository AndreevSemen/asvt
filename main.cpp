#include <iostream>
#include <memory>
#include <utility>
#include <list>
#include <fstream>
#include <vector>

// От данного класса наследуются все классы используемые
// для создания формата файла .tex
class Object {
public:
    virtual std::string Dump() const = 0;
};

// Класс, представляющий текстовую строку
class String : public Object {
private:
    std::string _contents;

public:
    explicit String(std::string contents)
      : _contents(std::move(contents))
    {}

    std::string Dump() const override {
        return _contents;
    }
};

// Класс, представляющий строку с верхним подчеркиванием
class Overline : public String {
public:
    explicit Overline(std::string contents)
      : String(std::move(contents))
    {}

    std::string Dump() const override {
        return "\\overline{"+String::Dump()+"}";
    }
};

// Класс, представляющий строку c нижнем индексом
class WithIndex : public String {
private:
    std::string _index;
public:
    explicit WithIndex(std::string value, std::string index)
      : String(std::move(value))
      , _index(std::move(index))
    {}

    std::string Dump() const override {
        return "x_{" + _index + "}";
    }
};

// Класс, представляющий заголовок внутри файла .tex
class Header : public Object {
private:
    std::string _name;
    std::string _option;
    std::string _value;

public:
    Header(std::string name, std::string value)
      : _name(std::move(name))
      , _value(std::move(value))
    {}

    // Можно задать опцию жля заголовка
    Header(std::string name, std::string option, std::string value)
      : _name(std::move(name))
      , _option(std::move(option))
      , _value(std::move(value))
    {}

    std::string Dump() const override {
        if (_option.empty()) {
            return "\\" + _name + "{" + _value + "}";
        } else {
            return "\\" + _name + "[" + _option + "]" + "{" + _value + "}";
        }
    }
};

// Класс, представляющий теговые структуры в формате .tex
// К пр.: begin{article} ....some text.... end{article}
class Frame : public Object {
private:
    std::string _begin;
    std::string _end;
    std::list<std::shared_ptr<Object>> _contents;

public:
    Frame(std::string begin, std::string end)
      : _begin(std::move(begin))
      , _end(std::move(end))
    {}

    // В теговую структуру можно добавить несколько подструктур
    virtual void Append(std::shared_ptr<Object> item) {
        _contents.push_back(std::move(item));
    }

    bool Empty() const {
        return _contents.empty();
    }

    std::string Dump() const override {
        std::string listDump;
        for (const auto& item : _contents) {
            listDump += item->Dump() + "\n";
        }
        return "\n" + _begin + listDump + _end;
    }
};

// Специализация теговой структуры Frame для тегов begin{document}/end{document}
class Document : public Frame {
public:
    Document()
      : Frame("\\begin{document}", "\\end{document}")
    {}
};

// Специализация теговой структуры Frame для математических формул
class Math : public Frame {
public:
    Math()
      : Frame("\\begin{equation*}", "\\end{equation*}")
    {}
};

// Специализация теговой структуры Frame для тегов массива begin{array}{cc}/end{array}
// Опция {cc} означает оцентровку формулы
class Array : public Frame {
public:
    Array()
      : Frame("\\begin{array}{cc}", "\\end{array}")
    {}

    void Append(std::shared_ptr<Object> item) override {
        if (!Empty()) {
            Frame::Append(std::make_shared<String>("\\\\"));
        }
        Frame::Append(std::make_shared<String>(item->Dump()));
    }
};

class File : public Frame {
public:
    File()
      : Frame("", "")
    {}
};


// Функция переводящая строку таблицы истинности (импликанту)
// в математическую формулу в формате .tex
// К пр.: Из следующий строки таблицы истинности
// x3 x2 x1 x0 f
// 1  0  1  1  1
// получим отформатированную импликанту:
// x_{3}\overline{x_2}x_{1}x_{0}
template <size_t LineSize>
std::shared_ptr<Object> GetImplicant(const std::array<bool, LineSize>& line) {
    std::string implicant; // В результате отформатированная импликанта окажется в данной строке
    // Итерируемся по всем значениям xi (последним элементом массива будет значение функции)
    for (size_t i = 0; i < LineSize - 1; ++i) {
        auto xi = WithIndex("x", std::to_string(i + 1)).Dump(); // Приводим в вид: x_{i}
        if (line[i] == false) { // Если xi == 0, то необходимо поставить инверсию в импликанте над xi
            xi = Overline(xi).Dump();
        }
        implicant += xi; // Конкатинируем все xi
    }
    return std::make_shared<String>(implicant);
}

// Функция проходит по строкам таблицы истинности и форматирует те строки, на которых функция f истина
template <size_t LinesNumber, size_t LineSize>
std::vector<std::shared_ptr<Object>> GetImplicants(const std::array<std::array<bool, LineSize>, LinesNumber>& table) {
    std::vector<std::shared_ptr<Object>> implicants;
    for (const auto& line : table) {
        if (line.back() == 1) {
            auto implicant = GetImplicant(line);
            implicants.push_back(implicant);
        }
    }
    return implicants;
}

// Функцию формирует из вектора значений функции f таблицу истинности функции f
template <size_t LinesNumber, size_t LineSize>
std::array<std::array<bool, LineSize>, LinesNumber> MakeTable(const std::array<bool, 64>& f) {
    std::array<std::array<bool, LineSize>, LinesNumber> table{};
    for (size_t i = 0; i < LinesNumber; ++i) {
        // Переводим i в двоичную систему и записываем в строке таблицы
        for (size_t j = 0; j < LineSize - 1; ++j) {
            table[i][j] = i & (0b00000001 << (LineSize - 2 - j));
        }
        table[i].back() = f[i];
    }

    return table;
}

// Функцию группирует входные импликанты по строкам, чтобы вывод был красивым
std::shared_ptr<Object> MakeLaTeXFormula(const std::vector<std::shared_ptr<Object>>& implicants) {
    // Создаем массив для формата .tex
    auto array = std::make_shared<Array>();
    array->Append(std::make_shared<String>("f(x_0, x_1, x_2, x_3, x_4, x_5) ="));

    size_t subSize = 0;
    std::string row;
    // Формируем строки, в которых не более 4 импликант
    for (size_t i = 0; i < implicants.size(); ++i) {
        if (row.empty() && i != 0) { // Ставим плюс, перенесенный с предыдущей строки
            row += "+";
        }
        row += implicants[i]->Dump(); // Выводим отформатированную импликанту
        if (i != implicants.size() - 1) { // Ставим после импликанты плюс, если она не последняя
            row += "+";
        }
        subSize++;
        if (subSize > 4) { // Заносим в массив отформатированную строку
            array->Append(std::make_shared<String>(row));
            row = "";
            subSize = 0;
        }
    }
    if (subSize != 0) { // Заносим в массив отаточную строку
        array->Append(std::make_shared<String>(row));
    }

    return array;
}

int main() {
    // Здесь я формировал вектор значений функции из карты Карно
    // Делал несколько преобразований для удобства
    /* Порядковые номера значений в карте Карно
     * 36|37|39|38||34|35|33|32
     * 44|45|47|46||42|43|41|40
     * 60|61|63|62||58|59|57|56
     * 52|53|55|54||50|51|49|48
     * ========================
     * 20|21|23|22||18|19|17|16
     * 28|29|31|30||26|27|25|24
     * 12|13|15|14||10|11| 9| 8
     *  4| 5| 7| 6|| 2| 3| 1| 0
     * */
    /* Моя карта Карно
     * | | | |1||1|1|1| |
     * |1|1|1| || |1|1| |
     * | |1|1| || | |1| |
     * |1| |1|1||1|1| |1|
     * ==================
     * |1| | | ||1|1| |1|
     * | | |1|1|| |1| | |
     * | |1|1|1|| |1|1| |
     * | | | |1|| |1| | |
     * */
    /* Оставил только те порядковые номера, где значение функции истино
     *   |  |  |38||34|35|33|
     * 44|45|47|  ||  |43|41|
     *   |61|63|  ||  |  |57|
     * 52|  |55|54||50|51|  |48
     * ========================
     * 20|  |  |  ||18|19|  |16
     *   |  |31|30||  |27|  |
     *   |13|15|14||  |11| 9|
     *   |  |  | 6||  | 3|  |
     * */
    // Составил вектор значений
    std::array<bool, 64> f{
        0, 0, 0, 1, 0, 0, 1, 0, 0, 1, // 00-09
        0, 1, 0, 1, 1, 1, 1, 0, 1, 1, // 10-19
        1, 0, 0, 0, 0, 0, 0, 1, 0, 0, // 20-29
        1, 1, 0, 1, 1, 1, 0, 0, 1, 0, // 30-39
        0, 1, 0, 1, 1, 1, 0, 1, 1, 0, // 40-49
        1, 1, 1, 0, 1, 1, 0, 1, 0, 0, // 50-59
        0, 1, 0, 1                   // 60-63
    };

    auto implicants = GetImplicants(MakeTable<64, 7>(f));
    auto array = MakeLaTeXFormula(implicants);

    auto math = std::make_shared<Math>();
    math->Append(array);


    auto doc = std::make_shared<Document>();
    doc->Append(math);

    auto texFile = std::make_shared<File>();
    texFile->Append(std::make_shared<Header>("documentclass", "article"));
    texFile->Append(std::make_shared<Header>("usepackage", "amsmath"));
    texFile->Append(std::make_shared<Header>("usepackage", "utf8", "inputenc"));
    texFile->Append(doc);

    std::ofstream file("/home/insomniac/Git/asvt/latex-kmk/text.tex");
    if (!file.is_open()) {
        throw std::invalid_argument{"bad file path"};
    }
    file << texFile->Dump();

    return 0;
}
