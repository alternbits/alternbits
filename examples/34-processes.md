# Processes

```zig
const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const alloc = gpa.allocator();

    // Run a child process, capture output
    const result = try std.process.Child.run(.{
        .allocator = alloc,
        .argv = &.{ "uname", "-s" },
    });
    defer alloc.free(result.stdout);
    defer alloc.free(result.stderr);

    std.debug.print("os: {s}", .{result.stdout});
    std.debug.print("exit: {}\n", .{result.term.Exited});

    // Read environment variable
    const path = std.process.getEnvVarOwned(alloc, "PATH") catch "not set";
    defer if (!std.mem.eql(u8, path, "not set")) alloc.free(path);
    std.debug.print("PATH length: {}\n", .{path.len});

    // Exit with code
    // std.process.exit(0);
}
```

---
[← Previous](33-file-io.md) | [Index](../README.md) | [Next →](35-json.md)
