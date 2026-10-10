package services

import "strings"

var englishToChineseTeam = map[string]string{
	// ===== 英超 Premier League =====
	"Manchester City": "曼城",
	"Manchester United": "曼联",
	"Liverpool": "利物浦",
	"Chelsea": "切尔西",
	"Arsenal": "阿森纳",
	"Tottenham": "热刺",
	"Newcastle": "纽卡斯尔",
	"Aston Villa": "阿斯顿维拉",
	"Brighton": "布莱顿",
	"West Ham": "西汉姆",
	"Bournemouth": "伯恩茅斯",
	"Crystal Palace": "水晶宫",
	"Wolverhampton": "狼队",
	"Fulham": "富勒姆",
	"Everton": "埃弗顿",
	"Brentford": "布伦特福德",
	"Nottingham Forest": "诺丁汉森林",
	"Luton": "卢顿",
	"Sheffield United": "谢菲尔德联",
	"Burnley": "伯恩利",

	// ===== 西甲 La Liga =====
	"Real Madrid": "皇家马德里",
	"Barcelona": "巴塞罗那",
	"Atletico Madrid": "马德里竞技",
	"Sevilla": "塞维利亚",
	"Real Sociedad": "皇家社会",
	"Real Betis": "皇家贝蒂斯",
	"Villarreal": "比利亚雷亚尔",
	"Athletic Club": "毕尔巴鄂",
	"Valencia": "瓦伦西亚",
	"Osasuna": "奥萨苏纳",
	"Girona": "赫罗纳",
	"Rayo Vallecano": "巴列卡诺",
	"Celta Vigo": "塞尔塔",
	"Mallorca": "马洛卡",
	"Getafe": "赫塔费",
	"Las Palmas": "拉斯帕尔马斯",
	"Alaves": "阿拉维斯",
	"Cadiz": "加的斯",
	"Granada CF": "格拉纳达",

	// ===== 德甲 Bundesliga =====
	"Bayern Munich": "拜仁慕尼黑",
	"Borussia Dortmund": "多特蒙德",
	"RB Leipzig": "莱比锡红牛",
	"Bayer Leverkusen": "勒沃库森",
	"Eintracht Frankfurt": "法兰克福",
	"VfL Wolfsburg": "沃尔夫斯堡",
	"SC Freiburg": "弗赖堡",
	"TSG Hoffenheim": "霍芬海姆",
	"Union Berlin": "柏林联合",
	"Borussia Monchengladbach": "门兴格拉德巴赫",
	"VfB Stuttgart": "斯图加特",
	"Mainz 05": "美因茨",
	"Werder Bremen": "云达不来梅",
	"Augsburg": "奥格斯堡",
	"Heidenheim": "海登海姆",
	"FC Koln": "科隆",
	"Hertha Berlin": "柏林赫塔",
	"Schalke 04": "沙尔克04",
	"Hamburger SV": "汉堡",

	// ===== 意甲 Serie A =====
	"Inter": "国际米兰",
	"AC Milan": "AC米兰",
	"Juventus": "尤文图斯",
	"Napoli": "那不勒斯",
	"Roma": "罗马",
	"Lazio": "拉齐奥",
	"Atalanta": "亚特兰大",
	"Fiorentina": "佛罗伦萨",
	"Bologna": "博洛尼亚",
	"Torino": "都灵",
	"Monza": "蒙扎",
	"Genoa": "热那亚",
	"Cagliari": "卡利亚里",
	"Lecce": "莱切",
	"Udinese": "乌迪内斯",
	"Sassuolo": "萨索洛",
	"Empoli": "恩波利",
	"Verona": "维罗纳",
	"Salernitana": "萨勒尼塔纳",
	"Frosinone": "弗罗西诺内",

	// ===== 法甲 Ligue 1 =====
	"Paris Saint-Germain": "巴黎圣日耳曼",
	"Marseille": "马赛",
	"Lyon": "里昂",
	"Monaco": "摩纳哥",
	"Lille": "里尔",
	"Rennes": "雷恩",
	"Nice": "尼斯",
	"Lens": "朗斯",
	"Montpellier": "蒙彼利埃",
	"Toulouse": "图卢兹",
	"Strasbourg": "斯特拉斯堡",
	"Nantes": "南特",
	"Reims": "兰斯",
	"Brest": "布雷斯特",
	"Le Havre": "勒阿弗尔",
	"Metz": "梅斯",
	"Clermont": "克莱蒙",
	"Lorient": "洛里昂",

	// ===== 世界杯/国家队 (FIFA) =====
	"Brazil": "巴西",
	"Argentina": "阿根廷",
	"France": "法国",
	"England": "英格兰",
	"Spain": "西班牙",
	"Germany": "德国",
	"Italy": "意大利",
	"Netherlands": "荷兰",
	"Portugal": "葡萄牙",
	"Belgium": "比利时",
	"Croatia": "克罗地亚",
	"Uruguay": "乌拉圭",
	"Mexico": "墨西哥",
	"USA": "美国",
	"Japan": "日本",
	"Korea Republic": "韩国",
	"South Korea": "韩国",
	"Australia": "澳大利亚",
	"Saudi Arabia": "沙特阿拉伯",
	"Iran": "伊朗",
	"Switzerland": "瑞士",
	"Denmark": "丹麦",
	"Sweden": "瑞典",
	"Poland": "波兰",
	"Serbia": "塞尔维亚",
	"Czech Republic": "捷克",
	"Turkey": "土耳其",
	"Colombia": "哥伦比亚",
	"Chile": "智利",
	"Ecuador": "厄瓜多尔",
	"Canada": "加拿大",
	"Morocco": "摩洛哥",
	"Senegal": "塞内加尔",
	"Nigeria": "尼日利亚",
	"Cameroon": "喀麦隆",
	"Ghana": "加纳",
	"Tunisia": "突尼斯",
	"Egypt": "埃及",
	"Qatar": "卡塔尔",

	// ===== 中超 =====
	"Beijing Guoan": "北京国安",
	"Shanghai Shenhua": "上海申花",
	"Shanghai Port": "上海海港",
	"Guangzhou": "广州队",
	"Shandong Taishan": "山东泰山",
	"Wuhan Three Towns": "武汉三镇",
	"Chengdu Rongcheng": "成都蓉城",
	"Zhejiang Professional": "浙江队",
	"Henan Songshan Longmen": "河南嵩山龙门",
	"Tianjin Jinmen Tiger": "天津津门虎",
}

// englishToChineseLeague api-football 联赛名 → sporttery 中文联赛名
// sporttery 的 leagueAllName 已经是中文(如"英超"),需要把 api-football 的英文名映射过去做辅助验证。
// 主流竞彩开售联赛均已覆盖,未匹配项留空(只靠日期+队名匹配)。
var englishToChineseLeague = map[string]string{
	"Premier League":         "英超",
	"La Liga":                "西甲",
	"Bundesliga":             "德甲",
	"Serie A":                "意甲",
	"Ligue 1":                "法甲",
	"UEFA Champions League":  "欧冠",
	"UEFA Europa League":     "欧联",
	"FIFA World Cup":         "世界杯",
	"UEFA European Championship": "欧洲杯",
	"Copa America":           "美洲杯",
	"AFC Champions League":   "亚冠",
	"Chinese Super League":   "中超",
	"Serie A Brasil":         "巴甲",
	"Copa Libertadores":      "解放者杯",
}

// translateTeamName 把 api-football 返回的英文俱乐部/国家队名翻译成 sporttery 用的中文名。
// 命中返回中文名,未命中返回 ""(调用方应跳过该条记录走降级)。
func translateTeamName(en string) string {
	if en == "" {
		return ""
	}
	if cn, ok := englishToChineseTeam[en]; ok {
		return cn
	}
	for k, v := range englishToChineseTeam {
		if strings.EqualFold(k, en) {
			return v
		}
	}
	return ""
}
