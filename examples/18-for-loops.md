# For Loops

`for` iterates over arrays, slices, and ranges.

```zig
const std = @import("std");

pub fn main() void {
    const fruits = [_][]const u8{ "apple", "banana", "cherry" };

    // Value only
    for (fruits) |fruit| {
        std.debug.print("{s}\n", .{fruit});
    }

    // Value and index
    for (fruits, 0..) |fruit, i| {
        std.debug.print("{}: {s}\n", .{ i, fruit });
    }

    // Range (start..end, end exclusive)
    for (1..6) |n| {
        std.debug.print("{} ", .{n});
    }
    std.debug.print("\n", .{});

    // Zip two slices of equal length
    const a = [_]i32{ 1, 2, 3 };
    const b = [_]i32{ 10, 20, 30 };
    for (a, b) |x, y| {
        std.debug.print("{} ", .{x + y});
    }
    std.debug.print("\n", .{});

    // break / continue work the same as in while
    for (0..10) |n| {
        if (n % 2 == 0) continue;
        std.debug.print("{} ", .{n});
    }
    std.debug.print("\n", .{});
}
```

---
[← Previous](17-while-loops.md) | [Index](../README.md) | [Next →](19-defer.md)
