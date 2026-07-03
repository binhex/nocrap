#!/usr/bin/env python3
"""Generate a comprehensive Python test corpus covering all branching constructs."""
import os

CORPUS_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "corpus_py")


def write_file(name, content):
    path = os.path.join(CORPUS_DIR, name)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        f.write(content)


def generate():
    os.makedirs(CORPUS_DIR, exist_ok=True)

    # File 1: Simple functions
    write_file("simple.py", '''"""Simple functions for baseline CRAP testing."""

def add(a, b):
    return a + b

def identity(x):
    return x

def always_true():
    return True
''')

    # File 2: All branching constructs
    write_file("branches.py", '''"""All Python branching constructs."""

def all_branches(x, y, items):
    if x > 0:
        result = 1
    elif x == 0:
        result = 0
    else:
        result = -1

    while y > 0:
        y -= 1

    for item in items:
        result += item

    try:
        result = 1 / y
    except ZeroDivisionError:
        result = 0
    except (ValueError, TypeError):
        result = -1

    with open("/dev/null") as f:
        f.read()

    if x > 0 and y > 0:
        result = 2

    if x > 0 or y > 0:
        result = 3

    return result
''')

    # File 3: Nested, decorators, methods
    write_file("nested.py", '''"""Nested functions, decorators, class methods."""

def outer():
    def inner():
        return 1
    return inner

def with_decorator():
    """Has a decorator."""
    return True

class Calculator:
    def __init__(self, initial=0):
        self.value = initial

    def add(self, x):
        self.value += x
        return self.value

    @property
    def value_squared(self):
        return self.value ** 2

    @staticmethod
    def static_help():
        return "I can add numbers"
''')

    # File 4: Edge cases
    write_file("edge_cases.py", '''"""Edge cases: empty bodies, docstring-only, lambdas."""

def empty_pass():
    pass

def docstring_only():
    """Only a docstring here."""

def single_line(): return 42

def with_lambda():
    f = lambda x: x + 1
    return f(5)

async def async_func():
    return 42

def match_case(value):
    match value:
        case 1:
            return "one"
        case 2:
            return "two"
        case _:
            return "other"
''')

    # File 5: High complexity
    write_file("complex.py", '''"""High complexity function for testing upper CRAP ranges."""

def very_complex(a, b, c, d, e):
    if a:
        if b:
            if c:
                return 1
            elif d:
                return 2
            else:
                return 3
        elif e:
            return 4
        else:
            return 5
    else:
        if b or c:
            if d and e:
                return 6
            return 7
        return 8
''')

    # File 6: Test file that exercises all functions to generate coverage data
    write_file("test_corpus.py", '''"""Test suite for CRAP validation corpus."""
from simple import add, identity, always_true
from branches import all_branches
from nested import outer, with_decorator, Calculator
from edge_cases import empty_pass, docstring_only, single_line, with_lambda, match_case
from complex import very_complex


def test_basics():
    assert add(1, 2) == 3
    assert identity(42) == 42
    assert always_true()


def test_branches():
    result = all_branches(0, 1, [1, 2])
    assert result is not None


def test_nested():
    fn = outer()
    assert fn() == 1
    assert with_decorator()


def test_calculator():
    calc = Calculator(10)
    calc.add(5)
    assert calc.value_squared == 225
    assert Calculator.static_help() == "I can add numbers"


def test_edge_cases():
    empty_pass()
    docstring_only()
    assert single_line() == 42
    assert with_lambda() == 6
    assert match_case(1) == "one"
    assert match_case(2) == "two"
    assert match_case("other") == "other"


def test_complex():
    assert very_complex(True, True, True, True, True) == 1
    assert very_complex(True, True, False, True, True) == 2
    assert very_complex(True, True, False, False, True) == 3
    assert very_complex(True, True, False, False, False) == 3
    assert very_complex(True, False, False, False, True) == 4
    assert very_complex(True, False, False, False, False) == 5
    assert very_complex(False, True, True, True, True) == 6
    assert very_complex(False, True, False, False, False) == 7
    assert very_complex(False, False, False, False, False) == 8
''')

    print(f"Generated corpus in {CORPUS_DIR}")


if __name__ == "__main__":
    generate()
