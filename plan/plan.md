# 执行计划

## 状态：已完成

## 进度：5/5 任务完成

## 任务列表

### 任务 1：修复配置向导 API 保存
- 状态：completed
- 笔记：
  - 修改了 LoginManager，添加 customAPI 字段
  - 修改了 Login 函数，优先使用 customAPI
  - 修改了 New 和 UpdateConfig 函数，传入 CustomLoginAPI
  - 编译测试通过
- 依赖：无
- 预估时间：15 分钟
- 复杂度：中等
- 验收标准：
  - [ ] 配置向导解析 cURL 后，自定义 API 正确保存到 cfg.CustomLoginAPI
  - [ ] 登录时优先使用 CustomLoginAPI（如果存在）
  - [ ] 重启程序后自定义 API 仍然生效
- 笔记：

### 任务 2：cURL 双格式解析
- 状态：completed
- 笔记：
  - 修改了 parseCurlCommand 函数
  - 将单引号替换为双引号，统一处理
  - 改进了 URL 提取正则表达式
  - 编译测试通过
- 依赖：无
- 预估时间：15 分钟
- 复杂度：中等
- 验收标准：
  - [ ] 支持解析 bash 格式：curl 'http://xxx' -H 'xxx'
  - [ ] 支持解析 cmd 格式：curl "http://xxx" -H "xxx"
  - [ ] 单引号和双引号都能正确处理
  - [ ] 解析失败时给出明确错误提示
- 笔记：

### 任务 3：更新默认登录 API
- 状态：completed
- 笔记：
  - 添加了 getLocalIP 和 getLocalMAC 函数
  - 自动获取本机 IP 和 MAC 地址
  - 默认 API 格式已更新
  - 编译测试通过
- 依赖：无
- 预估时间：10 分钟
- 复杂度：简单
- 验收标准：
  - [ ] 默认登录 API 使用新格式（包含 wlan_user_ip 等参数）
  - [ ] wlan_user_ip 从当前连接自动获取
  - [ ] 首次使用时能正常登录
- 笔记：

### 任务 4：修复检测间隔生效问题
- 状态：completed
- 笔记：
  - 添加了 intervalCh channel
  - 修改了 checkLoop 函数，响应间隔变化
  - 修改了 setInterval 函数，发送间隔变化通知
  - 编译测试通过
- 依赖：无
- 预估时间：15 分钟
- 复杂度：中等
- 验收标准：
  - [ ] 修改检测间隔后，下次检测使用新间隔
  - [ ] 不需要重启程序
  - [ ] 日志显示"检测间隔已更新为 X 秒"
- 笔记：

### 任务 5：设置页面显示 API 信息
- 状态：completed
- 笔记：
  - 添加了 custom_api textarea 字段
  - 添加了默认 API 说明
  - 修改了保存逻辑，处理 custom_api 字段
  - 编译测试通过
- 依赖：任务 1
- 预估时间：20 分钟
- 复杂度：中等
- 验收标准：
  - [ ] 设置页面显示认证服务器地址
  - [ ] 设置页面显示当前登录 API（完整 URL）
  - [ ] 可以编辑认证服务器地址
  - [ ] 可以编辑登录 API
  - [ ] 保存后立即生效
- 笔记：

## 迭代日志

（待执行时填写）
