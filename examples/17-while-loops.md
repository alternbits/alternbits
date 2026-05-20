# While Loops

```zig
const std = @import("std");

pub fn main() void {
    // Basic while
    var i: u32 = 0;
    while (i < 5) : (i += 1) {
        std.debug.print("{} ", .{i});
    }
    std.debug.print("\n", .{});

    // While with continue/break
    var n: u32 = 0;
    while (true) {
        n += 1;
        if (n == 3) continue;
        if (n == 6) break;
        std.debug.print("{} ", .{n});
    }
    std.debug.print("\n", .{});

    // While as expression (returns value on break)
    var count: u32 = 0;
    const found = while (count < 100) : (count += 1) {
        if (count * count > 50) break count;
    } else 0;
    std.debug.print("found: {}\n", .{found}); // 8

    // Unwrap optional with while
    const items = [_]?u32{ 1, 2, null, 3 };
    var idx: usize = 0;
    while (idx < items.len) : (idx += 1) {
        if (items[idx]) |v| std.debug.print("item: {}\n", .{v});
    }
}
```

---
[← Previous](16-switch.md) | [Index](../README.md) | [Next →](18-for-loops.md)
