# Testing

Zig has first-class testing with `test` blocks.

```zig
const std = @import("std");
const testing = std.testing;

fn factorial(n: u64) u64 {
    if (n == 0) return 1;
    return n * factorial(n - 1);
}

test "factorial of 0" {
    try testing.expectEqual(@as(u64, 1), factorial(0));
}

test "factorial of 5" {
    try testing.expectEqual(@as(u64, 120), factorial(5));
}

test "factorial of 10" {
    try testing.expectEqual(@as(u64, 3_628_800), factorial(10));
}

// Run all tests in this file with:
// zig test factorial.zig
```

```
$ zig test factorial.zig
All 3 tests passed.
```

Tests can use allocators too:

```zig
test "arraylist" {
    var list = std.ArrayList(i32).init(testing.allocator);
    defer list.deinit();

    try list.append(1);
    try list.append(2);
    try testing.expectEqual(@as(usize, 2), list.items.len);
}
```

`testing.allocator` detects leaks at the end of each test.

---
[← Previous](30-linked-list.md) | [Index](../README.md) | [Next →](32-formatting-and-print.md)
