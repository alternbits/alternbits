# Integers

Zig integers are sized and either signed (`i`) or unsigned (`u`). Any bit-width from 0 to 65535 is valid.

```zig
const std = @import("std");

pub fn main() void {
    const a: u8  = 255;
    const b: i8  = -128;
    const c: u32 = 4_294_967_295;
    const d: i64 = -9_223_372_036_854_775_808;

    // Overflow is a compile error on const, and a detectable runtime error on var
    // in Debug mode. Use wrapping operators to opt in: +%, -%, *%
    var e: u8 = 200;
    e +%= 100; // wraps to 44

    std.debug.print("{} {} {} {} {}\n", .{ a, b, c, d, e });

    // Casting
    const big: u32 = 300;
    const small: u8 = @truncate(big); // 44
    std.debug.print("truncated: {}\n", .{small});
}
```

---
[← Previous](03-variables.md) | [Index](../README.md) | [Next →](05-floats.md)
