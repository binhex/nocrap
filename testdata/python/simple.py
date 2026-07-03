"""Simple module with basic functions for testing."""

def add(a, b):
    """Add two numbers."""
    return a + b

def multiply(a, b):
    return a * b

async def async_fetch():
    return 42

class Calculator:
    """A simple calculator class."""

    def __init__(self, initial=0):
        self.value = initial

    def add(self, x):
        self.value += x
        return self.value

    def get_value(self):
        return self.value
