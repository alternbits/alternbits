# Switch

`switch` must be exhaustive. It works on integers, enums, and union tags.

```zig
const std = @import("std");

pub fn main() void {
    const n: u8 = 7;

    const desc = switch (n) {
        0       => "zero",
        1, 2, 3 => "small",
        4...9   => "medium",     // inclusive range
        10...99 => "large",
        else    => "huge",
    };
    std.debug.print("{s}\n", .{desc}); // medium

    // Switch on enum
    const Day = enum { mon, tue, wed, thu, fri, sat, sun };
    const day = Day.sat;

    const kind = switch (day) {
        .sat, .sun => "weekend",
        else       => "weekday",
    };
    std.debug.print("{s}\n", .{kind});

    // Switch as expression with capture
    const val: i32 = switch (n) {
        0...4 => @as(i32, n) * 10,
        else  => -1,
    };
    std.debug.print("{}\n", .{val}); // -1
}
```

---
[← Previous](15-if-else.md) | [Index](../README.md) | [Next →](17-while-loops.md)
