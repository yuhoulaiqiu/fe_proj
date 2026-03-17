# 社区便民公益互助平台（前端 + 后台 CRUD）

面向社区居民/志愿者/物业工作人员的演示网站，包含 3 个功能模块：公益活动、便民服务目录、失物招领；并提供后台管理端，对“失物招领”模块实现完整 CRUD（增删改查）。

## 目录结构
- src/：前端（React + React Router + Axios）
- server/：后端（Go + Gin + MySQL），提供 /api 接口与数据表

## 功能入口
- 用户端：
  - /（首页）
  - /activities（公益活动）
  - /services（便民服务）
  - /lost-found（失物招领）
- 后台管理端：
  - /admin/login（登录）
  - /admin/lost-items（失物招领管理 CRUD）

## 本地开发运行

### 1) 启动后端 API（Go）
前置条件：已安装 Go（建议 1.19+）。

```bash
cd server
go run .
```

默认启动：
- API 地址：http://localhost:8080
- MySQL 数据库：community_help_hub（首次启动自动创建表；数据库需提前创建）

默认管理员账号（可通过环境变量覆盖）：
- 用户名：admin
- 密码：admin123

可选环境变量：
- PORT：后端端口（默认 8080）
- MYSQL_DSN：MySQL 连接串（优先使用）
- MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PASSWORD / MYSQL_DATABASE：用于拼接 DSN（默认 host=127.0.0.1, port=3306, user=root, database=community_help_hub）
- ADMIN_USERNAME / ADMIN_PASSWORD：初始化管理员账号密码
- ADMIN_INIT：是否初始化管理员账号（默认开启；可设为 0/false/off/no 关闭）
- CORS_ORIGINS：允许的前端来源（逗号分隔），例如：
  - CORS_ORIGINS=http://localhost:5173,http://127.0.0.1:5173

初始化数据库与表（示例）：
```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < server/schema.mysql.sql
```

### 2) 启动前端（React + Vite）
前置条件：Node.js 版本需满足 Vite 要求（建议 20.19+ 或 22.12+）。

```bash
npm i
npm run dev -- --host 0.0.0.0 --port 5173
```

浏览器访问：
- http://localhost:5173

后端 API 默认走同域的 /api（开发态可配置）：

PowerShell 示例：
```powershell
$env:VITE_API_BASE_URL="http://localhost:8080"
npm run dev -- --host 0.0.0.0 --port 5173
```

macOS/Linux 示例：
```bash
VITE_API_BASE_URL=http://localhost:8080 npm run dev -- --host 0.0.0.0 --port 5173
```

## 生产部署（远程终端可访问）

### 方案 A：快速演示（Vite preview）
适合临时验收，不建议长期生产使用。

```bash
npm i
npm run build
npm run preview -- --host 0.0.0.0 --port 4173
```

访问：
- http://<server-ip>:4173

### 方案 B：静态托管 + 反向代理（推荐）
1) 前端构建产物：dist/
```bash
npm i
npm run build
```

2) 后端作为独立服务运行
```bash
cd server
PORT=8080 CORS_ORIGINS=http://<your-domain-or-ip> go run .
```

3) Nginx 反代示例（核心思路）
- / 走静态文件
- /api/ 代理到后端

```nginx
server {
  listen 80;
  server_name _;

  root /var/www/community-help-hub/dist;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
  }
}
```

### 防火墙/安全组
按你的部署方式放行端口：
- 演示 preview：4173
- 生产 Nginx：80/443
- 后端直连（若不走 Nginx）：8080

## 验收与联调（curl）

健康检查：
```bash
curl http://localhost:8080/health
```

获取公益活动列表：
```bash
curl "http://localhost:8080/api/activities?page=1&pageSize=10"
```

后台登录获取 token：
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
```

创建失物招领（将 <TOKEN> 替换为登录返回的 token）：
```bash
curl -X POST http://localhost:8080/api/admin/lost-items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d "{\"title\":\"测试记录\",\"itemType\":\"lost\",\"status\":\"open\",\"location\":\"A区门口\",\"occurredAt\":\"2026-03-16 14:30\",\"description\":\"测试描述\",\"contact\":\"张三 138****0000\"}"
```

## 常见问题

### Vite/Node 版本提示
如果出现 “Vite requires Node.js version …”：
- 升级 Node.js 到 20.19+ 或 22.12+
- Windows 可用 nvm-windows 管理多版本 Node

### 依赖安装异常（可选依赖/原生 binding）
若 npm 安装后构建提示缺少可选依赖：
- 删除 node_modules 与 package-lock.json 后重新执行 npm i
