# 社区便民公益互助平台（前端+后台CRUD）Spec

## Why
后端开发者缺乏前端经验时，容易做出“能跑但不好用”的页面。本项目以“社区便民+公益互助”为主题，提供清晰的信息展示与可落地的后台增删改查能力，满足课程/作品对模块化、交互、请求响应与数据来源（数据库）的要求。

## What Changes
- 新增一个面向普通用户的前端网站：包含至少 3 个清晰区分的功能模块与多页面路由
- 新增一个独立的后台管理端（Web 管理页面）：对至少 1 个模块提供完整 CRUD（增删改查）
- 新增一个后端 HTTP API（请求/响应）与数据库表结构：为前端大部分页面数据提供来源
- 提供可远程访问的部署方案：开发态与生产态均可从远程终端启动并访问
- 统一项目目录结构与引用方式（React Router、API 客户端、页面/组件分层）

## Impact
- Affected specs: 页面路由与导航、列表/详情展示、表单交互、鉴权登录、CRUD 管理、API 规范、数据库结构、部署与运行流程
- Affected code: 
  - 前端：src（路由/页面/组件/API/样式），vite.config.js（必要时开放 host）
  - 后端：新增 server/（Go 服务）或独立后端仓库（以本 spec 为准）
  - 部署：可选 docker-compose.yml / Nginx 配置（实现阶段按任务落地）

## ADDED Requirements

### Requirement: 主题与使用群体
系统 SHALL 以“社区便民公益互助”为主题，面向社区居民/志愿者/物业工作人员，内容正向积极，并在现实生活具备明确使用场景。

#### Scenario: 用户进入首页获取信息
- **WHEN** 用户打开网站首页
- **THEN** 能看到平台定位说明、三个模块入口与最新内容摘要

### Requirement: 三个明显区分的功能模块
系统 SHALL 至少包含以下三个可明显区分的内容模块，且每个模块均可通过页面导航进入并完成主要操作：
1) 公益活动：活动列表、活动详情、报名/取消报名（可选）  
2) 便民服务：服务目录（例如：家政、维修、医疗、办事指南），支持搜索与分类筛选  
3) 失物招领：失物/招领信息列表与详情，普通用户可浏览与筛选

#### Scenario: 用户通过导航在模块间切换
- **WHEN** 用户点击顶部导航或首页模块卡片
- **THEN** 页面路由切换流畅，URL 清晰（例如 /activities、/services、/lost-found）

### Requirement: 后台管理端对“失物招领”提供 CRUD
系统 SHALL 提供一个独立后台管理端页面（/admin），对“失物招领”模块提供增删改查：
- 新增：发布失物/招领信息（标题、类型、地点、时间、描述、联系方式、状态）
- 查询：列表分页/搜索（至少按关键词或状态）
- 修改：更新字段与状态（未认领/已认领/已归还等）
- 删除：支持删除记录（软删除优先；若硬删需说明影响）

#### Scenario: 管理员新增一条失物信息
- **WHEN** 管理员登录后台并提交新增表单
- **THEN** 后端返回创建成功响应，前端列表即时刷新并可查看详情

### Requirement: 请求与响应效果正常
系统 SHALL 通过 HTTP API 完成数据获取与提交，并在前端体现加载、成功与失败状态（例如 Toast/提示条/错误页）。

#### Scenario: 后端暂时不可用
- **WHEN** 前端请求接口超时或返回 5xx
- **THEN** 页面提示“稍后再试”，并允许用户重试，不出现页面崩溃

### Requirement: 数据主要由数据库提供
系统 SHALL 让“活动、便民服务、失物招领”三类数据主要来自后端数据库（非前端硬编码）。

#### Scenario: 前端刷新页面
- **WHEN** 用户刷新任一模块页面
- **THEN** 页面数据由后端 API 再次拉取，内容与数据库一致

### Requirement: 远程终端可访问
系统 SHALL 支持在远程服务器通过终端启动服务，并在局域网/公网通过浏览器访问：
- 开发态：前后端分别启动（前端 dev server、后端 API）
- 生产态：前端静态资源 + 反向代理到后端 API

#### Scenario: 远程启动后访问
- **WHEN** 运维在服务器执行启动命令
- **THEN** 访问 http://<server-ip>:<port>/ 可打开网站并产生正常 API 请求

## MODIFIED Requirements
无（新增项目能力为主）。

## REMOVED Requirements
无。

## API 设计（最小可用）
以下为建议的最小接口集合，实现阶段可微调，但需保持前后端一致并记录变更。

### Auth
- POST /api/auth/login  
  - req: { username, password }
  - resp: { token, expiresAt, user: { id, username, role } }
- POST /api/auth/logout（可选）
- GET /api/auth/me（可选）

### Activities（公益活动）
- GET /api/activities?keyword=&page=&pageSize=
- GET /api/activities/:id

### Services（便民服务目录）
- GET /api/services?category=&keyword=&page=&pageSize=
- GET /api/services/:id

### Lost&Found（失物招领）
- GET /api/lost-items?type=&status=&keyword=&page=&pageSize=
- GET /api/lost-items/:id
- POST /api/admin/lost-items（鉴权）
- PUT /api/admin/lost-items/:id（鉴权）
- DELETE /api/admin/lost-items/:id（鉴权）

## 数据库结构（建议）
实现阶段可选 MySQL/PostgreSQL/SQLite（开发态 SQLite 优先，部署态可切 MySQL）。

- users: id, username, password_hash, role, created_at
- activities: id, title, cover_url, summary, content, location, start_time, end_time, created_at
- services: id, name, category, phone, address, description, updated_at
- lost_items: id, title, item_type, status, location, occurred_at, description, contact, images_json, created_at, updated_at, deleted_at(可选)

## 前端信息架构（建议）
- /：首页（平台介绍 + 最新内容摘要）
- /activities：公益活动列表
- /activities/:id：活动详情
- /services：便民服务目录（分类/搜索）
- /lost-found：失物招领列表（筛选/搜索）
- /lost-found/:id：详情
- /admin/login：后台登录
- /admin/lost-items：失物招领管理列表（CRUD）
- /admin/lost-items/new 与 /admin/lost-items/:id/edit：编辑表单

## 部署与运行（目标输出）
实现阶段需最终交付一份可复制执行的流程，至少覆盖：
- 本地开发：安装依赖、启动前端、启动后端、连通性验证
- 远程部署：构建前端、启动后端、配置反向代理或端口映射、开放防火墙端口、验证 URL
