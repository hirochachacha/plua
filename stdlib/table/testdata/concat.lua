assert(table.concat({1, 2, 3}) == "123")
assert(table.concat({1, 2, 3}, ", ") == "1, 2, 3")
assert(table.concat({1, 2, 3}, "", 2) == "23")
assert(table.concat({1, 2, 3}, "", 2, 2) == "2")
