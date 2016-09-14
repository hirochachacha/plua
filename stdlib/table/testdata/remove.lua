a = {1, 2, 3}
assert(table.remove(a, 4) == nil)
assert(table.remove(a, 1) == 1 and a[1] == 2)
a = {1, 2, 3}
assert(table.remove(a) == 3 and a[3] == nil)
