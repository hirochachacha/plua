f, t, k = utf8.codes("あい")

k, v = f(t, k)
assert(k == 1 and v == 12354)

k, v = f(t, k)
assert(k == 4 and v == 12356)

assert(f(t, k) == nil)
