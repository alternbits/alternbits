# Linked List

Zig's standard library provides an intrusive doubly-linked list.

```zig
const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const alloc = gpa.allocator();

    var list = std.DoublyLinkedList(i32){};

    // Allocate nodes
    var a = try alloc.create(std.DoublyLinkedList(i32).Node);
    defer alloc.destroy(a);
    a.data = 1;

    var b = try alloc.create(std.DoublyLinkedList(i32).Node);
    defer alloc.destroy(b);
    b.data = 2;

    list.append(a);
    list.append(b);

    std.debug.print("len: {}\n", .{list.len});

    var node = list.first;
    while (node) |n| : (node = n.next) {
        std.debug.print("{}\n", .{n.data});
    }
}
```

---
[← Previous](29-hashmap.md) | [Index](../README.md) | [Next →](31-testing.md)
