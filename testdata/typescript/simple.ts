// Simple TypeScript functions

function add(a: number, b: number): number {
    return a + b;
}

const multiply = (a: number, b: number): number => a * b;

async function fetchData<T>(): Promise<T> {
    return {} as T;
}

class Calculator {
    constructor(private initial: number = 0) {}

    add(x: number): number {
        this.initial += x;
        return this.initial;
    }
}
