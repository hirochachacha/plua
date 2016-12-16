a = {1, 2, 3}
a1 = table.move(a, 2, 3, 1)
assert(a[1] == 2 and a[2] == 3 and a[3] == 3)
assert(a1[1] == 2 and a1[2] == 3 and a1[3] == 3)

a = {1, 2, 3}
b = {4, 5, 6}
a1 = table.move(b, 1, 3, 4, a)
assert(a[1] == 1 and a[2] == 2 and a[3] == 3 and a[4] == 4 and a[5] == 5 and a[6] == 6)
assert(a1[1] == 1 and a1[2] == 2 and a1[3] == 3 and a1[4] == 4 and a1[5] == 5 and a1[6] == 6)

a = {1, 2, 3}
table.move(a, 1, 3, 2)
assert(a[1] == 1 and a[2] == 1 and a[3] == 2 and a[4] == 3)
