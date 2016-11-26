assert(utf8.codepoint("あい") == 12354)
assert(utf8.codepoint("あい", 1) == 12354)
assert(utf8.codepoint("あい", -3) == 12356)
a, i = utf8.codepoint("あい", 1, 4); assert(a == 12354 and i == 12356)
a, i = utf8.codepoint("あい", 1, -1); assert(a == 12354 and i == 12356)

ok, err = pcall(utf8.codepoint, "あい", 2)
assert(not ok)
assert(err == "invalid UTF-8 code")
assert(not pcall(utf8.codepoint, "あい", 0))

assert(utf8.codepoint("", 1, -1) == nil)
