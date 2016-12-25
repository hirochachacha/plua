-- example code is taken from https://tour.golang.org/concurrency/5

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
      print("quit")
      return
    end
  end
end

ch = goroutine.newchannel()
quit = goroutine.newchannel()

goroutine.wrap(function()
  for i = 1, 10 do
    print(ch:recv())
  end
  quit:send(nil)
end)()

fibonacci(ch, quit)
