# Tasks
- [x] Task 1: 梳理现有样式与输出视觉基线
  - [x] 盘点全局变量与通用类（src/index.css、src/App.css）
  - [x] 补齐设计令牌（颜色/圆角/阴影/间距/字体层级）并给出命名约定

- [x] Task 2: 抽取并落地可复用的 UI 基础能力
  - [x] 统一 Button / Card / Alert / Badge / FormField 的样式与使用方式
  - [x] 为 Loading / Empty / Error 三态提供统一展示（组件或样式类）

- [x] Task 3: 美化前台公共布局与首页
  - [x] 优化 Header/Nav（active/hover、间距、移动端换行体验）
  - [x] 优化 Footer（信息层级与弱化文本）
  - [x] 首页 Hero 与入口卡片强化层级与引导（不新增依赖）

- [x] Task 4: 美化前台列表页与详情页
  - [x] Activities/Services/LostFound 列表：查询区布局、列表项信息密度与操作区一致性
  - [x] 各详情页：标题区、元信息、正文排版与返回入口统一
  - [x] 清理页面内联 style，替换为可复用类/组件

- [x] Task 5: 美化后台管理端
  - [x] AdminLayout：侧栏/主区层级、当前导航态与移动端折叠/换行
  - [x] AdminLogin：表单卡片、错误提示与按钮层级
  - [x] AdminLostItems：筛选区、记录卡片、状态/类型徽标、操作区布局
  - [x] AdminLostItemForm：字段分组、必填提示、提交反馈样式

- [x] Task 6: 体验验收与回归验证
  - [x] 覆盖明暗主题（prefers-color-scheme）在关键页面的对比度与可读性
  - [x] 响应式与键盘可访问性检查（焦点态、可点击面积）
  - [x] 运行 lint/build 并修复引入的样式或结构问题

# Task Dependencies
- Task 3 depends on Task 2
- Task 4 depends on Task 2
- Task 5 depends on Task 2
- Task 6 depends on Task 3 and Task 4 and Task 5
