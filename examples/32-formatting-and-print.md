# Formatting and Print

```zig
const std = @import("std");

pub fn main() !void {
    const stdout = std.io.getStdOut().writer();

    // Basic
    try stdout.print("Hello, {s}!\n", .{"world"});

    // Integer formats
    try stdout.print("decimal: {d}\n", .{255});
    try stdout.print("hex:     {x}\n", .{255});       // ff
    try stdout.print("HEX:     {X}\n", .{255});       // FF
    try stdout.print("octal:   {o}\n", .{255});       // 377
    try stdout.print("binary:  {b}\n", .{255});       // 11111111

    // Float formats
    try stdout.print("float:   {}\n",    .{3.14159});
    try stdout.print("fixed:   {d:.2}\n", .{3.14159}); // 3.14
    try stdout.print("sci:     {e:.2}\n", .{12345.0}); // 1.23e+04

    // Width and padding
    try stdout.print("|{d:>10}|\n", .{42});    // right-align
    try stdout.print("|{d:<10}|\n", .{42});    // left-align
    try stdout.print("|{d:0>6}|\n",  .{42});   // zero-pad

    // Any value
    try stdout.print("{any}\n", .{[_]u8{ 1, 2, 3 }});

    // Format to a buffer
    var buf: [64]u8 = undefined;
    const s = try std.fmt.bufPrint(&buf, "{d} + {d} = {d}", .{ 1, 2, 3 });
    try stdout.print("{s}\n", .{s});

    // Format to an allocated string
    const alloc = std.heap.page_allocator;
    const msg = try std.fmt.allocPrint(alloc, "value = {d}", .{42});
    defer alloc.free(msg);
    try stdout.print("{s}\n", .{msg});
}
```

---
[← Previous](31-testing.md) | [Index](../README.md) | [Next →](33-file-io.md)
