# Variables

`var` declares a mutable binding. All variables must be initialized.

```zig
const std = @import("std");

pub fn main() void {
    var x: i32 = 1;
    x += 1;
    std.debug.print("x = {}\n", .{x});

    // Type inference with var
    var y = @as(f32, 2.5);
    y *= 2.0;
    std.debug.print("y = {}\n", .{y});
}
```

Zig will refuse to compile a `var` that is never mutated — use `const` instead.

---
[← Previous](02-values.md) | [Index](../README.md) | [Next →](04-integers.md)
