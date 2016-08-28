package pattern

import "testing"

import "fmt"

func TestMain(t *testing.T) {
	// fmt.Println(ReplaceString("D/", "D/?.lua;D/?.lc;D/?;D/??x?;D/L", "libs/", -1))

	input := []byte("fake")
	// fmt.Println(Find("fake", input))
	// fmt.Println(Find("(fake)", input))
	// fmt.Println(Find("a(k)e", input))
	// fmt.Println(Find("(f)ake", input))
	// fmt.Println(Find("f.(ke)", input))
	// fmt.Println(Find("fa(ke)", input))
	// fmt.Println(Find("f.ke", input))
	// fmt.Println(Find(".ake", input))
	// fmt.Println(Find("(.a(k)e?)", input))
	// fmt.Println(Find("(.a(k)ef?)", input))
	// fmt.Println(Find("fak", input))
	// fmt.Println(Find("ak", input))

	// fmt.Println(MatchString("fake", string(input)))
	// fmt.Println(MatchString("(fake)", string(input)))
	// fmt.Println(MatchString("a(ke)", string(input)))
	// fmt.Println(MatchString("ak(e)", string(input)))
	// fmt.Println(MatchString("(f)ake", string(input)))
	// fmt.Println(MatchString("f.(ke)", string(input)))
	fmt.Println(MatchString("fa(ke)", string(input)))
	fmt.Println(ReplaceFuncString("fa(ke)", string(input), func(s string) string { return s + s }, -1))
	// fmt.Println(MatchString("f.ke", string(input)))
	// fmt.Println(MatchString(".ake", string(input)))
	// fmt.Println(MatchString("(.a(k)e?)", string(input)))
	// fmt.Println(MatchString("(.a(k)ef?)", string(input)))
	// fmt.Println(MatchString("fak", string(input)))
	// fmt.Println(MatchString("ak", string(input)))

	// fmt.Println(MatchString("a+", []byte("faaaaa")))
	// fmt.Println(MatchString("a+", []byte("aaaaa")))
	// fmt.Println(MatchString("a*", []byte("aaaaa")))
	// fmt.Println(MatchString("a-", []byte("aaaaa")))
	// fmt.Println(MatchString("a*", []byte("faaaaa")))
	// fmt.Println(MatchString("a*", []byte("aaaaa")))
	// fmt.Println(MatchString("a-", []byte("aaaaa")))

	// fmt.Println(MatchString(".*fa", []byte("fafa")))
	// fmt.Println(MatchString(".-fa", []byte("fafa")))

	// fmt.Println(MatchString("%d+", []byte("fafa11223dd")))
	// fmt.Println(MatchString("()aa()", []byte("flaap")))
	// fmt.Println(Find("()aa()", []byte("flaap")))

	// fmt.Println(MatchString("%b()", []byte("r(f(aa221afaka)a(roo)ta)a")))
	// fmt.Println(MatchString("%b()", []byte("a(b(c)d(e)f")))
	// fmt.Println(Find("%b()", []byte("a(b(c)d(e)f")))

	// fmt.Println(Find("%f[123]", []byte("aa111123")))
	// fmt.Println(FindString("(.ake)%1", "fakefake"))
	// fmt.Println(MatchString("(.ake)%1", "fakefakefakefake"))
	// fmt.Println(MatchStringAll("(.ake)%1", "fakefakefakefake"))
	// fmt.Println(FindString("(.ake)%1", "fakefakefakefake"))
	// fmt.Println(FindStringAll("(.ake)%1", "fakefakefakefake"))
	// fmt.Println(FindStringAll("(.ake)fake", "fakefakefakefake"))
	// fmt.Println(FindStringAll("(.ake)%1", "fakefakefakefake"))
	// fmt.Println(FindStringAll("(.ake)%1", "fakefake fakefake"))
	// fmt.Println(FindStringAll(".ake", "fakefakefakefake"))
	// fmt.Println(FindStringAll("fake", "fakefakefakefake"))
	// fmt.Println(FindString(".ake", "fakefakefakefake"))
	// fmt.Println(Find("(.ake)%1", []byte("fakefake")))
	// fmt.Println(FindString("\n", config))
	// fmt.Println(FindStringAll("\n", config))
	// fmt.Println(FindAll("||", []byte(art)))
	// fmt.Println(ReplaceString("\n", config, "|", -1))
}
