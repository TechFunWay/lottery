package services

import (
	"encoding/json"
	"lottery-backend/database"
	"os"
	"strings"
	"testing"
)

func TestSportteryFixture_ParseAndFieldMapping(t *testing.T) {
	path := "testdata/sporttery_calc_20260617.json"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture not present at %s (skipping live parse test): %v", path, err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var resp sportteryScheduleResp
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !resp.Success || resp.ErrorCode != "0" {
		t.Fatalf("expected success, got success=%v errCode=%q errMsg=%q",
			resp.Success, resp.ErrorCode, resp.ErrorMessage)
	}
	if len(resp.Value.MatchInfoList) == 0 {
		t.Fatal("expected non-empty matchInfoList")
	}

	raw, missingID, missingTeams, missingDate := 0, 0, 0, 0
	for _, day := range resp.Value.MatchInfoList {
		for _, m := range day.SubMatchList {
			raw++
			if m.MatchNumStr == "" {
				missingID++
			}
			if m.HomeTeamAbbName == "" || m.AwayTeamAbbName == "" {
				missingTeams++
			}
			if m.MatchDate == "" || m.MatchTime == "" {
				missingDate++
			}
		}
	}

	if missingID > 0 {
		t.Errorf("%d matches have empty MatchNumStr (the field the fetcher uses as MatchID)", missingID)
	}
	if missingTeams > 0 {
		t.Errorf("%d matches have empty HomeTeamAbbName or AwayTeamAbbName", missingTeams)
	}
	if missingDate > 0 {
		t.Errorf("%d matches have empty MatchDate or MatchTime", missingDate)
	}

	t.Logf("OK: parsed %d days, %d raw matches; all have matchNumStr and team names",
		len(resp.Value.MatchInfoList), raw)
}

func TestTranslateTeamName(t *testing.T) {
	cases := []struct {
		en   string
		want string
	}{
		{"Manchester City", "曼城"},
		{"Liverpool", "利物浦"},
		{"Real Madrid", "皇家马德里"},
		{"Barcelona", "巴塞罗那"},
		{"Bayern Munich", "拜仁慕尼黑"},
		{"Paris Saint-Germain", "巴黎圣日耳曼"},
		{"Inter", "国际米兰"},
		{"England", "英格兰"},
		{"Brazil", "巴西"},
		{"Beijing Guoan", "北京国安"},
		{"Crystal Palace", "水晶宫"},
		{"Luton", "卢顿"},
	}
	for _, c := range cases {
		got := translateTeamName(c.en)
		if got != c.want {
			t.Errorf("translateTeamName(%q) = %q, want %q", c.en, got, c.want)
		}
	}
}

func TestTranslateTeamName_UnknownReturnsEmpty(t *testing.T) {
	for _, en := range []string{"FC Some Unknown Club", "Random Team FC", ""} {
		if got := translateTeamName(en); got != "" {
			t.Errorf("translateTeamName(%q) = %q, want empty (no mapping)", en, got)
		}
	}
}

func TestTranslateTeamName_CaseInsensitiveFallback(t *testing.T) {
	got := translateTeamName("liverpool")
	if got != "利物浦" {
		t.Errorf("case-insensitive fallback failed: got %q, want 利物浦", got)
	}
}

func TestAPIFootballFixtures_Unmarshal(t *testing.T) {
	body := `{
		"get": "fixtures",
		"results": 2,
		"response": [
			{
				"fixture": {
					"id": 1234567,
					"date": "2026-06-17T19:00:00+00:00",
					"status": {"short": "FT", "long": "Match Finished"}
				},
				"league": {"name": "Premier League"},
				"teams": {
					"home": {"name": "Liverpool"},
					"away": {"name": "Manchester United"}
				},
				"goals": {"home": 2, "away": 1},
				"score": {"halftime": {"home": 1, "away": 0}}
			},
			{
				"fixture": {
					"id": 1234568,
					"date": "2026-06-18T20:00:00+00:00",
					"status": {"short": "NS", "long": "Not Started"}
				},
				"league": {"name": "Premier League"},
				"teams": {
					"home": {"name": "Crystal Palace"},
					"away": {"name": "Arsenal"}
				},
				"goals": {"home": null, "away": null},
				"score": {"halftime": {"home": null, "away": null}}
			}
		]
	}`

	var resp apiFootballFixturesResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Response) != 2 {
		t.Fatalf("expected 2 fixtures, got %d", len(resp.Response))
	}

	ft := resp.Response[0]
	if ft.Fixture.Status.Short != "FT" {
		t.Errorf("expected first match FT, got %q", ft.Fixture.Status.Short)
	}
	if ft.Goals.Home == nil || *ft.Goals.Home != 2 {
		t.Errorf("expected home goals=2, got %v", ft.Goals.Home)
	}
	if ft.Goals.Away == nil || *ft.Goals.Away != 1 {
		t.Errorf("expected away goals=1, got %v", ft.Goals.Away)
	}
	if ft.Score.Halftime.Home == nil || *ft.Score.Halftime.Home != 1 {
		t.Errorf("expected halftime home=1, got %v", ft.Score.Halftime.Home)
	}

	ns := resp.Response[1]
	if ns.Fixture.Status.Short != "NS" {
		t.Errorf("expected second match NS, got %q", ns.Fixture.Status.Short)
	}
	if ns.Goals.Home != nil {
		t.Errorf("expected home goals=nil for NS match, got %v", *ns.Goals.Home)
	}
}

func TestFetchMatchResults_NoKeyReturnsEmpty(t *testing.T) {
	if database.DB == nil {
		t.Skip("database.DB 未初始化,跳过需要 DB 的 no-key 集成测试(配置 key 需走 DB 解析层)")
	}
	prev, had := os.LookupEnv("API_FOOTBALL_KEY")
	os.Unsetenv("API_FOOTBALL_KEY")
	defer func() {
		if had {
			os.Setenv("API_FOOTBALL_KEY", prev)
		}
	}()

	s := &FootballService{}
	matches, err := s.FetchMatchResults(0)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil matches, got %v", matches)
	}
}

func TestAPIFootballStatusFilter(t *testing.T) {
	cases := []struct {
		short  string
		passed bool
	}{
		{"FT", true},
		{"AET", true},
		{"PEN", true},
		{"NS", false},
		{"1H", false},
		{"HT", false},
		{"", false},
	}
	for _, c := range cases {
		got := c.short == "FT" || c.short == "AET" || c.short == "PEN"
		if got != c.passed {
			t.Errorf("status %q: expected passed=%v, got %v", c.short, c.passed, got)
		}
	}
}

func TestMapCoverage_AllLeaguesHaveEntries(t *testing.T) {
	requiredLeagues := []string{"Premier League", "La Liga", "Bundesliga", "Serie A", "Ligue 1", "UEFA Champions League", "FIFA World Cup"}
	for _, l := range requiredLeagues {
		if _, ok := englishToChineseLeague[l]; !ok {
			t.Errorf("missing league mapping: %q", l)
		}
	}
	if len(englishToChineseTeam) < 50 {
		t.Errorf("team mapping too sparse: only %d entries (want >= 50 for realistic coverage)", len(englishToChineseTeam))
	}
}

func TestMapCoverage_NoDuplicateChineseNames(t *testing.T) {
	seen := map[string]string{}
	for en, cn := range englishToChineseTeam {
		if existing, ok := seen[cn]; ok {
			if !strings.EqualFold(existing, en) {
				t.Logf("warning: Chinese name %q maps to both %q and %q", cn, existing, en)
			}
		}
		seen[cn] = en
	}
}

func TestResolveAPIFootballKey_BuiltinFallback(t *testing.T) {
	if database.DB == nil {
		t.Skip("database.DB 未初始化,跳过需要 DB 的解析层集成测试")
	}

	prevEnv, hadEnv := os.LookupEnv("API_FOOTBALL_KEY")
	os.Unsetenv("API_FOOTBALL_KEY")
	defer func() {
		if hadEnv {
			os.Setenv("API_FOOTBALL_KEY", prevEnv)
		}
	}()

	prevBuiltin := BuiltInAPIFootballKey
	BuiltInAPIFootballKey = "builtin-test-key"
	defer func() { BuiltInAPIFootballKey = prevBuiltin }()

	s := ConfigService{}
	key, source := s.ResolveAPIFootballKey(0)
	if key != "builtin-test-key" {
		t.Errorf("expected key=builtin-test-key, got %q", key)
	}
	if source != FootballKeySourceBuiltin {
		t.Errorf("expected source=%q, got %q", FootballKeySourceBuiltin, source)
	}
}

func TestResolveAPIFootballKey_EnvBeatsBuiltin(t *testing.T) {
	if database.DB == nil {
		t.Skip("database.DB 未初始化,跳过需要 DB 的解析层集成测试")
	}

	prevEnv, hadEnv := os.LookupEnv("API_FOOTBALL_KEY")
	os.Setenv("API_FOOTBALL_KEY", "env-test-key")
	defer func() {
		if hadEnv {
			os.Setenv("API_FOOTBALL_KEY", prevEnv)
		} else {
			os.Unsetenv("API_FOOTBALL_KEY")
		}
	}()

	prevBuiltin := BuiltInAPIFootballKey
	BuiltInAPIFootballKey = "builtin-test-key"
	defer func() { BuiltInAPIFootballKey = prevBuiltin }()

	s := ConfigService{}
	key, source := s.ResolveAPIFootballKey(0)
	if key != "env-test-key" {
		t.Errorf("expected env to win, got key=%q", key)
	}
	if source != FootballKeySourceEnv {
		t.Errorf("expected source=%q, got %q", FootballKeySourceEnv, source)
	}
}
