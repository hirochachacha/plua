t = table.pack(); assert(#t == 0)
t = table.pack(1, 2, 3); assert(t[1] == 1 and t[2] == 2 and t[3] == 3)
