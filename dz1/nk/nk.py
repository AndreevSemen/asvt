from itertools import combinations as c


def get_all_equations(n, values):
    with open('all_equations.txt', 'w', encoding='utf8') as file:
        result = []
        for i in range(2 ** n):
            b_num = bin(i ^ 2 ** n)[:2:-1]
            row = [int(values[i]), b_num[::-1]]
            for j in range(1, n + 1):
                for coefficient in c(enumerate(b_num, 1), j):
                    low = ''.join([str(coefficient[i][0]) for i in reversed(range(len(coefficient)))])
                    up = ''.join([str(coefficient[i][1]) for i in reversed(range(len(coefficient)))])
                    row.append('K_{}^{}'.format(low, up))
            result.append(row)
            file.write(' ?'.join(row[n + 1:1:-1])
                       + ' ?'
                       + ' ?'.join(row[n + 2::])
                       + ' = f({})'.format(row[1])
                       + ' = {}'.format(row[0]) + '\n')
    return result


def find_zeros_coefficients(coefficients):
    zeros = set()
    for row in coefficients:
        if not row[0]:
            for i in range(2, len(row)):
                zeros.add(row[i])
    return zeros


def delete_zeros_coefficients(coefficients, zeros):
    for row in coefficients:
        for i in range(2, len(row)):
            if row[i] in zeros:
                row[i] = None
    coefficients = [[row[i] for i in range(len(row)) if row[i]] for row in coefficients if row[0]]
    return coefficients


def get_non_zeros_coefficients(n, coefficients, values):
    with open('without_zeros.txt', 'w', encoding='utf8') as file:
        zeros = find_zeros_coefficients(coefficients)
        non_zeros = delete_zeros_coefficients(coefficients, zeros)
        for i in range(len(non_zeros)):
            file.write(' ?'.join(non_zeros[i][2:])
                       + ' = f({})'.format(non_zeros[i][1])
                       + ' = {}'.format(non_zeros[i][0]) + '\n')
    return non_zeros


def number_of_duplicates(minimal_coefficients):
    repetitions = {}
    all_coefficients = []
    for row in minimal_coefficients:
        all_coefficients.extend(row)
    for coefficient in all_coefficients:
        if coefficient not in repetitions:
            repetitions[coefficient] = all_coefficients.count(coefficient)
        else:
            continue
    return repetitions


def delete_duplicates(minimal_coefficients, duplicates):
    without_duplicates = set()
    for row in minimal_coefficients:
        if not any(filter(lambda x: x in without_duplicates, row)):
            minimal = min(row)
            for i in range(len(row)):
                if duplicates[row[i]] > duplicates[minimal]:
                    minimal = row[i]
            without_duplicates.add(minimal)
    return without_duplicates


def get_minimal_coefficients(non_zeros):
    minimal_coefficients = []
    with open('equations_with_minimal_coefficients.txt', 'w', encoding='utf8') as file:
        for row in non_zeros:
            minimal_length = min(row[2:], key=len)
            minimal_coefficients.append(list(filter(lambda x: len(x) == len(minimal_length), row[2:])))
        duplicates = number_of_duplicates(minimal_coefficients)
        minimal_coefficients = delete_duplicates(sorted(minimal_coefficients, key=len), duplicates)
        for row in non_zeros:
            for i in range(2, len(row)):
                if row[i] not in minimal_coefficients:
                    row[i] = None
        non_zeros = [[row[i] for i in range(len(row)) if row[i]] for row in non_zeros]
        for i in range(len(non_zeros)):
            file.write(' ?'.join(non_zeros[i][2:])
                       + ' = f({})'.format(non_zeros[i][1])
                       + ' = {}'.format(non_zeros[i][0]) + '\n')
    return minimal_coefficients

def printSLLE(slle):
    for i in range(len(slle)):
        for j in range(2, len(slle[i])):
            print(slle[i][j], end=" v ")
        print("=", slle[i][0])


if __name__ == '__main__':
    print('Enter number of variables: 6')
    num_of_variables = 6
    print('\nEnter {} function values:'.format(2 ** num_of_variables))
    function_str = '0001001001010111101110000001001101110010010111011011101101000101'
    function_values = ''
    for vector in range(2 ** num_of_variables):
        print('f({}) = '.format(bin(vector ^ 2 ** num_of_variables)[3:]) + function_str[vector])
        function_values += ""
    all_equations = get_all_equations(num_of_variables, function_str)
    printSLLE(all_equations)
    print('\nNow let\'s delete nullable coefficients:\n')
    non_zeros_coefficients = get_non_zeros_coefficients(num_of_variables, all_equations, function_values)
    printSLLE(non_zeros_coefficients)
    print("\nAnd let's find resulting coefficients.")
    result = get_minimal_coefficients(non_zeros_coefficients)
    print('\nMinimal coefficients:')
    for coefficient in result:
        print(coefficient + ' = 1')

