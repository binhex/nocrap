function no_branches(x) {
    return x + 1;
}

function single_if(x) {
    if (x > 0) return 1;
    return 0;
}

function if_else_if(x) {
    if (x > 0) return 1;
    else if (x < 0) return -1;
    else return 0;
}

function nested_if(x, y) {
    if (x > 0) {
        if (y > 0) return 1;
    }
    return 0;
}

function for_loop(lst) {
    for (let i = 0; i < lst.length; i++) {}
    return;
}

function for_with_if(lst) {
    for (let i = 0; i < lst.length; i++) {
        if (lst[i] > 0) return lst[i];
    }
    return 0;
}

function while_loop(n) {
    while (n > 0) { n--; }
    return n;
}

function try_catch(x) {
    try {
        return Math.floor(1 / x);
    } catch(e) {
        if (x === 0) return -1;
        return 1;
    }
    return 0;
}

function boolean_ops(a, b, c) {
    if (a && b || c) return 1;
    return 0;
}

function early_return(val) {
    if (val < 0) return -1;
    if (val === 0) return 0;
    return 1;
}

function ternary(x, a, b) {
    return x ? a : b;
}

function switch_case(x) {
    switch(x) {
        case 1: return 'a';
        case 2: return 'b';
        default: return 'z';
    }
}
