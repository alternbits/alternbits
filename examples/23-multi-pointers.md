# Multi-Pointers

`[*]T` is a many-item pointer — like a C pointer. It has no length.

```zig
const std = @import("std");

pub fn main() void {
    var arr = [_]u32{ 10, 20, 30, 40 };

    // Coerce array to many-pointer
    const p: [*]u32 = &arr;

    std.debug.print("{}\n", .{p[0]}); // 10
    std.debug.print("{}\n", .{p[2]}); // 30

    // Arithmetic
    const p2 = p + 1;
    std.debug.print("{}\n", .{p2[0]}); // 20

    // Convert back to a slice with known length
    const s: []u32 = p[0..arr.len];
    std.debug.print("len: {}\n", .{s.len});
}
```

Use `[*]T` when interoperating with C APIs. Prefer slices (`[]T`) in pure Zig code.

---
[← Previous](22-pointers.md) | [Index](../README.md) | [Next →](24-slices-pointers.md)
