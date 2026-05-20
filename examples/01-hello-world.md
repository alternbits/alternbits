# Hello, World

The entry point of every Zig program is `pub fn main`.

```zig
const std = @import("std");

pub fn main() void {
    std.debug.print("Hello, World!\n", .{});
}
```

`@import` is a built-in that loads a package. `std` is the standard library.  
`std.debug.print` writes to stderr. For stdout, use a `Writer` (shown later).

```
$ zig run hello.zig
Hello, World!
```

---
[Index](../README.md) | [Next →](02-values.md)
