package libmozhi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// The fixtures keep the structure of the real answers (nov 2026), trimmed.

func TestJaDictWord(t *testing.T) {
	for text, want := range map[string]bool{
		" 食べる ": true, "たべる": true, "カタカナ": true, "一口食べる": true,
		"": false, "eat": false, "今日は天気がいい。": false, "食べる\n飲む": false,
		"とてもながいぶんしょうはじしょでしらべません": false,
	} {
		if _, got := JaDictWord(text); got != want {
			t.Errorf("JaDictWord(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestJishoParse(t *testing.T) {
	body := []byte(`{"meta":{"status":200},"data":[
	{"slug":"食べる","is_common":true,"jlpt":["jlpt-n5"],"japanese":[{"word":"食べる","reading":"たべる"}],
	 "senses":[{"english_definitions":["to eat"],"parts_of_speech":["Ichidan verb","Transitive verb"],"tags":[],"info":[]},
	           {"english_definitions":["to live on (e.g. a salary)","to live off"],"parts_of_speech":[],"tags":["Colloquial"],"info":[]},
	           {"english_definitions":["Taberu"],"parts_of_speech":["Wikipedia definition"],"tags":[],"info":[]}]},
	{"slug":"食べるラー油","is_common":false,"jlpt":[],"japanese":[{"word":"食べるラー油","reading":"たべるラーゆ"}],
	 "senses":[{"english_definitions":["chili oil"],"parts_of_speech":["Noun"],"tags":[],"info":[]}]}]}`)
	got, err := jishoParse(body, "食べる")
	if err != nil {
		t.Fatal(err)
	}
	want := []JaDictEntry{{Word: "食べる", Reading: "たべる", Tags: []string{"common", "JLPT N5"}, Senses: []JaDictSense{
		{POS: "Ichidan verb, Transitive verb", Glosses: []string{"to eat"}},
		{Glosses: []string{"to live on (e.g. a salary)", "to live off", "(Colloquial)"}},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	// Loose matches only (ぽぽぽぽぽ → 波 "Poland"): nothing.
	got, _ = jishoParse([]byte(`{"data":[{"japanese":[{"word":"波","reading":"ぽ"}],"senses":[{"english_definitions":["Poland"]}]}]}`), "ぽぽぽぽぽ")
	if len(got) != 0 {
		t.Errorf("loose match kept: %+v", got)
	}
	if _, err := jishoParse([]byte(`<html>`), "x"); err == nil {
		t.Error("an HTML answer must be an error")
	}
}

const jpdbDetails = `<html><body><div class="results details"><div id="result-0"><div class="result vocabulary">
<div>Conjugated as: past <a class="what-is-this" href="/conjugation/verb/た">?</a></div><div class="force-newline"></div>
<div class="vbox gap"><div class="hbox grow"><div class="hbox grow"><div class="vbox subsection-headword">
<div class="subsection-spelling with-furigana"><div class="primary-spelling"><div class="spelling"><div><ruby class="v">食<rt>た</rt>べ<rt></rt>る<rt></rt></ruby></div></div></div></div></div>
<div class="vbox grow gap"><div class="subsection-meanings"><h6 class="subsection-label">Meanings</h6><div class="subsection">
<div class="part-of-speech"><div>Verb (1-dan, transitive)</div></div>
<div class="description">1. to eat<div class="sense-example"><a class="icon-link example-audio" href="#"><i class="ti ti-volume"></i></a>
<div class="sense-example-text"><span class="jp">もう<span class="highlight"><ruby>食<rt>た</rt></ruby>べられない</span>わ！</span><span class="en">I can&#39;t eat anymore!</span></div></div></div>
<div class="description">2. to live on (e.g. a salary); to live off; to subsist on</div>
<div class="description">3. family<span class="note">after a surname</span></div>
</div></div></div></div>
<div class="tags xbox"><div class="tag tooltip">Top 300</div></div></div></div></div></div></body></html>`

func TestJpdbParse(t *testing.T) {
	got, err := jpdbParse([]byte(jpdbDetails), "食べた")
	if err != nil {
		t.Fatal(err)
	}
	want := []JaDictEntry{{Word: "食べる", Reading: "たべる", Note: "Conjugated as: past", Tags: []string{"Top 300"}, Senses: []JaDictSense{
		{POS: "Verb (1-dan, transitive)", Glosses: []string{"to eat"}, Examples: []JaDictExample{{Japanese: "もう食べられないわ！", Translation: "I can't eat anymore!"}}},
		{Glosses: []string{"to live on (e.g. a salary)", "to live off", "to subsist on"}},
		{Glosses: []string{"family after a surname"}},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	got, err = jpdbParse([]byte(`<div class="results search">No results found.</div>`), "zzqxw")
	if err != nil || len(got) != 0 {
		t.Errorf("no results: got %+v, %v", got, err)
	}
	// A search list keeps exact matches only: guesses for a made-up word go.
	guess := `<div class="results search"><div class="result vocabulary"><div>Conjugated as: stem</div><div class="primary-spelling"><ruby class="v">ぽい</ruby></div>
	<div class="subsection-meanings"><div class="subsection"><div class="description">1. -ish</div></div></div></div></div>`
	if got, _ = jpdbParse([]byte(guess), "ぽぽぽぽぽ"); len(got) != 0 {
		t.Errorf("guess kept: %+v", got)
	}
	if got, _ = jpdbParse([]byte(guess), "ぽい"); len(got) != 1 {
		t.Errorf("exact match dropped: %+v", got)
	}
	if _, err := jpdbParse([]byte(`<html><body>Maintenance</body></html>`), "x"); err == nil {
		t.Error("a page without results must be an error")
	}
}

func TestMojiParse(t *testing.T) {
	body := []byte(`{"result":{"result":{"word":[
	{"spell":"食べる","pron":"たべる","accent":"②","romaji":"taberu","excerpt":"[他动·一段] 吃，进食；生活；维持生计","tags":"N2#N5#考研","subdetailsIds":["81746","81747"]},
	{"spell":"食べ物","pron":"たべもの","accent":"③","excerpt":"[名] 食物","tags":"N5"}],
	"subdetails":[{"relaId":"81746","title":"吃，进食","lang":"zh-CN"},{"relaId":"81746","title":"飲食物をいただく。","lang":"ja"},
	              {"relaId":"81747","title":"生活；维持生计","lang":"zh-CN"},{"relaId":"81747","title":"生計を立てる。","lang":"ja"}]}},"code":200}`)
	got, err := mojiParse(body, "食べる")
	if err != nil {
		t.Fatal(err)
	}
	want := []JaDictEntry{{Word: "食べる", Reading: "たべる", Romaji: "taberu", Accent: "②", Tags: []string{"JLPT N2", "JLPT N5"}, Senses: []JaDictSense{
		{POS: "他动·一段", Glosses: []string{"吃，进食", "飲食物をいただく。"}},
		{Glosses: []string{"生活；维持生计", "生計を立てる。"}},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	got, _ = mojiParse([]byte(`{"result":{"result":{"word":[{"spell":"ぽぽ様","pron":"ぽぽさま","excerpt":"[名] 蚕"}],"subdetails":[]}}}`), "ぽぽぽぽぽ")
	if len(got) != 0 {
		t.Errorf("loose match kept: %+v", got)
	}
	got, _ = mojiParse([]byte(`{"result":{"result":{"word":[{"spell":"家","pron":"うち","accent":"⓪","excerpt":"[名] 自家；家庭"}],"subdetails":[]}}}`), "家")
	if len(got) != 1 || got[0].Senses[0].POS != "名" || !reflect.DeepEqual(got[0].Senses[0].Glosses, []string{"自家", "家庭"}) {
		t.Errorf("excerpt fallback: %+v", got)
	}
}

const weblioPage = `<html><body>
<div class="kijiWrp"><div class="kiji"><h2 class=midashigo title="home">home</h2><div class=Jtnhj><p>別表記：ホーム、家</p><p>家・家庭・自宅</p></div></div></div>
<div class="kijiWrp"><div class="kiji"><h2 class=midashigo title="食べる">食べる</h2><div class=Tssmj><p>作者某</p></div></div></div>
<div class="kijiWrp"> <div class="kiji"> <h2 class=midashigo title="食べる">た・べる【食べる】</h2> <div class=Sgkdj>
<p><!--AVOID_CROSSLINK-->読み方：たべる<!--/AVOID_CROSSLINK--></p>
<p><span class="hinshi">［動バ<a class=crosslink>下一</a>］</span>［文］た・ぶ</p>
<p><b>１</b> <a class=crosslink>食物</a>をかんで、のみこむ。「生(なま)で―・べる」</p>
<p><b>２</b> <a class=crosslink>暮らしを立てる</a>。生活する。</p>
<p><b>３</b></p>
<p>㋐「食う」「飲む」の謙譲語。いただく。</p>
<p>㋑「食う」「飲む」を丁寧にいう語。</p>
</div></div></div></body></html>`

// 家: a novel titled 家 first, then one デジタル大辞泉 block with several
// headwords: a redirect (け【化／…／家】), the definition (titled いえ), a
// novel's summary (いえ【家】).
const weblioIe = `<html><body>
<div class="kijiWrp"><div class="kiji"><h2 class=midashigo title="家">家</h2><div class=Tssmj>作者 アイプ・ロシディ</div></div></div>
<div class="kijiWrp"><div class="kiji">
<h2 class=midashigo title="内">うち【内】</h2><div class=Sgkdj><p>読み方：うち</p><p><b>１</b> ある一定の区域。</p></div>
<h2 class=midashigo title="化">け【化／仮／ ▽花／家／華】</h2><div class=Sgkdj><p>読み方：け</p><p>〈化〉⇒か 〈家〉⇒か</p></div>
<h2 class=midashigo title="いえ">いえ〔いへ〕【家】</h2><div class=Sgkdj><p>読み方：いえ</p><p><b>１</b> 人の住むための建物。すまい。</p><p><b>２</b> 家族。</p></div>
<h2 class=midashigo title="家">いえ〔いへ〕【家】</h2><div class=Sgkdj><p>読み方：いえ</p><p>島崎藤村の小説。</p></div>
</div></div>
<div class="kijiWrp"><div class="kiji"><h2 class=midashigo title="家">「家」の例文・使い方・用例・文例</h2><div class=Wnryj>家族を捨てる</div></div></div>
</body></html>`

const weblioTabeta = `<html><body><div class="kijiWrp"><div class="kiji"><h2 class=midashigo title="食べた">食べた</h2><div class=Nhgkt>
<div>読み方：たべた</div><div>バ行下一段活用の動詞「<a>食べる</a>」の連用形である「食べ」に、過去・完了の助動詞「た」が付いた形。</div>
<table><tr><td>終止形</td><td>食べる</td></tr></table></div></div></div></body></html>`

func TestWeblioParse(t *testing.T) {
	got, err := weblioParse([]byte(weblioPage), "食べる")
	if err != nil {
		t.Fatal(err)
	}
	want := []JaDictEntry{{Word: "た・べる【食べる】", Reading: "たべる", Note: "［動バ下一］［文］た・ぶ", Senses: []JaDictSense{
		{Glosses: []string{"１ 食物をかんで、のみこむ。「生(なま)で―・べる」"}},
		{Glosses: []string{"２ 暮らしを立てる。生活する。"}},
		{Glosses: []string{"３ ㋐「食う」「飲む」の謙譲語。いただく。", "㋑「食う」「飲む」を丁寧にいう語。"}},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	got, _ = weblioParse([]byte(weblioIe), "家")
	want = []JaDictEntry{{Word: "いえ〔いへ〕【家】", Reading: "いえ", Senses: []JaDictSense{
		{Glosses: []string{"１ 人の住むための建物。すまい。"}}, {Glosses: []string{"２ 家族。"}},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("家: got %+v\nwant %+v", got, want)
	}
	got, _ = weblioParse([]byte(weblioTabeta), "食べた")
	if len(got) != 1 || got[0].Reading != "たべた" || !strings.HasPrefix(got[0].Senses[0].Glosses[0], "バ行下一段活用の動詞「食べる」の連用形") {
		t.Errorf("食べた: %+v", got)
	}
	// 実用日本語表現辞典: no numbers, its paragraphs.
	got, _ = weblioParse([]byte(weblioPage), "home")
	if len(got) != 1 || got[0].Senses[0].Glosses[0] != "別表記：ホーム、家 家・家庭・自宅" {
		t.Errorf("home: %+v", got)
	}
	got, err = weblioParse([]byte(weblioPage), "飲む")
	if err != nil || len(got) != 0 {
		t.Errorf("no exact entry: got %+v, %v", got, err)
	}
	if _, err := weblioParse([]byte(`<html><body>error</body></html>`), "x"); err == nil {
		t.Error("a page without entries must be an error")
	}
}

func TestRubyWordReading(t *testing.T) {
	for _, c := range []struct{ html, word, reading string }{
		{`<div><ruby class="v">食<rt>た</rt>べ<rt></rt>る<rt></rt></ruby></div>`, "食べる", "たべる"},
		{`<div><ruby class="v">家<rt>いえ</rt></ruby></div>`, "家", "いえ"},
		{`<div><ruby>今日<rt>きょう</rt></ruby><ruby>は</ruby></div>`, "今日は", "きょうは"},
		{`<div><ruby>たべ</ruby><ruby>る</ruby></div>`, "たべる", ""},
	} {
		doc := mustDoc(t, c.html)
		w, r := rubyWordReading(doc.Find("ruby"))
		if w != c.word || r != c.reading {
			t.Errorf("%s: got %q %q, want %q %q", c.html, w, r, c.word, c.reading)
		}
	}
}

func mustDoc(t *testing.T, s string) *goquery.Document {
	t.Helper()
	d, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestJaDictLookupEmptyEntries(t *testing.T) {
	jaDictSources["test"] = jaDictSource{"Test", "en", func(string) string { return "" }, func(string) ([]JaDictEntry, error) { return nil, nil }}
	defer delete(jaDictSources, "test")
	out, err := JaDictLookup("test", "ぽぽぽぽぽ")
	if err != nil || out.Entries == nil || len(out.Entries) != 0 {
		t.Errorf("got %+v, %v: want empty, non-nil entries", out, err)
	}
}

func TestJaDictJoin(t *testing.T) {
	for _, c := range []struct {
		in   []string
		want string
	}{
		{[]string{"house; family", "after a surname"}, "house; family after a surname"},
		{[]string{"to live on ", "(e.g. a salary)"}, "to live on (e.g. a salary)"},
		{[]string{"to eat", "; to consume"}, "to eat; to consume"},
		{[]string{"もう", "食", "べられない"}, "もう食べられない"},
	} {
		if got := jaDictJoin(c.in); got != c.want {
			t.Errorf("jaDictJoin(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
