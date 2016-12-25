ch = goroutine.newchannel(1)

ch:send(nil)

val, ok  = ch:recv()

assert(ok and val == nil)

ch:send("test")

val, ok  = ch:recv()

assert(ok and val == "test")

ch:close()

val, ok  = ch:recv()

assert(not ok and val == nil)
