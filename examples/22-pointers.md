# Pointers

`*T` is a single-item pointer. `*const T` is read-only.

```zig
const std = @import("std");

fn increment(p: *i32) void {
    p.* += 1;
}

pub fn main() void {
    var x: i32 = 10;

    // Take address
    const p: *i32 = &x;

    // Dereference
    std.debug.print("before: {}\n", .{p.*});
    increment(p);
    std.debug.print("after: {}\n", .{x}); // 11

    // Pointer to struct field
    var point = struct { x: f32, y: f32 }{ .x = 1.0, .y = 2.0 };
    const px: *f32 = &point.x;
    px.* = 99.0;
    std.debug.print("point.x = {}\n", .{point.x});

    // Pointer equality
    var a: u8 = 1;
    var b: u8 = 1;
    std.debug.print("same: {}\n", .{&a == &a}); // true
    std.debug.print("diff: {}\n", .{&a == &b}); // false
}
```

---
[← Previous](21-optionals.md) | [Index](../README.md) | [Next →](23-multi-pointers.md)
