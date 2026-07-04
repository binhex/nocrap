int range_for(int* items, int len) {
    int sum = 0;
    for (int i = 0; i < len; i++) {
        sum += items[i];
    }
    return sum;
}

int multi_catch(int a, int b) {
    try {
        return a / b;
    } catch (int e) {
        return 0;
    } catch (double e) {
        return -1;
    }
}

void with_lambda() {
    auto fn = []() { return 42; };
    if (fn()) {
        (void)0;
    }
}
