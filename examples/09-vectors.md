# Vectors

SIMD vectors allow data-parallel math. The size must be a power of two.

```zig
const std = @import("std");

pub fn main() void {
    const a: @Vector(4, f32) = .{ 1, 2, 3, 4 };
    const b: @Vector(4, f32) = .{ 10, 20, 30, 40 };

    const c = a + b;
    std.debug.print("{}\n", .{c}); // { 11, 22, 33, 44 }

    // Reduce to scalar
    const total = @reduce(.Add, c);
    std.debug.print("sum: {}\n", .{total}); // 110
}
```

---
[← Previous](08-slices.md) | [Index](../README.md) | [Next →](10-structs.md)
