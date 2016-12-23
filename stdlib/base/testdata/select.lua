a, b = select(1, 2, 3)
assert(a == 2 and b == 3)

a = select(-1, 2, 3)
assert(a == 3)

a = select("#", 2, 3)
assert(a == 2)

assert(not pcall(select, math.mininteger))
