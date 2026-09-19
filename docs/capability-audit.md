# 门店领域能力核查

## 核查范围

核对了 `dextea-admin` 的 `stores`、`store-catalog`、菜单门店关联、仪表盘，以及 `dextea-store`、`dextea-customer-api`、`dextea-trade` 当前对门店数据的读取/写入方式。

## Store Service 原有能力

已有能力足以覆盖门店主实体的基本生命周期：单条读取、按账号读取、创建、分页列表、关键字查询、附近门店、资料/位置/状态更新，以及凭证操作。当前 RPC 已按调用平面拆分为 `StoreAdminService`、`StoreBusinessService` 和 `StoreCredentialService`。管理视图与顾客业务视图使用不同消息，业务视图不包含账号、邮箱和时间字段。

## 已补能力

| 缺口 | 统一能力 | 解决的迁移场景 |
| --- | --- | --- |
| 各服务需要批量加载门店 | `GetStores` | customer-api 附近结果回填、trade 订单历史批量展示、商品/菜单跨域校验 |
| 管理端按区域找门店 | `ListStores` 的省/市/区过滤 | 菜单按区域分发前获取门店集合 |
| 顾客端需要城市枚举 | `StoreBusinessService.ListStoreCities` | 城市筛选，不再直接查 stores |
| 仪表盘需要门店计数 | `GetStoreStatistics` | 总数和状态分布，不再直接聚合 stores |
| 位置更新无法部分修改 | 经纬度改为 optional | 只修改地址文本或只修改区域字段时不覆盖坐标 |
| 输入边界不完整 | 坐标、附近距离、返回数量校验 | 避免无效坐标和无限制查询 |
| gRPC 业务接口默认无保护 | 可配置服务令牌拦截器（默认关闭） | 迁移完成后可在不改领域代码的情况下开启服务间认证 |

## 明确不补的能力

`store_menus`、`product_store_status`、`customization_option_store_status`、`store_ingredients` 不应放入 Store Service。它们描述的是商品/菜单/库存与门店的关系，应该由商品服务或库存域拥有；Store Service 仅提供门店事实查询，供这些服务通过 RPC 校验和补全展示。

同理，地理编码、Redis GEO 缓存同步、JWT/session、admin RBAC 和 HTTP 响应包装都属于调用方或基础设施适配，不是门店领域核心 RPC。

## RPC 平面边界

- `StoreAdminService`：门店创建、管理查询、资料/位置/状态更新、统计和系统重置密码。
- `StoreBusinessService`：顾客详情、批量门店、带距离排序的搜索、城市列表和附近门店；只返回业务可见字段。
- `StoreCredentialService`：门店账号认证和门店主动改密；不授予顾客业务调用方。
