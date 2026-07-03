function process(items) {
    var count = 0;
    for (var i = 0; i < items.length; i++) {
        var x = items[i];
        if (x > 0) {
            count++;
        }
        if (x < 0) {
            count--;
        }
    }
    return count;
}
