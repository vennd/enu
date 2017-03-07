package mneumonic

import (
	"strings"
	"testing"
)

func TestGenerateRandom(t *testing.T) {
	m := GenerateRandom(128)
	if len(m.ToHex()) != 32 {
		t.Errorf("Expected: length of hex string from 128 random bits is 32, got: %d\n", len(m.ToHex()))
	}
	if len(m.ToWords()) != 12 {
		t.Errorf("Expected: number of words in mneumonic from 128 random bits is 12, got: %d\n", len(m.ToWords()))
	}
}

var testdata = []struct {
	wordArray []string
	hexString string
}{
	{[]string{"like", "like", "like", "like", "like", "like", "like", "like", "like", "like", "like", "like"}, "00000000000000000000000000000000"},
	{[]string{"weary", "weary", "weary", "weary", "weary", "weary", "weary", "weary", "weary", "weary", "weary", "weary"}, "00000659000006590000065900000659"},
	{[]string{"harmony", "beside", "common", "between", "count", "thrown", "stir", "clothes", "mystery", "quick", "glance", "ready"}, "8609d12d5c64d32a23e0149ed99edada"},
	{[]string{"knee", "knowledge", "nothing", "seven", "spill", "hill", "crush", "million", "treat", "shock", "screw", "confusion"}, "855c400ff0efd64cfeac24e7fd977560"},
	{[]string{"respect", "plain", "line", "softly", "ocean", "grandma", "strife", "fly", "selfish", "against", "whether", "eat"}, "76da27a0824d8664759799c5959e9466"},
	{[]string{"descend", "ruin", "inside", "crimson", "fact", "put", "innocent", "salt", "save", "carefully", "size", "love"}, "8b935afde55bf8a54c0689ab5a49dfdf"},
	{[]string{"ground", "grew", "everyone", "felicity", "soak", "mix", "victim", "raise", "understand", "normal", "sometimes", "drama"}, "c8f3f675bc77d486b531079dcf012246"},
	{[]string{"shift", "grief", "pants", "respond", "enough", "respond", "roll", "rule", "shirt", "danger", "busy", "welcome"}, "dc276397ca0f71d5f9f6c1bbaa2dc368"},
	{[]string{"repeat", "guilty", "street", "student", "constantly", "burst", "anyone", "curve", "anyone", "course", "commit", "student"}, "318deb5303c496898c0251b90363c9c4"},
	{[]string{"forget", "such", "grass", "stolen", "blink", "tender", "itself", "witch", "against", "home", "make", "strife"}, "4206ad37410fc091325a87e2eaa378f1"},
	{[]string{"birth", "party", "beside", "joke", "total", "finger", "yet", "dirty", "accept", "asleep", "next", "knife"}, "fc4b10f02caf3fdfd68c5e6c37161855"},
	{[]string{"yesterday", "opposite", "weight", "ball", "step", "back", "shape", "mention", "bit", "conversation", "valley", "choice"}, "7a77ed4edd3a95323ddf24c27733f833"},
	{[]string{"compare", "finger", "age", "poet", "paper", "coffee", "worst", "mirror", "frame", "planet", "quickly", "bee"}, "0edd8cb378a4e2bc54c3c0366b69a490"},
	{[]string{"pattern", "dwell", "ache", "retreat", "lick", "country", "spirit", "few", "shimmer", "survive", "gotten", "visit"}, "815b68b77dc1105e923aa719d5ba877f"},
	{[]string{"insane", "neighbor", "mom", "power", "fix", "bone", "bottle", "course", "worst", "bitter", "enough", "corner"}, "6e58cfadd243634f273706c52598e99d"},
	{[]string{"lick", "start", "lovely", "grand", "liquid", "capture", "disappear", "asleep", "existence", "draw", "advice", "glorious"}, "6efa3caa0ddf2f87345ad27424241a7b"},
	{[]string{"idea", "rebel", "beauty", "led", "coffee", "pool", "disappear", "delight", "awkward", "nature", "also", "tease"}, "37e2353bcd79f2d0f4034d7879f58f92"},
	{[]string{"surface", "search", "son", "rabbit", "fist", "forest", "journey", "swell", "single", "dust", "guy", "stir"}, "013bba5ec897daf45b9a8e4b9f4bf0ff"},
	{[]string{"breeze", "soak", "gently", "goose", "between", "salty", "space", "book", "duck", "learn", "passion", "mess"}, "a0f9d6bec2d60b8cbde15beaeba22846"},
	{[]string{"beauty", "maybe", "river", "cheek", "string", "ache", "pound", "slice", "bloom", "kingdom", "state", "chill"}, "55de2a9bb7a41d5abc89437bef4bd3d2"},
	{[]string{"bliss", "church", "empty", "eternity", "dig", "total", "themselves", "probably", "study", "turn", "crack", "certain"}, "b5570b012d9800372fe4962d9d9f07e8"},
	{[]string{"steel", "hollow", "mock", "brand", "black", "became", "appear", "weight", "strong", "deadly", "stun", "cruel"}, "2fbc465648d5e6ffbc5188f09d69bc4d"},
	{[]string{"inch", "listen", "eventually", "eat", "stay", "practice"}, "9a0d2375d6c8042e"},
	{[]string{"neck", "prayer", "special", "apologize", "group", "understand"}, "bdee9bb662018f7a"},
	{[]string{"relationship", "prayer", "through", "grandma", "wife", "maybe"}, "72e893aa9dae35de"},
	{[]string{"left", "put", "stray", "metal", "card", "little"}, "bf294d2352eef3ba"},
	{[]string{"ready", "loose", "surface", "question", "mountain", "men"}, "f9f90a12de8b234e"},
	{[]string{"ruin", "wrap", "dad", "good", "freeze", "add"}, "e1a1783ca8bfc1a7"},
	{[]string{"corner", "cold", "retreat", "dear", "arrow", "beyond"}, "c104a80767ced79d"},
	{[]string{"fly", "complete", "beam", "concrete", "pray", "straight"}, "57fb3d3916e9317c"},
	{[]string{"season", "idea", "act", "government", "reveal", "wipe"}, "dd859e5a28c3636a"},
	{[]string{"grown", "bathroom", "woman", "weave", "threw", "course"}, "7549bc02b3bdb844"},
	{[]string{"numb", "secret", "sick", "lord", "abuse", "heavy"}, "1261f990abcc967a"},
	{[]string{"reason", "goodbye", "prepare", "loser", "pass", "fantasy"}, "42fd6bb76b5e6e4c"},
	{[]string{"lot", "grief", "inhale", "health", "park", "false"}, "5a3518661c1ea6e8"},
	{[]string{"choice", "bit", "squeeze", "yard", "matter", "blush"}, "c56fa260dee14562"},
	{[]string{"rip", "scratch", "yard", "early", "another", "corner"}, "e8af7dd02be53144"},
	{[]string{"agony", "opposite", "endless", "space", "childhood", "give"}, "7c544c8650470f0c"},
	{[]string{"second", "surely", "suicide", "parent", "lust", "fresh"}, "00febc670261757a"},
	{[]string{"dirt", "hurry", "six", "order", "nature", "confuse"}, "68669e2cfb54ab99"},
	{[]string{"upon", "cruel", "king", "sure", "candle", "safe"}, "c99d650095a32529"},
	{[]string{"lovely", "danger", "shame", "dinner", "giggle", "depth"}, "867f691ba0c94a82"},
	{[]string{"boat", "teeth", "great"}, "8fdc42c3"},
	{[]string{"nightmare", "mine", "yell"}, "34e799a6"},
	{[]string{"meant", "petal", "fear"}, "2730f7e4"},
	{[]string{"wrap", "king", "frame"}, "5a221168"},
	{[]string{"desire", "visit", "release"}, "e91ea466"},
	{[]string{"north", "sink", "young"}, "8dc6e29f"},
	{[]string{"sanctuary", "truck", "happy"}, "7cc6ebc8"},
	{[]string{"path", "arrow", "whatever"}, "6af4b8de"},
	{[]string{"serious", "class", "fairy"}, "6ff5f783"},
	{[]string{"weak", "neighbor", "angry"}, "867efc04"},
	{[]string{"gotten", "torture", "sunset"}, "57ec90d8"},
	{[]string{"together", "hallway", "sat"}, "9ad7429b"},
	{[]string{"insane", "jeans", "unseen"}, "48840b03"},
	{[]string{"unlike", "flirt", "cool"}, "4c0089d8"},
	{[]string{"gas", "shiver", "season"}, "f3e6e284"},
	{[]string{"bid", "wish", "girlfriend"}, "bb8facbd"},
	{[]string{"week", "shade", "shout"}, "cf461369"},
	{[]string{"footstep", "desert", "piece"}, "7ac3eb75"},
	{[]string{"handle", "broken", "pierce"}, "79987fc8"},
	{[]string{"suffer", "flutter", "leaf"}, "8900715f"},
}

func TestToHex(t *testing.T) {
	for _, data := range testdata {
		r := FromWords(data.wordArray).ToHex()

		if r != data.hexString {
			t.Errorf("Expected: %s, Got: %s\n", data.hexString, r)
		}
	}
}

func TestToWords(t *testing.T) {
	for _, data := range testdata {
		r := strings.Join(FromHexstring(data.hexString).ToWords(), " ")

		if r != strings.Join(data.wordArray, " ") {
			t.Errorf("Expected: %s, Got: %s\n", strings.Join(data.wordArray, " "), r)
		}
	}

}
