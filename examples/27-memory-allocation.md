# Memory Allocation

Zig has no implicit allocations. You choose the allocator.

```zig
const std = @import("std");

pub fn main() !void {
    // GeneralPurposeAllocator detects leaks and double-frees in debug builds
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    // Allocate a slice
    const buf = try allocator.alloc(u8, 64);
    defer allocator.free(buf);

    @memset(buf, 0);
    buf[0] = 'Z';
    buf[1] = 'i';
    buf[2] = 'g';
    std.debug.print("{s}\n", .{buf[0..3]});

    // Allocate a single value
    const p = try allocator.create(i32);
    defer allocator.destroy(p);
    p.* = 42;
    std.debug.print("{}\n", .{p.*});
}
```

Common allocators:
- `std.heap.page_allocator` — direct OS pages, no overhead tracking
- `std.heap.GeneralPurposeAllocator` — debug-friendly, leak detection
- `std.heap.ArenaAllocator` — free everything at once
- `std.heap.FixedBufferAllocator` — allocates from a stack buffer, no OS call

---
[← Previous](26-generics.md) | [Index](../README.md) | [Next →](28-arraylist.md)
