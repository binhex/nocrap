"""Module with nested functions, decorators, and edge cases."""

def outer():
    """Outer function."""
    def inner():
        return 1
    return inner

def decorated_func():
    """A decorated function."""
    return True

def docstring_only():
    """Only a docstring, no body."""

def empty_body():
    pass
