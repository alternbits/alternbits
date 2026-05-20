# HashMap

`std.AutoHashMap(K, V)` handles integer and pointer keys automatically.

```zig
const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const alloc = gpa.allocator();

    var map = std.AutoHashMap(u32, []const u8).init(alloc);
    defer map.deinit();

    try map.put(1, "one");
    try map.put(2, "two");
    try map.put(3, "three");

    // Lookup
    if (map.get(2)) |val| {
        std.debug.print("2 => {s}\n", .{val});
    }

    // getOrPut — insert default if absent
    const result = try map.getOrPut(4);
    if (!result.found_existing) result.value_ptr.* = "four";
    std.debug.print("4 => {s}\n", .{map.get(4).?});

    // Remove
    _ = map.remove(1);

    // Iterate
    var iter = map.iterator();
    while (iter.next()) |entry| {
        std.debug.print("{} => {s}\n", .{ entry.key_ptr.*, entry.value_ptr.* });
    }
}
```

For string keys use `std.StringHashMap(V)`.

---
[← Previous](28-arraylist.md) | [Index](../README.md) | [Next →](30-linked-list.md)
