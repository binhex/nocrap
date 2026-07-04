// Optional chaining and nullish coalescing for CC testing
function optionalChaining(obj, fallback) {
    const a = obj?.prop?.nested;
    const b = obj?.method?.() ?? fallback;
    const c = a ?? "default";
    return [a, b, c];
}
