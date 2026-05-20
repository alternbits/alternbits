# ArrayList

`std.ArrayList(T)` is a growable array backed by an allocator.

```zig
const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();

    var list = std.ArrayList(i32).init(gpa.allocator());
    defer list.deinit();

    try list.append(10);
    try list.append(20);
    try list.append(30);
    try list.appendSlice(&.{ 40, 50 });

    std.debug.print("len: {}\n", .{list.items.len});
    std.debug.print("third: {}\n", .{list.items[2]});

    // Remove by index (swap-removes for O(1))
    _ = list.swapRemove(0);
    std.debug.print("after remove: {any}\n", .{list.items});

    // Convert to owned slice
    const owned = try list.toOwnedSlice();
    defer gpa.allocator().free(owned);
    std.debug.print("owned: {any}\n", .{owned});
}
```

---
[← Previous](27-memory-allocation.md) | [Index](../README.md) | [Next →](29-hashmap.md)
