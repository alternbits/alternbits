# Unions

A union stores one value from a set of types. Tagged unions carry the active tag.

```zig
const std = @import("std");

// Bare union — you track the active field yourself
const FloatOrInt = union {
    float: f64,
    int: i64,
};

// Tagged union — the compiler tracks the active field
const Value = union(enum) {
    int: i64,
    float: f64,
    boolean: bool,
    none,

    pub fn print(self: Value) void {
        switch (self) {
            .int     => |v| std.debug.print("int: {}\n", .{v}),
            .float   => |v| std.debug.print("float: {}\n", .{v}),
            .boolean => |v| std.debug.print("bool: {}\n", .{v}),
            .none    =>     std.debug.print("none\n", .{}),
        }
    }
};

pub fn main() void {
    var v = Value{ .int = 42 };
    v.print();
    v = .{ .float = 3.14 };
    v.print();
    v = .none;
    v.print();
}
```

---
[← Previous](11-enums.md) | [Index](../README.md) | [Next →](13-functions.md)
