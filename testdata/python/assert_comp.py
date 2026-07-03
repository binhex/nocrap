def process_items(items, threshold, labels):
    """Process items with filtering and validation."""
    assert threshold >= 0, "Threshold must be non-negative"

    valid = [
        x * 2
        for x in items
        if x > threshold
    ]

    if not valid:
        return None

    label = labels.get(valid[0], "default") if valid else "none"

    return {
        key: value
        for key, value in label.items()
        if value is not None
    }


def nullable(value):
    """Check if value is None."""
    return value is None
