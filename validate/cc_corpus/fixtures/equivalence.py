def no_branches(x):
    return x + 1


def single_if(x):
    if x > 0:
        return 1
    return 0


def if_else_if(x):
    if x > 0:
        return 1
    elif x < 0:
        return -1
    else:
        return 0


def nested_if(x, y):
    if x > 0:
        if y > 0:
            return 1
    return 0


def for_loop(lst):
    for i in lst:
        pass
    return


def for_with_if(lst):
    for i in lst:
        if i > 0:
            return i
    return 0


def while_loop(n):
    while n > 0:
        n -= 1
    return n


def try_catch(x):
    try:
        return 1 // x
    except ZeroDivisionError:
        if x == 0:
            return -1
    return 0


def boolean_ops(a, b, c):
    if a and b or c:
        return 1
    return 0


def early_return(val):
    if val < 0:
        return -1
    if val == 0:
        return 0
    return 1


def ternary(x, a, b):
    if x:
        return a
    else:
        return b


def switch_case(x):
    if x == 1:
        return 'a'
    elif x == 2:
        return 'b'
    elif x == 3:
        return 'c'
    else:
        return 'z'
