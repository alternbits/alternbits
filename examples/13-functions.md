# Functions

Functions are first-class and can be passed as values.

```zig
const std = @import("std");

// Basic function
fn add(a: i32, b: i32) i32 {
    return a + b;
}

// Multiple return values via a struct
fn divmod(a: u32, b: u32) struct { q: u32, r: u32 } {
    return .{ .q = a / b, .r = a % b };
}

// Function taking a function pointer
fn applyTwice(f: fn (i32) i32, x: i32) i32 {
    return f(f(x));
}

fn double(x: i32) i32 {
    return x * 2;
}

pub fn main() void {
    std.debug.print("{}\n", .{add(3, 4)});

    const dm = divmod(17, 5);
    std.debug.print("17 / 5 = {} rem {}\n", .{ dm.q, dm.r });

    std.debug.print("{}\n", .{applyTwice(double, 3)}); // 12

    // Inline function pointer variable
    const op: fn (i32, i32) i32 = add;
    std.debug.print("{}\n", .{op(10, 20)});
}
```

---
[← Previous](12-unions.md) | [Index](../README.md) | [Next →](14-blocks-and-statements.md)
