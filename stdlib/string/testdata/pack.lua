assert(string.pack(">i9", 1) == "\0\0\0\0\0\0\0\0\1")
assert(string.pack("<i9", 1) == "\1\0\0\0\0\0\0\0\0")
assert(string.pack(">i10", -3) == "\255\255\255\255\255\255\255\255\255\253")
assert(string.pack("<!4 z Xi4", "ab") == "ab\0\0")
