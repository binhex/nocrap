function switch_fallthrough(x: number): string {
    let result = "";
    switch (x) {
        case 1:
            result += "one";
            // fallthrough
        case 2:
            result += "two";
            break;
        case 3:
            result += "three";
            break;
        default:
            result += "other";
            break;
    }
    return result;
}

function optional_chaining(obj: { nested?: { value?: string } } | null): string | undefined {
    return obj?.nested?.value;
}

function nullish_coalescing(a: string | null, b: string | null): string {
    return a ?? b ?? "default";
}

function for_in_loop(obj: Record<string, number>): number {
    let sum = 0;
    for (const key in obj) {
        sum += obj[key];
    }
    return sum;
}

function for_of_loop(items: number[]): number {
    let sum = 0;
    for (const x of items) {
        sum += x;
    }
    return sum;
}
