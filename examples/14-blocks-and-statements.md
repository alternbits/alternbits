# Blocks and Statements

Blocks are expressions and can return a value with a labeled `break`.

```zig
const std = @import("std");

pub fn main() void {
    // Block as expression
    const x = blk: {
        const a = 10;
        const b = 20;
        break :blk a + b;
    };
    std.debug.print("x = {}\n", .{x}); // 30

    // Nested labeled blocks
    const result = outer: {
        var sum: u32 = 0;
        for (0..10) |i| {
            if (i == 5) break :outer sum;
            sum += @intCast(i);
        }
        break :outer sum;
    };
    std.debug.print("result = {}\n", .{result}); // 0+1+2+3+4 = 10
}
```

---
[← Previous](13-functions.md) | [Index](../README.md) | [Next →](15-if-else.md)
