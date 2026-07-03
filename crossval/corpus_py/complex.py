"""High complexity function for testing upper CRAP ranges."""

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
