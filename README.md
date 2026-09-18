# dextea-store-service

德贤茶门店领域微服务。该服务预留为所有业务端共享的门店领域核心，使用 Go 与 gRPC/RPC 实现。

当前阶段只提供项目骨架，不包含具体门店业务接口和数据库读写逻辑。

## 目录结构

```text
dextea-store-service/
├── cmd/
│   └── server/              # 服务启动入口
├── configs/                 # 配置文件目录
├── docs/                    # 服务设计与 RPC 文档
├── internal/
│   ├── config/              # 配置加载
│   ├── model/               # 领域模型（待实现）
│   ├── registry/            # 服务注册与发现（待实现）
│   ├── repository/          # 数据访问层（待实现）
│   ├── rpc/                 # gRPC 服务实现与适配层
│   └── service/             # 领域服务（待实现）
├── migrations/              # 数据库迁移（待实现）
├── scripts/                 # 运维/生成脚本（待实现）
├── go.mod
└── Makefile
```

## 本地运行

要求 Go 1.23+。

```bash
go mod download
go run ./cmd/server -addr :9092
```

当前进程提供 gRPC health check 和 reflection，业务 RPC 协议将在 `dextea-proto` 中定义后接入。

## 后续接入顺序

1. 在 `dextea-proto/proto/store/v1/` 定义门店领域 RPC 契约并生成 Go 代码。
2. 实现 `internal/repository`、`internal/service` 与 `internal/rpc`。
3. 已接入 MySQL 与 Nacos 服务注册；通过 `configs/config.yaml` 配置 `dextea-store-service` 的 gRPC 注册信息。
4. 逐步让管理端、店铺端和顾客 API 停止直接访问门店表。
