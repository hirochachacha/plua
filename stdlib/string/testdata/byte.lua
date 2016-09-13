assert(string.byte("a") == 97)

a, b, c = string.byte("abcdefghi", 1, 3)

assert(a == 97 and b == 98 and c == 99)

h, i = string.byte("abcdefghi", -2, -1)

assert(h == 104 and i == 105)
