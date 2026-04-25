# 彩票助手功能增强 Spec

## Why
当前彩票助手已有基础的购买记录、开奖管理和中奖查询功能，但缺少自动化的数据获取、复式投注支持、奖金计算准确性、追加/倍数/多期投注能力，以及分页和开发者支持功能。本次增强旨在提升用户体验，减少手动操作，并修复已知问题。

## What Changes

### 1. 定时自动抓取开奖号码
- **新增**: 后端定时任务，定期自动从惠鸟彩票API抓取最新开奖结果
- **新增**: 支持可配置的抓取间隔（默认每天一次）
- **新增**: 支持手动触发立即抓取
- **影响**: `backend/main.go`, `backend/services/draw_service.go`, 新增 `backend/services/scheduler.go`

### 2. 复式号码支持
- **新增**: 购买记录支持复式投注（多选号码组合）
- **新增**: 中奖计算支持复式投注的奖级判定
- **新增**: 前端号码输入组件支持复式模式
- **影响**: `backend/rules/calculator.go`, `backend/models/models.go`, `frontend/src/components/NumberInput.vue`

### 3. 修复双色球六等奖奖金
- **修复**: 双色球六等奖（1蓝球命中）奖金从错误的15元修正为5元
- **影响**: `backend/rules/calculator.go`

### 4. 追加和倍数支持
- **新增**: 购买记录增加 `multiple`（倍数）和 `append`（追加）字段
- **新增**: 中奖计算根据倍数和追加计算实际奖金
- **新增**: 前端购买表单支持倍数和追加选项
- **影响**: `backend/models/models.go`, `backend/rules/calculator.go`, `backend/handlers/purchase_handler.go`, `frontend/src/views/PurchaseView.vue`

### 5. 多期投注支持
- **新增**: 购买记录增加 `periods`（期数）字段，支持一次录入多期
- **新增**: 录入时自动拆分为多条记录，每期自动递增期号
- **新增**: 前端支持选择期数（1-10期）
- **影响**: `backend/models/models.go`, `backend/handlers/purchase_handler.go`, `frontend/src/views/PurchaseView.vue`

### 6. 分页支持
- **新增**: 所有列表页面（购买记录、开奖管理、中奖记录）支持分页显示
- **新增**: 后端API统一返回分页数据（data, total, page, size）
- **新增**: 前端分页组件
- **影响**: `frontend/src/views/PurchaseView.vue`, `frontend/src/views/DrawView.vue`, `frontend/src/views/WinningsView.vue`, `frontend/src/api/index.ts`

### 7. 开发者支持弹窗
- **新增**: 页面增加【支持】按钮，点击弹出支持弹窗
- **新增**: 弹窗显示搞怪文案、两个二维码、两个跳转链接
- **影响**: `frontend/src/components/SupportModal.vue`, `frontend/src/App.vue` 或 `frontend/src/components/NavBar.vue`

## Impact
- 受影响模块：后端定时任务、中奖计算引擎、数据模型、前端所有列表页、号码输入组件
- 数据库变更：`purchase_records` 表新增 `multiple`, `append`, `periods` 字段
- 兼容性：向后兼容，现有单式投注不受影响

## ADDED Requirements

### Requirement: 定时自动抓取开奖号码
The system SHALL provide a scheduled task to automatically fetch draw results from external APIs.

#### Scenario: 定时抓取
- **GIVEN** 应用已启动
- **WHEN** 到达配置的抓取时间
- **THEN** 自动抓取所有支持类型的最新开奖结果并保存

#### Scenario: 手动触发
- **GIVEN** 管理员点击"立即抓取"按钮
- **WHEN** 按钮被点击
- **THEN** 立即执行抓取并返回结果

### Requirement: 复式投注支持
The system SHALL support multiple number combinations (复式) for purchase records.

#### Scenario: 复式投注录入
- **GIVEN** 用户选择复式模式
- **WHEN** 用户选择超过单式要求的号码数量
- **THEN** 系统记录为复式投注，计算时生成所有组合

#### Scenario: 复式中奖计算
- **GIVEN** 一条复式投注记录
- **WHEN** 开奖结果匹配时
- **THEN** 计算所有组合的中奖情况，返回最高奖级

### Requirement: 追加和倍数支持
The system SHALL support multiple bets and append options.

#### Scenario: 倍数投注
- **GIVEN** 用户选择倍数为3
- **WHEN** 中奖时
- **THEN** 奖金 = 基础奖金 × 3

#### Scenario: 大乐透追加
- **GIVEN** 用户购买大乐透并选择追加
- **WHEN** 中得一至三等奖
- **THEN** 奖金 = 基础奖金 × 1.6（追加后）

### Requirement: 多期投注支持
The system SHALL support purchasing multiple consecutive periods.

#### Scenario: 10期投注
- **GIVEN** 用户选择期数10，期号2024001
- **WHEN** 提交购买
- **THEN** 系统自动创建10条记录：2024001~2024010

### Requirement: 分页支持
The system SHALL paginate all list views.

#### Scenario: 分页显示
- **GIVEN** 列表有100条数据
- **WHEN** 用户查看列表
- **THEN** 每页显示20条，显示分页控件

### Requirement: 开发者支持弹窗
The system SHALL display a support modal with developer information.

#### Scenario: 打开支持弹窗
- **GIVEN** 用户点击【支持】按钮
- **WHEN** 按钮被点击
- **THEN** 弹出模态框，显示文案、二维码、跳转链接

## MODIFIED Requirements

### Requirement: 双色球奖金计算
**修改**: 六等奖奖金从15元修正为5元
**原因**: 双色球官方规则六等奖为5元

## REMOVED Requirements
无
