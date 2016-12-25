-- example code is taken from https://tour.golang.org/concurrency/5

fibs = {0, 1, 1, 2, 3, 5, 8, 13, 21, 34}

function fibonacci(ch, quit)
  local x, y = 0, 1
  while true do
    local chosen, recv, recvOK = goroutine.select(
      goroutine.case("send", ch, x),
      goroutine.case("recv", quit)
    )

    if chosen == 1 then
      x, y = y, x+y
    elseif chosen == 2 then
      return
    end
  end
end

ch = goroutine.newchannel()
quit = goroutine.newchannel()

goroutine.wrap(function()
  for i = 1, 10 do
    assert(ch:recv() == fibs[i])
  end
  quit:send(nil)
end)()

fibonacci(ch, quit)
