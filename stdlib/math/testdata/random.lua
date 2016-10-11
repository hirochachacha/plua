i = math.random()
assert(0 <= i and i <= 1)

i = math.random(4)
assert(1 <= i and i <= 4)

i = math.random(4, 9)
assert(4 <= i and i <= 9)
