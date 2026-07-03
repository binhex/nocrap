// Simple JavaScript functions for testing

function add(a, b) {
    return a + b;
}

const multiply = function(a, b) {
    return a * b;
};

const divide = (a, b) => a / b;

async function fetchData() {
    return 42;
}

class Calculator {
    constructor(initial = 0) {
        this.value = initial;
    }

    add(x) {
        this.value += x;
        return this.value;
    }

    getValue() {
        return this.value;
    }
}
