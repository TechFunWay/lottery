# 数据库升级功能说明

## 概述

本系统实现了自动化的数据库版本升级功能，当用户从旧版本升级到新版本时，会自动执行相应的数据库变更脚本。

## 设计特点

### 1. 升级表

表名：`system_upgrade`

字段说明：
- `version`: 版本号（如 v1.0.0, v1.1.0）
- `name`: 升级名称（如"初始化数据库"、"添加用户头像字段"）
- `status`: 升级状态（0-待执行, 1-成功, 2-失败）
- `start_time`: 开始时间
- `end_time`: 结束时间
- `remark`: 备注/错误信息

### 2. 升级日志

升级过程会自动记录到日志文件中，日志文件位于：

```
<data_dir>/logs/upgrade-YYYY-MM-DD.log
```

日志文件按日期分割，每天一个文件。日志包含以下信息：
- 升级开始和结束时间
- 当前数据库版本
- 待升级的版本列表
- 每个版本的升级开始、成功/失败状态
- 升级耗时

示例日志：
```
[2026-03-22 00:21:34] ========================================
[2026-03-22 00:21:34] 开始检查数据库升级
[2026-03-22 00:21:34] ========================================
[2026-03-22 00:21:34] 当前数据库版本: v1.0.0
[2026-03-22 00:21:34] 发现 1 个待升级版本
[2026-03-22 00:21:34]   - v1.1.0
[2026-03-22 00:21:34] 🔨 开始升级到 v1.1.0: 添加用户头像字段
[2026-03-22 00:21:34] ✅ 升级 v1.1.0 成功，耗时: 4.405834ms
[2026-03-22 00:21:34] 🎉 所有升级完成，当前版本: v1.1.0
[2026-03-22 00:21:34] ========================================
[2026-03-22 00:21:34] 升级检查结束
[2026-03-22 00:21:34] ========================================
```

### 3. 版本号格式

版本号采用三段式：`v主版本.次版本.修订号`

例如：
- `v1.0.0` - 主版本1.0，次版本0，修订号0
- `v1.1.0` - 主版本1.0，次版本1，修订号0
- `v1.2.1` - 主版本1.0，次版本2，修订号1

版本号比较规则：
- 按从左到右逐段比较数字大小
- v1.1.0 < v1.2.0
- v1.9.0 < v1.10.0
- v1.2.1 < v1.2.2

### 4. 升级流程

1. 程序启动时自动检查数据库中的升级记录
2. 查询当前已执行成功的最新版本
3. 找出所有大于当前版本的升级脚本
4. 按版本号从小到大依次执行升级
5. 每个版本执行成功后，在`system_upgrade`表中记录
6. 如果某个版本执行失败，停止升级并记录错误信息

### 5. 幂等性保证

每个升级函数都是幂等的，可以重复执行而不影响数据：

- **检查表是否存在**：使用`db.Migrator().HasTable()`
- **检查列是否存在**：使用`db.Migrator().HasColumn()`
- **检查数据是否存在**：使用`db.Where().First()`

示例：
```go
func upgrade_v1_1_0(db *gorm.DB) error {
    // 方法1：检查表是否存在
    if !db.Migrator().HasTable("new_table") {
        db.Exec("CREATE TABLE new_table ...")
    }
    
    // 方法2：检查列是否存在
    if !db.Migrator().HasColumn("users", "avatar") {
        db.Exec("ALTER TABLE users ADD COLUMN avatar ...")
    }
    
    // 方法3：检查数据是否存在
    var count int64
    db.Table("system_configs").Where("key = ?", "new_config").Count(&count)
    if count == 0 {
        db.Exec("INSERT INTO system_configs ...")
    }
    
    return nil
}
```

## 使用方法

### 添加新版本升级

1. 在 `backend/migrations/upgrade.go` 中添加新的升级函数

```go
// upgrade_v1_1_0 v1.1.0升级
func upgrade_v1_1_0(db *gorm.DB) error {
    // 数据库变更逻辑
    // ...
    return nil
}
```

2. 在 `UpgradeScripts` 映射中注册新版本

```go
var UpgradeScripts = map[string]UpgradeScript{
    "v1.0.0": {
        Name: "初始化数据库",
        Func: upgrade_v1_0_0,
    },
    "v1.1.0": {  // 新增
        Name: "添加用户头像字段",
        Func: upgrade_v1_1_0,
    },
}
```

3. 重新编译并部署新版本

### 升级示例场景

#### 场景1：用户从 v1.0.0 升级到 v1.1.0

1. 数据库中已有 v1.0.0 的升级记录（status=1）
2. 程序检测到 v1.1.0 > v1.0.0
3. 执行 v1.1.0 的升级函数
4. 插入 v1.1.0 的升级记录（status=1）

#### 场景2：用户从 v1.0.0 升级到 v1.3.0（跨越多个版本）

1. 数据库中已有 v1.0.0 的升级记录
2. 程序检测到 v1.1.0, v1.2.0, v1.3.0 都大于 v1.0.0
3. 按顺序执行：v1.1.0 → v1.2.0 → v1.3.0
4. 插入三个版本的升级记录

#### 场景3：重复启动程序

1. 数据库中已有 v1.3.0 的升级记录（status=1）
2. 程序检测到当前版本 v1.3.0 是最新的
3. 输出"✅ 当前已是最新版本: v1.3.0"
4. 跳过所有升级，继续启动

## 升级函数最佳实践

### 1. 表结构变更

```go
func upgrade_v1_1_0(db *gorm.DB) error {
    // 使用 AutoMigrate 自动创建表和添加列
    return db.AutoMigrate(&NewModel{})
}
```

### 2. 添加新配置

```go
func upgrade_v1_1_0(db *gorm.DB) error {
    var config models.SystemConfig
    err := db.Where("key = ?", "new_key").First(&config).Error
    
    if err == gorm.ErrRecordNotFound {
        config = models.SystemConfig{
            Key:    "new_key",
            Value:  "default_value",
            Remark: "配置说明",
        }
        return db.Create(&config).Error
    }
    
    return nil
}
```

### 3. 数据迁移

```go
func upgrade_v1_1_0(db *gorm.DB) error {
    // 迁移旧数据到新结构
    var oldRecords []OldModel
    db.Find(&oldRecords)
    
    for _, old := range oldRecords {
        newRecord := NewModel{
            Field1: old.OldField1,
            Field2: old.OldField2,
        }
        db.Create(&newRecord)
    }
    
    return nil
}
```

## 注意事项

1. **版本号严格格式**：必须以 `v` 开头，三段式数字（如 v1.0.0）
2. **幂等性**：每个升级函数必须可以安全地重复执行
3. **错误处理**：升级失败时，程序会停止并记录错误信息
4. **数据库备份**：建议在升级前备份数据库
5. **测试验证**：在测试环境充分测试后再部署到生产环境

## 文件结构

```
backend/
  migrations/
    upgrade.go          - 升级函数集合
  models/
    system_upgrade.go   - 升级记录模型
  services/
    upgrade_service.go  - 升级服务
  main.go              - 启动时调用升级服务
```

## API接口（可选）

如需提供升级管理的API接口，可以添加以下端点：

- `GET /api/upgrade/version` - 查询当前数据库版本
- `GET /api/upgrade/status` - 查询升级状态
- `POST /api/upgrade/run` - 手动触发升级
- `GET /api/upgrade/history` - 查询升级历史
