# Random Numbers

```zig
const std = @import("std");

pub fn main() void {
    // Seed from the OS
    var prng = std.rand.DefaultPrng.init(blk: {
        var seed: u64 = undefined;
        std.posix.getrandom(std.mem.asBytes(&seed)) catch unreachable;
        break :blk seed;
    });
    const rand = prng.random();

    // Random integer in [0, max)
    const n = rand.intRangeAtMost(u32, 1, 100);
    std.debug.print("dice: {}\n", .{n});

    // Random float in [0.0, 1.0)
    const f = rand.float(f64);
    std.debug.print("float: {d:.4}\n", .{f});

    // Shuffle a slice
    var arr = [_]u8{ 1, 2, 3, 4, 5, 6, 7, 8 };
    rand.shuffle(u8, &arr);
    std.debug.print("shuffled: {any}\n", .{arr});
}
```

---
[← Previous](35-json.md) | [Index](../README.md) | [Next →](37-sorting.md)
