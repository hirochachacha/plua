assert(utf8.offset("あい", 1) == 1)
assert(utf8.offset("あい", 2) == 4)
assert(utf8.offset("あい", 1, 4) == 4)
assert(utf8.offset("あい", 1, 1) == 1)
assert(utf8.offset("あい", -1, 7) == 4)

assert(utf8.offset("あい", 0, 1) == 1)
assert(utf8.offset("あい", 0, 2) == 1)
assert(utf8.offset("あい", 0, 3) == 1)
assert(utf8.offset("あい", 0, 4) == 4)
assert(utf8.offset("あい", 0, 5) == 4)
assert(utf8.offset("あい", 0, 6) == 4)
assert(utf8.offset("あい", 0, 7) == 7)

assert(not pcall(utf8.offset, "あい", 1, 2))
assert(not pcall(utf8.offset, "あい", 1, 3))
assert(not pcall(utf8.offset, "あい", 1, 5))
assert(not pcall(utf8.offset, "あい", 1, 6))

assert(utf8.offset("a", 2, 1) == 2)
