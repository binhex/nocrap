// TypeScript branching constructs for CC testing

function allBranches(x: number, y: number, items: number[]): number {
    let result = 0;
    if (x > 0) {
        result = 1;
    } else if (x === 0) {
        result = 0;
    } else {
        result = -1;
    }

    while (y > 0) {
        y--;
    }

    for (const item of items) {
        result += item;
    }

    do {
        y++;
    } while (y < 10);

    try {
        result = 1 / y;
    } catch (e) {
        result = 0;
    }

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

    const t = x > 0 ? 1 : -1;

    if (x > 0 && y > 0) {
        result = 2;
    }

    if (x > 0 || y > 0) {
        result = 3;
    }

    return result;
}
