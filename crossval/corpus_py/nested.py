"""Nested functions, decorators, class methods."""

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
