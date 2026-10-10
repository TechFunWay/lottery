package services

// BuiltInAPIFootballKey 编译时通过 -ldflags 注入的 API-Football Key,
// 作为「开箱即用」的最后兜底,让 NAS 单用户场景下完全零配置。
//
// 发布命令:
//
//	API_FOOTBALL_KEY=xxx make release
//	或
//	go build -ldflags "-X 'lottery-backend/services.BuiltInAPIFootballKey=xxx'" ...
//
// 解析优先级(详见 ConfigService.ResolveAPIFootballKey):
//  1. 用户自配(per-user,SystemConfig)
//  2. 管理员全局(SystemConfig)
//  3. 环境变量 API_FOOTBALL_KEY(开发者本地覆盖)
//  4. 本内置 Key(发布时注入,默认兜底)
//
// 留空时行为不变,降级到原有手动配置流程。
var BuiltInAPIFootballKey = ""
