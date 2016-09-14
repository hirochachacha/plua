a = {1, 2, 3}
table.move(a, 2, 3, 1)
assert(a[1] == 2 and a[2] == 3 and a[3] == 3)

a = {1, 2, 3}
b = {4, 5, 6}
table.move(b, 1, 3, 4, a)
assert(a[1] == 1 and a[2] == 2 and a[3] == 3 and a[4] == 4 and a[5] == 5 and a[6] == 6)

