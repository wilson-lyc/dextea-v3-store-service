# Store Service 设计

共享 RPC 契约统一放在 `dextea-proto/proto/store/v1/store.proto`，本服务只实现门店聚合的领域规则和持久化。

## 领域边界

Store Service 拥有 `stores` 表及以下能力：

- 门店资料、区域、地址、坐标、营业状态和账号唯一性；
- 创建、分页列表、按账号/ID读取、批量按 ID 读取、关键字/区域查询；
- 附近门店查询、区域枚举、状态统计；
- 凭证校验、店铺端改密、管理端重置密码。

调用方继续拥有自己的协议适配：HTTP、JWT、session、地理编码和响应格式不下沉到本服务。

## RPC 平面拆分

### `StoreAdminService`

供 `dextea-admin` 等管理端使用，包含门店创建、分页、按账号/ID查询、状态统计、资料/位置/状态变更和系统重置密码。返回的 `Store` 是管理视图，可以包含账号、邮箱和时间字段。

### `StoreBusinessService`

供 customer 等顾客业务端使用，只提供门店可见事实：详情、批量查询、按区域/关键字并按距离排序的搜索、城市列表和附近门店。业务视图不包含账号、邮箱、创建更新时间或任何密码字段；距离计算由 Store Service 完成。

### `StoreCredentialService`

供门店账号相关流程使用，独立提供登录校验和门店主动改密。它不属于顾客只读业务接口，也不与管理端的系统重置密码混在一起。

## 服务间认证

`auth.enabled` 默认关闭以兼容本地开发。生产环境启用后，所有非公开 RPC 请求都必须携带 metadata `x-service-token`；health check 和 reflection 保持可用。必须分别配置 `auth.admin-token`、`auth.business-token` 和 `auth.credential-token`，三类令牌不做回退。

商品、菜单、客制化、原料及其门店状态属于商品域。`store_menus` 是商品/菜单关联，不能因为管理端页面挂在“门店”下就迁入 Store Service。订单、取餐码、员工和 RBAC 也不属于本服务。

## 通用查询约定

- 业务端批量查询单次最多接收 500 个 ID，按请求 ID 的首次出现顺序返回，重复 ID 去重；不存在的 ID 省略，只返回可见门店。
- 业务端搜索必须携带经纬度，默认过滤不可用门店，并由 Store Service 按距离升序返回。
- 业务端城市列表只返回去重且稳定排序的城市名称，调用方可以自行做拼音分组。
- 业务端附近门店只查询有有效坐标的可用门店，距离按公里升序返回；当前实现使用 MySQL 读取后在 Go 内计算 Haversine，后续可替换为空间索引而不改变 RPC。
- 管理端分页列表支持关键字、省、市、区/县和精确状态筛选。
- 管理端统计返回四种门店状态的完整计数，包含计数为 0 的状态，便于仪表盘稳定渲染。

## 凭证边界

`AuthenticateStore` 只返回门店 ID、名称和状态，不签发 token；`ChangeStorePassword` 必须校验旧密码；`ResetStorePassword` 不需要旧密码，仅供受控的管理流程使用。普通 `Store` 消息永远不暴露密码哈希。

## 管理端迁移核对表

| 现有能力 | 迁移目标 | 归属 |
| --- | --- | --- |
| 门店 CRUD、分页、资料/位置/状态 | `StoreAdminService` | Store Service |
| 门店登录校验、店铺端改密 | `StoreCredentialService` | Store Service 持有凭证，BFF 持有 HTTP/JWT |
| 顾客门店详情、搜索、附近、城市列表 | `StoreBusinessService` | Store Service |
| 管理端仪表盘门店总数和状态分布 | `StoreAdminService.GetStoreStatistics` | Store Service |
| 菜单与门店绑定、按区域分发菜单 | 商品服务的菜单域；通过 `StoreBusinessService` 查询门店 | 商品服务拥有关联 |
| 商品/客制化选项门店状态 | 商品服务 RPC | 商品服务 |
| 门店原料数据 | 商品/库存域 | 不属于 Store Service |
| Redis GEO 同步 | 调用方缓存或 Store Service 内部实现细节 | 不作为外部 RPC 能力 |

迁移顺序建议是先让 customer 只读切换到 `StoreBusinessService`，再让 admin 切换到 `StoreAdminService`，随后迁移门店账号流程到 `StoreCredentialService`，最后删除所有调用方对 `stores` 的直接 SQL。
