# JSON

```zig
const std = @import("std");

const Config = struct {
    host: []const u8,
    port: u16,
    debug: bool,
};

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const alloc = gpa.allocator();

    // Parse JSON into a typed struct
    const json_str =
        \\{"host": "localhost", "port": 8080, "debug": true}
    ;

    const parsed = try std.json.parseFromSlice(Config, alloc, json_str, .{});
    defer parsed.deinit();

    const cfg = parsed.value;
    std.debug.print("host: {s}\n", .{cfg.host});
    std.debug.print("port: {}\n", .{cfg.port});
    std.debug.print("debug: {}\n", .{cfg.debug});

    // Parse into a dynamic value
    const dynamic = try std.json.parseFromSlice(std.json.Value, alloc, json_str, .{});
    defer dynamic.deinit();

    if (dynamic.value.object.get("port")) |port| {
        std.debug.print("dynamic port: {}\n", .{port.integer});
    }

    // Stringify a struct
    var out = std.ArrayList(u8).init(alloc);
    defer out.deinit();

    try std.json.stringify(cfg, .{ .whitespace = .indent_2 }, out.writer());
    std.debug.print("{s}\n", .{out.items});
}
```

---
[← Previous](34-processes.md) | [Index](../README.md) | [Next →](36-random-numbers.md)
