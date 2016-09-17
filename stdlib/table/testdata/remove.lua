a = {1, 2, 3}
assert(table.remove(a, 4) == nil)
assert(table.remove(a, 1) == 1 and a[1] == 2)
a = {1, 2, 3}
assert(table.remove(a) == 3 and a[3] == nil)
assert(table.remove(a) == 2)
assert(table.remove(a) == 1)
assert(table.remove(a) == nil)

a = {1, 2, 3}
assert(table.remove(a, 1) == 1)
assert(table.remove(a, 1) == 2)
assert(table.remove(a, 1) == 3)
assert(table.remove(a, 1) == nil)

a = {1, 2, 3}
assert(table.remove(a, 3) == 3)
assert(table.remove(a, 3) == nil)
