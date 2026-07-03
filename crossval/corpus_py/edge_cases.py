"""Edge cases: empty bodies, docstring-only, lambdas."""

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
