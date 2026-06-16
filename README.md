# OneRSS

一个现代化的、注重隐私的自托管 RSS 阅读器，具有 AI 增强功能。

## 功能特性

- 🌐 **多源支持** - 标准 RSS/Atom、HTML+XPath、自定义脚本、邮件订阅
- 🤖 **AI 增强** - 自动翻译、智能摘要、AI 聊天
- 🔌 **丰富集成** - Obsidian、Notion、Zotero、FreshRSS、RSSHub
- 🐳 **Docker 部署** - 一键部署，轻松自托管
- 🔒 **隐私优先** - 本地存储，无外部追踪

## 技术栈

### 后端
- Go 1.25+
- SQLite (现代纯 Go 实现)
- 标准库 HTTP 服务器

### 前端
- Vue 3.5+ (Composition API)
- TypeScript
- Pinia (状态管理)
- Tailwind CSS 3.4+
- Vite 5+

## 快速开始

### Docker 部署 (推荐)

```bash
# 构建镜像
docker build -t one-rss:latest .

# 运行容器
docker run -d \
  --name one-rss \
  -p 6011:6011 \
  -v one-rss-data:/data \
  one-rss:latest
```

访问 http://localhost:6011 即可使用。

### 开发模式

```bash
# 安装依赖
make install

# 启动开发服务器
make dev
```

### 构建生产版本

```bash
# 构建二进制文件
make build

# 运行
./bin/one-rss
```

## 环境变量

| 变量 | 默认值 | 描述 |
|------|--------|------|
| PORT | 6011 | 服务器端口 |
| DB_PATH | one-rss.db | 数据库文件路径 |

## 项目结构

```
one-rss/
├── main.go                    # 应用入口
├── internal/                  # 后端 Go 代码
│   ├── ai/                    # AI 功能
│   ├── cache/                 # 缓存管理
│   ├── config/                # 配置管理
│   ├── database/              # 数据库操作
│   ├── discovery/             # 订阅源发现
│   ├── feed/                  # RSS/Atom 解析
│   ├── handlers/              # HTTP API 处理器
│   ├── models/                # 数据模型
│   ├── rules/                 # 过滤规则
│   ├── statistics/            # 统计功能
│   ├── summary/               # 摘要生成
│   ├── translation/           # 翻译服务
│   └── utils/                 # 工具函数
├── frontend/                  # Vue 3 前端
│   ├── src/
│   │   ├── components/        # Vue 组件
│   │   ├── composables/       # 可复用逻辑
│   │   ├── stores/            # Pinia 状态管理
│   │   ├── types/             # TypeScript 类型
│   │   ├── i18n/              # 国际化
│   │   └── utils/             # 工具函数
│   └── package.json
├── docs/                      # 文档
├── Makefile                   # 构建命令
└── Dockerfile                 # Docker 配置
```

## API 文档

### 健康检查

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/health | 健康检查 |

### 订阅源管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/feeds | 获取所有订阅源 |
| POST | /api/feeds/add | 添加订阅源 |
| POST | /api/feeds/delete | 删除订阅源 |
| POST | /api/feeds/update | 更新订阅源 |
| POST | /api/feeds/refresh | 刷新订阅源 |

### 文章管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/articles | 获取文章列表 |
| GET | /api/articles/content | 获取文章内容 |
| POST | /api/articles/read | 标记已读 |
| POST | /api/articles/favorite | 切换收藏 |
| POST | /api/articles/toggle-read-later | 切换稍后阅读 |
| POST | /api/articles/mark-all-read | 全部标记已读 |

## 开发命令

```bash
make dev          # 启动开发服务器
make build        # 构建生产版本
make test         # 运行测试
make lint         # 代码检查
make clean        # 清理构建产物
make install      # 安装依赖
make docker       # 构建 Docker 镜像
make docker-run   # 运行 Docker 容器
```

## 许可证

MIT License
