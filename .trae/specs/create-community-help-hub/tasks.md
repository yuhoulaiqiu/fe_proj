# Tasks
- [x] Task 1: 确定信息架构与路由骨架
  - [x] 定义页面路由与导航结构（Home/Activities/Services/Lost&Found/Admin）
  - [x] 统一布局组件（Header/Nav/Footer）与响应式布局基线

- [x] Task 2: 实现普通用户端三大模块页面
  - [x] 公益活动：列表、详情、加载与错误态
  - [x] 便民服务：列表、分类筛选、搜索、加载与错误态
  - [x] 失物招领：列表、详情、筛选/搜索、加载与错误态

- [x] Task 3: 设计并落地后端 API 与数据库（最小可用）
  - [x] 建立数据模型与迁移（users/activities/services/lost_items）
  - [x] 提供公开查询接口（GET 列表/详情）
  - [x] 提供管理员鉴权接口（login + token 校验）
  - [x] 提供失物招领管理接口（POST/PUT/DELETE）

- [x] Task 4: 实现后台管理端（独立入口）与失物招领 CRUD
  - [x] 后台登录页（表单校验、错误提示、token 持久化）
  - [x] 管理列表页（分页/搜索/状态筛选）
  - [x] 新增/编辑表单页（字段校验、提交反馈）
  - [x] 删除与状态更新（确认弹窗、乐观/悲观更新策略）

- [x] Task 5: 部署与运行流程文档化 + 可远程访问验证
  - [x] 本地开发流程：前端、后端、环境变量、联调验证
  - [x] 生产部署流程：前端构建、静态托管/反代、后端启动、端口与防火墙
  - [x] 验收脚本：curl/浏览器验证关键接口与页面

# Task Dependencies
- Task 2 depends on Task 1
- Task 4 depends on Task 3
- Task 5 depends on Task 2 and Task 4
