assert(utf8.offset("あい", 1) == 1)
assert(utf8.offset("あい", 2) == 4)
assert(utf8.offset("あい", 1, 4) == 4)
assert(utf8.offset("あい", 1, 1) == 1)

assert(utf8.offset("あい", 0, 2) == 1)
