x, y = string.find("abcabc", "bca")

assert(x == 2 and y == 4)

x, y = string.find("abcabc", "b", 3)

assert(x == 5 and y == 5)

x, y = string.find("abcabc", "$")

assert(x == 7 and y == 6)

x, y = string.find("abcabc", "^")

assert(x == 1 and y == 0)

x, y = string.find("abcabc", ".c")

assert(x == 2 and y == 3)

x, y = string.find(" abc abc", "%w+")

assert(x == 2 and y == 4)

x, y = string.find("abc", "^[a-z]+$")

assert(x == 1 and y == 3)

assert(string.find("()", "^[a-z]+$") == nil)

x, y = string.find("", "")

assert(x == 1 and y == 0)
