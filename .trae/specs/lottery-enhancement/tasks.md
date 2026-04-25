# Tasks

## Task 1: 修复双色球六等奖奖金
修复双色球六等奖（1蓝球命中）奖金从15元修正为5元。
- [ ] SubTask 1.1: 修改 `backend/rules/calculator.go` 中双色球六等奖奖金为5元
- [ ] SubTask 1.2: 编译验证

## Task 2: 定时自动抓取开奖号码
- [ ] SubTask 2.1: 创建 `backend/services/scheduler.go` 定时任务服务
- [ ] SubTask 2.2: 在 `backend/main.go` 中启动定时任务
- [ ] SubTask 2.3: 支持可配置抓取间隔（默认每天一次）
- [ ] SubTask 2.4: 添加手动触发API
- [ ] SubTask 2.5: 编译验证

## Task 3: 复式号码支持
- [ ] SubTask 3.1: 修改 `backend/models/models.go` 支持复式号码存储
- [ ] SubTask 3.2: 修改 `backend/rules/calculator.go` 支持复式中奖计算
- [ ] SubTask 3.3: 修改 `frontend/src/components/NumberInput.vue` 支持复式模式
- [ ] SubTask 3.4: 修改 `frontend/src/views/PurchaseView.vue` 支持复式录入
- [ ] SubTask 3.5: 编译验证

## Task 4: 追加和倍数支持
- [ ] SubTask 4.1: 修改 `backend/models/models.go` 新增 `multiple` 和 `append` 字段
- [ ] SubTask 4.2: 修改 `backend/rules/calculator.go` 奖金计算考虑倍数和追加
- [ ] SubTask 4.3: 修改 `backend/handlers/purchase_handler.go` 处理倍数和追加
- [ ] SubTask 4.4: 修改 `frontend/src/views/PurchaseView.vue` 支持倍数和追加选项
- [ ] SubTask 4.5: 数据库迁移脚本更新
- [ ] SubTask 4.6: 编译验证

## Task 5: 多期投注支持
- [ ] SubTask 5.1: 修改 `backend/models/models.go` 新增 `periods` 字段
- [ ] SubTask 5.2: 修改 `backend/handlers/purchase_handler.go` 多期自动拆分
- [ ] SubTask 5.3: 修改 `frontend/src/views/PurchaseView.vue` 支持期数选择
- [ ] SubTask 5.4: 数据库迁移脚本更新
- [ ] SubTask 5.5: 编译验证

## Task 6: 分页支持
- [ ] SubTask 6.1: 修改后端API返回分页数据（购买记录、开奖管理、中奖记录）
- [ ] SubTask 6.2: 创建 `frontend/src/components/Pagination.vue` 分页组件
- [ ] SubTask 6.3: 修改 `frontend/src/views/PurchaseView.vue` 支持分页
- [ ] SubTask 6.4: 修改 `frontend/src/views/DrawView.vue` 支持分页
- [ ] SubTask 6.5: 修改 `frontend/src/views/WinningsView.vue` 支持分页
- [ ] SubTask 6.6: 编译验证

## Task 7: 开发者支持弹窗
- [ ] SubTask 7.1: 创建 `frontend/src/components/SupportModal.vue` 支持弹窗组件
- [ ] SubTask 7.2: 在 `frontend/src/App.vue` 或导航栏添加【支持】按钮
- [ ] SubTask 7.3: 弹窗内容：搞怪文案 + 两个二维码 + 两个跳转链接
- [ ] SubTask 7.4: 编译验证

# Task Dependencies
- Task 3 依赖 Task 1（奖金计算基础）
- Task 4 依赖 Task 1（奖金计算基础）
- Task 5 依赖 Task 4（倍数/追加与多期结合）
- Task 6 与其他任务无依赖，可并行
- Task 7 与其他任务无依赖，可并行
