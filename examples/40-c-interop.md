# C Interop

Zig can call C libraries directly with zero overhead and no FFI layer.

```zig
// math_demo.zig
const std = @import("std");
const c = @cImport({
    @cInclude("math.h");
    @cInclude("stdio.h");
});

pub fn main() void {
    const result = c.sqrt(2.0);
    std.debug.print("C sqrt(2) = {d:.6}\n", .{result});

    // Calling C printf directly
    _ = c.printf("Hello from C: %d\n", @as(c_int, 42));
}
```

```
$ zig run math_demo.zig -lc
```

Exporting Zig functions to C:

```zig
// library.zig
export fn add(a: c_int, b: c_int) c_int {
    return a + b;
}
```

```
$ zig build-lib library.zig -dynamic
```

Linking a system library in `build.zig`:

```zig
exe.linkSystemLibrary("sqlite3");
exe.linkLibC();
```

---
[← Previous](39-build-system.md) | [Index](../README.md)
