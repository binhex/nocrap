function no_branches(x: number): number {
    return x + 1;
}

function single_if(x: number): number {
    if (x > 0) return 1;
    return 0;
}

function if_else_if(x: number): number {
    if (x > 0) return 1;
    else if (x < 0) return -1;
    else return 0;
}

function nested_if(x: number, y: number): number {
    if (x > 0) {
        if (y > 0) return 1;
    }
    return 0;
}

function for_loop(lst: number[]): void {
    for (let i = 0; i < lst.length; i++) {}
    return;
}

function for_with_if(lst: number[]): number {
    for (let i = 0; i < lst.length; i++) {
        if (lst[i] > 0) return lst[i];
    }
    return 0;
}

function while_loop(n: number): number {
    while (n > 0) { n--; }
    return n;
}

function try_catch(x: number): number {
    try {
        return Math.floor(1 / x);
    } catch(e) {
        if (x === 0) return -1;
        return 1;
    }
    return 0;
}

function boolean_ops(a: boolean, b: boolean, c: boolean): number {
    if (a && b || c) return 1;
    return 0;
}

function early_return(val: number): number {
    if (val < 0) return -1;
    if (val === 0) return 0;
    return 1;
}

function ternary(x: boolean, a: number, b: number): number {
    return x ? a : b;
}

function switch_case(x: number): string {
    switch(x) {
        case 1: return 'a';
        case 2: return 'b';
        default: return 'z';
    }
}
