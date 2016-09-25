co = coroutine.create(function (x)
  assert(x == 10)
  local a, b = coroutine.yield(x)
  assert(a == 111 and b == 120)
  return x
end)

ok, ret = coroutine.resume(co, 10)
assert(ok and ret == 10)

ok, ret = coroutine.resume(co, 111, 120)
assert(ok and ret == 10)

ok, ret = coroutine.resume(co, 1)
assert(not ok and ret ~= "")
