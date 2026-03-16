# 社区互助站点页面美化 Spec

## Why
现有页面功能已完整，但视觉层级、组件一致性与细节（间距/排版/状态反馈）仍偏基础。通过一次不改业务逻辑的 UI 美化，提升可读性、专业度与整体体验。

## What Changes
- 统一视觉规范：基于现有 CSS 变量补齐色板/字号/圆角/阴影/间距等设计令牌（design tokens）
- 统一基础组件外观：按钮、输入框/选择框、卡片、提示（success/warn/danger）、徽标（badge）、分割线
- 优化整体布局：Header/Nav/Footer、页面标题区、列表/详情双栏在移动端的响应式表现
- 优化状态反馈：加载、空数据、错误态使用一致的样式与文案布局（避免页面跳动）
- 清理不一致实现：减少页面内联样式（style={{...}}），改为可复用的样式类/组件
- **不引入新依赖**：不新增 UI 框架/组件库；优先复用现有 CSS（src/index.css、src/App.css）

## Impact
- Affected specs: 视觉规范、基础组件一致性、响应式布局、可访问性体验
- Affected code:
  - 样式：src/index.css、src/App.css、src/App.jsx
  - 布局组件：src/components/SiteLayout.jsx、src/components/AdminLayout.jsx
  - 页面：src/pages/*、src/pages/admin/*

## ADDED Requirements
### Requirement: 全站视觉一致性
系统 SHALL 在所有前台与后台页面提供一致的排版、间距与组件风格（按钮/表单/卡片/提示/徽标）。

#### Scenario: 浏览任意列表页
- **WHEN** 用户打开“公益活动/便民服务/失物招领/后台管理”任一列表页
- **THEN** 页面包含统一的标题区（标题+说明+可选统计）
- **AND** 查询区域使用统一表单布局（多列在窄屏自动换行/堆叠）
- **AND** 列表项使用一致的卡片层级（标题、摘要、元信息、操作区）

### Requirement: 状态反馈一致且不突兀
系统 SHALL 为加载、空数据与错误三类状态提供一致且清晰的视觉反馈，并避免布局抖动。

#### Scenario: 列表加载中
- **WHEN** 页面请求数据中
- **THEN** 显示一致的加载态（可为骨架屏或占位卡片），并保持列表区域高度相对稳定

#### Scenario: 无数据
- **WHEN** 请求成功但返回 items 为空
- **THEN** 显示统一的空态组件（标题/说明/可选操作），视觉不弱于错误态

#### Scenario: 请求失败
- **WHEN** 请求失败
- **THEN** 以统一的 Alert 样式展示错误信息，并提供明显的重试入口（若页面已有“搜索/刷新”入口则复用）

### Requirement: 响应式与可访问性
系统 SHALL 在常见屏宽下保持可读性，并提供可见的键盘焦点样式。

#### Scenario: 移动端浏览
- **WHEN** 屏宽小于等于 1024px
- **THEN** 顶部导航可换行且不遮挡主要内容
- **AND** 列表/详情页的“标题-内容-操作”布局自动堆叠，按钮不会溢出

#### Scenario: 键盘操作
- **WHEN** 用户使用 Tab 在链接、按钮、输入框间切换
- **THEN** 焦点样式清晰可见且与主题色一致

### Requirement: 后台管理的可扫描性提升
系统 SHALL 提升后台列表项的可扫描性：状态、关键字段与操作更易识别。

#### Scenario: 查看失物招领管理列表
- **WHEN** 管理员浏览列表
- **THEN** 每条记录展示状态徽标（未处理/已认领/已归还）与类型徽标（失物/招领）
- **AND** 操作按钮（编辑/删除）在窄屏下自动换行且保持可点击面积

## MODIFIED Requirements
### Requirement: 业务逻辑保持不变
系统 SHALL 保持现有路由、API 调用、表单校验与 CRUD 行为不变，仅调整 UI 展示与组件复用方式。

## REMOVED Requirements
无
