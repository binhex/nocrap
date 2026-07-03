"""Test suite for CRAP validation corpus."""
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
