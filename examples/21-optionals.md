# Optionals

An optional `?T` holds either a value of type `T` or `null`. It is not a pointer.

```zig
const std = @import("std");

fn findFirst(haystack: []const u8, needle: u8) ?usize {
    for (haystack, 0..) |c, i| {
        if (c == needle) return i;
    }
    return null;
}

pub fn main() void {
    const s = "hello world";

    // if capture
    if (findFirst(s, 'o')) |idx| {
        std.debug.print("found at {}\n", .{idx}); // 4
    }

    // orelse — provide a fallback
    const idx = findFirst(s, 'z') orelse s.len;
    std.debug.print("idx: {}\n", .{idx});

    // .? — unwrap or panic
    const guaranteed = findFirst(s, 'e').?;
    std.debug.print("guaranteed: {}\n", .{guaranteed});

    // Chaining with orelse and error handling
    const maybe: ?i32 = null;
    const val = maybe orelse 99;
    std.debug.print("val: {}\n", .{val});
}
```

---
[← Previous](20-errors.md) | [Index](../README.md) | [Next →](22-pointers.md)
