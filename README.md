# dextea-store-service

德贤茶门店领域微服务。该服务是所有业务端共享的门店领域核心，使用 Go 与 gRPC 实现。

服务只拥有门店聚合自身的数据和规则：门店资料、位置、营业状态、登录账号凭证，以及跨服务可复用的门店查询能力。HTTP/JWT/session 仍由各业务 BFF 负责。

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
├── migrations/              # 门店域数据库迁移
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

当前进程提供 gRPC health check、reflection，以及拆分后的 `StoreAdminService`、`StoreBusinessService` 和 `StoreCredentialService`。协议统一维护在 `dextea-proto/proto/store/v1/store.proto`。

## 后续接入顺序

1. customer 等业务调用方使用 `StoreBusinessService` 的位置感知查询、批量查询、城市列表和附近门店接口，不再直接访问 `stores` 表。
2. admin 管理调用使用 `StoreAdminService` 的创建、分页、资料/位置/状态更新、统计和重置密码接口。
3. 门店账号登录与改密使用独立的 `StoreCredentialService`；顾客业务 token 不应获得该接口权限。
4. 已接入 MySQL 与 Nacos 服务注册；通过 `configs/config.yaml` 配置 `dextea-store-service` 的 gRPC 注册信息。
5. 商品、菜单、客制化、原料及其门店可售状态不属于本服务，仍由商品服务拥有；订单、取餐码等交易数据也不迁入本服务。
