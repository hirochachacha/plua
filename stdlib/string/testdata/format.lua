x = string.format("%q", 'a string with "quotes"')
y = '"a string with \\"quotes\\""'

assert(x == y)

assert(string.format("%05d", 10) == "00010")
assert(string.format("%x", 10123324) == "9a783c")
