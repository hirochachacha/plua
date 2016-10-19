a = {2, 3, 1}
table.sort(a)

assert(a[1] == 1 and a[2] == 2 and a[3] == 3)
table.sort(a, function(i, j) return i > j end)
assert(a[1] == 3 and a[2] == 2 and a[3] == 1)
