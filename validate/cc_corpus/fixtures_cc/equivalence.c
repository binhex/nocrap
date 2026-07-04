int no_branches() {
    return 42;
}

int single_if(int x) {
    if (x > 0) {
        return 1;
    } else {
        return 0;
    }
}

int if_else_if(int x) {
    if (x > 0) {
        return 1;
    } else if (x < 0) {
        return -1;
    } else {
        return 0;
    }
}

int nested_if(int a, int b) {
    if (a > 0) {
        if (b > 0) {
            return 1;
        }
    }
    return 0;
}

int for_loop(int n) {
    int s = 0;
    for (int i = 0; i < n; i++) {
        s += i;
    }
    return s;
}

int for_with_if(int* items, int len) {
    int count = 0;
    for (int i = 0; i < len; i++) {
        if (items[i] > 0) {
            count++;
        }
    }
    return count;
}

int while_loop(int x) {
    int n = 0;
    while (x > 0) {
        x--;
        n++;
    }
    return n;
}

int try_catch() {
    // C has no try/catch — stub, skipped by skip_c in expected.json
    return 0;
}

int boolean_ops(int a, int b, int c) {
    if (a && b || c) {
        return 1;
    }
    return 0;
}

int early_return(int x) {
    if (x > 0) {
        return 1;
    }
    if (x < 0) {
        return -1;
    }
    return 0;
}

int ternary(int x) {
    return x > 0 ? 1 : 0;
}

int switch_case(int x) {
    switch (x) {
        case 1: return 1;
        case 2: return 2;
        default: return 0;
    }
}
