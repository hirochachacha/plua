assert(utf8.len("abc") == 3)
assert(utf8.len("あい") == 2)
assert(utf8.len("あい", 4) == 1)
assert(utf8.len("あい", 1, 4) == 2)

ok, pos = utf8.len("あい", 2)
assert(not ok and pos == 2)
