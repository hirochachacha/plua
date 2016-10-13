a = {1, 2, 3}
table.insert(a, 4); assert(a[4] == 4)
table.insert(a, 2, 0); assert(a[2] == 0)

assert(not pcall(table.insert, {}, 2, 1))
assert(not pcall(table.insert, {}, 0, 1))
