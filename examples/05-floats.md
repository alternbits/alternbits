# Floats

```zig
const std = @import("std");
const math = std.math;

pub fn main() void {
    const a: f32 = 1.5;
    const b: f64 = 3.14159265358979;
    const c: f128 = 1.0 / 3.0;

    std.debug.print("f32: {}\n", .{a});
    std.debug.print("f64: {d:.5}\n", .{b});
    std.debug.print("f128: {d:.10}\n", .{c});

    std.debug.print("sqrt(2) = {d:.6}\n", .{math.sqrt(@as(f64, 2.0))});
    std.debug.print("inf: {}\n", .{math.inf(f64)});
    std.debug.print("nan: {}\n", .{math.nan(f64)});
}
```

---
[← Previous](04-integers.md) | [Index](../README.md) | [Next →](06-strings.md)
