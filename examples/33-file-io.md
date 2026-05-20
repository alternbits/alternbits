# File I/O

```zig
const std = @import("std");

pub fn main() !void {
    const alloc = std.heap.page_allocator;

    // Write a file
    {
        const file = try std.fs.cwd().createFile("hello.txt", .{});
        defer file.close();
        try file.writeAll("Hello, Zig!\nLine two.\n");
    }

    // Read entire file
    {
        const contents = try std.fs.cwd().readFileAlloc(alloc, "hello.txt", 4096);
        defer alloc.free(contents);
        std.debug.print("read:\n{s}", .{contents});
    }

    // Append to file
    {
        const file = try std.fs.cwd().openFile("hello.txt", .{ .mode = .read_write });
        defer file.close();
        try file.seekFromEnd(0);
        try file.writeAll("Line three.\n");
    }

    // Read line by line
    {
        const file = try std.fs.cwd().openFile("hello.txt", .{});
        defer file.close();

        var buf_reader = std.io.bufferedReader(file.reader());
        const reader = buf_reader.reader();

        var line_buf: [256]u8 = undefined;
        var line_num: u32 = 0;
        while (try reader.readUntilDelimiterOrEof(&line_buf, '\n')) |line| {
            line_num += 1;
            std.debug.print("{}: {s}\n", .{ line_num, line });
        }
    }

    // Delete
    try std.fs.cwd().deleteFile("hello.txt");
}
```

---
[← Previous](32-formatting-and-print.md) | [Index](../README.md) | [Next →](34-processes.md)
