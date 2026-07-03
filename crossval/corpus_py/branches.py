"""All Python branching constructs."""

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
