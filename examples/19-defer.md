# Defer

`defer` runs a statement when the enclosing block exits, in LIFO order.

```zig
const std = @import("std");

fn riskyOp(fail: bool) !void {
    std.debug.print("start\n", .{});
    defer std.debug.print("cleanup (always runs)\n", .{});

    if (fail) return error.Oops;

    std.debug.print("success\n", .{});
}

pub fn main() void {
    // defer stack — prints 3, 2, 1
    {
        defer std.debug.print("1\n", .{});
        defer std.debug.print("2\n", .{});
        defer std.debug.print("3\n", .{});
    }

    riskyOp(false) catch {};
    riskyOp(true)  catch {};
}
```

`errdefer` runs only when the enclosing function returns an error.

```zig
fn openFile() !void {
    const file = try acquireResource();
    errdefer releaseResource(file); // only on error path
    try doWork(file);
    releaseResource(file);          // normal path
}
```

---
[← Previous](18-for-loops.md) | [Index](../README.md) | [Next →](20-errors.md)
