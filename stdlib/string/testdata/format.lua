x = string.format("%q", 'a string with "quotes"')
y = '"a string with \\"quotes\\""'

assert(x == y)

assert(string.format("%05d", 10) == "00010")
assert(string.format("%x", 10123324) == "9a783c")

assert(string.format('%q', '\\') == '"\\\\"')

assert('\n' == '\n')
assert(string.format('%q', '\n') == '"\\\n"')

assert('\x10' == "\16")
assert(string.format('%q', '\x10') == '"\\16"')

assert('\u{100}' == 'Ā')
assert(string.format('%q', '\u{100}') == '"Ā"')

assert('\u{22222}' == '\u{22222}')
assert(string.format('%q','\u{22222}') == '"\u{22222}"')

assert('\100' == "\100")
assert(string.format('%q', '\100') == '"d"')

assert('\0' == "\0")
assert(string.format('%q', '\0') == '"\\0"')

assert('\01' == "\01")
assert(string.format('%q', '\01') == '"\\1"')

assert('\014' == "\014")
assert(string.format('%q', '\014') == '"\\14"')

assert(tonumber(string.format("%f", 10.3)) == 10.3)

assert(string.format("%.3s", "12345") == "123")
