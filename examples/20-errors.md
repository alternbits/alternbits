# Errors

Zig errors are values, not exceptions. An error union type is `ErrorSet!T`.

```zig
const std = @import("std");

// Define an error set
const ParseError = error{
    InvalidCharacter,
    Overflow,
    Empty,
};

fn parsePositive(s: []const u8) ParseError!u32 {
    if (s.len == 0) return ParseError.Empty;
    var result: u32 = 0;
    for (s) |c| {
        if (c < '0' or c > '9') return ParseError.InvalidCharacter;
        result = result * 10 + (c - '0');
    }
    return result;
}

pub fn main() void {
    // try — returns the error from the current function on failure
    // (here, main returns void so we use catch instead)

    // catch — handle or transform the error
    const n = parsePositive("123") catch |err| {
        std.debug.print("error: {}\n", .{err});
        return;
    };
    std.debug.print("parsed: {}\n", .{n});

    // catch with a default value
    const m = parsePositive("abc") catch 0;
    std.debug.print("default: {}\n", .{m});

    // switch on error
    const result = parsePositive("");
    switch (result) {
        error.Empty            => std.debug.print("empty input\n", .{}),
        error.InvalidCharacter => std.debug.print("bad char\n", .{}),
        error.Overflow         => std.debug.print("overflow\n", .{}),
        else                   => |v| std.debug.print("ok: {}\n", .{v}),
    }
}
```

---
[← Previous](19-defer.md) | [Index](../README.md) | [Next →](21-optionals.md)
