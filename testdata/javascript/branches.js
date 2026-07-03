// JavaScript branching constructs for CC testing

function allBranches(x, y, items) {
    let result = 0;
    // if/else
    if (x > 0) {
        result = 1;
    } else if (x === 0) {
        result = 0;
    } else {
        result = -1;
    }

    // while
    while (y > 0) {
        y--;
    }

    // for
    for (let item of items) {
        result += item;
    }

    // do/while
    do {
        y++;
    } while (y < 10);

    // try/catch
    try {
        result = 1 / y;
    } catch (e) {
        result = 0;
    }

    // switch/case
    switch (x) {
        case 1:
            result = 10;
            break;
        case 2:
            result = 20;
            break;
        default:
            result = 0;
    }

    // ternary
    const t = x > 0 ? 1 : -1;

    // logical operators in conditions
    if (x > 0 && y > 0) {
        result = 2;
    }

    if (x > 0 || y > 0) {
        result = 3;
    }

    return result;
}
