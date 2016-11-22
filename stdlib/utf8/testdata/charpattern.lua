local str = "ああ"
for r in str:gmatch(utf8.charpattern ) do
	assert(r == "あ")
end
