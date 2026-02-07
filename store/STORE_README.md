# WeChat 数据存储模块 (New Store)

本文档详细介绍了 `store` 目录下新一代存储模块的设计与使用。该模块旨在替代原有的 `wechatdb` 实现，提供更清晰的分层架构、更好的可测试性以及对多平台/多版本数据库的统一抽象。

## 1. 架构概览

新架构采用了严格的分层设计，自底向上分为四层：

1.  **Core (核心层)**: 负责最基础的资源管理。
    *   `ConnectionPool`: **(原 DBManager)** 管理 SQLite 连接池，处理文件的打开、关闭、并发安全以及只读模式设置。
    *   `Watcher`: 封装文件系统监听 (`fsnotify`)，提供事件驱动的能力。
2.  **Strategy (策略层)**: 负责屏蔽平台差异。
    *   定义了不同平台（如 Windows V3, macOS V4）的文件命名规则。
    *   负责识别文件类型（Message, Contact, Image 等）。
3.  **Bind (绑定/路由层)**: 负责逻辑到物理的映射。
    *   **TimelineRouter**: **(原 Binder)** 维护了一个基于时间线（Timeline）的索引。它利用 `MetadataReader` 提取的元信息，根据查询的时间范围和对话者（Talker），智能路由到具体的物理数据库分片（Shard）。
    *   **MetadataReader**: 负责执行具体的 SQL 操作以提取元数据（如 `DBInfo` 中的开始时间，`Name2ID` 映射）。
4.  **Repo (仓储层)**: 负责业务数据的存取。
    *   提供面向领域的接口（`GetMessages`, `GetContacts` 等）。
    *   将复杂的查询流程拆解为：路由 -> 分片查询 -> 聚合 -> 排序 -> 分页。

最上层通过 `Store` 接口对外暴露统一的能力。

## 2. 目录结构

```text
store/
├── store.go             # [接口] 定义对外的核心接口 Store
├── store_impl.go        # [实现] DefaultStore 的具体实现
├── types/               # [类型] 定义通用的查询参数结构体 (MessageQuery 等)
├── core/                # [核心层]
│   ├── db_manager.go    # SQLite 连接池实现 (ConnectionPool)
│   └── watcher.go       # 文件监控实现
├── strategy/            # [策略层]
│   ├── interface.go     # 策略接口定义
│   └── windows_v3.go    # Windows V3 版本的具体策略
├── bind/                # [绑定层]
│   ├── binder.go        # 消息路由逻辑 (TimelineRouter)
│   └── metadata.go      # 元数据提取工具 (MetadataReader)
└── repo/                # [仓储层]
    ├── repository.go    # Repo 入口
    ├── message.go       # 消息查询实现 (原子化拆分)
    ├── contact.go       # 联系人查询实现
    ├── chatroom.go      # 群聊查询实现 (含兜底逻辑)
    ├── session.go       # 会话查询实现
    └── media.go         # 媒体资源路径解析
```

## 3. 使用指南

### 初始化

使用 `NewStore` 方法即可初始化整个存储栈。它会自动扫描指定目录下的数据库文件，通过 `TimelineRouter` 建立索引并启动文件监控。

```go
import "github.com/afumu/wetrace/store"

func main() {
    workDir := "/path/to/wechat/data"
    
    // 初始化 Store
    // 内部会自动创建 ConnectionPool, TimelineRouter 和 Watcher
    s, err := store.NewStore(workDir)
    if err != nil {
        panic(err)
    }
    defer s.Close()

    // ... 使用 s 进行查询
}
```

### 查询数据

查询参数通过 `store/types` 包中的结构体传递。

```go
import (
    "context"
    "time"
    "github.com/afumu/wetrace/store/types"
)

func query(s store.Store) {
    ctx := context.Background()

    // 查询联系人
    contacts, _ := s.GetContacts(ctx, types.ContactQuery{
        Keyword: "zhangsan",
        Limit:   10,
    })

    // 查询消息 (自动跨库路由)
    // TimelineRouter 会自动计算需要查询哪些 MSGx.db 文件
    msgs, _ := s.GetMessages(ctx, types.MessageQuery{
        StartTime: time.Now().Add(-24 * time.Hour), // 最近24小时
        EndTime:   time.Now(),
        Talker:    "wxid_123456",
    })
}
```

### 监听变化

可以通过 `Watch` 方法监听文件系统的变化（如新消息到达）。当检测到新的消息数据库文件生成时，`DefaultStore` 会自动触发 `RebuildIndex` 以刷新路由表。

```go
s.Watch("Message", func(event fsnotify.Event) error {
    fmt.Printf("检测到文件变更: %s\n", event.Name)
    return nil
})
```

## 4. 扩展指南

### 添加新平台支持

1.  在 `store/strategy` 下新建文件（如 `macos_v4.go`）。
2.  实现 `Strategy` 接口，定义该平台的文件命名正则。
3.  在 `store_impl.go` 的 `NewStore` 工厂方法中，根据配置或自动检测选择新的策略。

### 添加新数据表支持

1.  在 `store/types/query.go` 中定义新的 Query 结构体。
2.  在 `store/store.go` 接口中添加新的方法签名。
3.  在 `store/repo` 下新建文件（或修改现有文件），实现具体的 SQL 查询逻辑。
4.  在 `store/store_impl.go` 中实现接口方法并委托给 Repo。

## 5. 设计决策 QA

**Q: 为什么要重构为 `ConnectionPool` 和 `TimelineRouter`？**
A: 原有的 `Binder` 承担了过多职责（扫描、读取 SQL、路由）。重构后：
*   `ConnectionPool` 专注于管理 SQL 连接的生命周期。
*   `TimelineRouter` 专注于路由算法（时间线重叠计算）。
*   `MetadataReader` 专注于具体的 SQL 提取逻辑。
这样每个组件都符合单一职责原则（SRP），代码更易读、易测。

**Q: `MetadataReader` 的作用是什么？**
它是一个无状态的工具类，专门负责从 SQLite 中读取特定的元数据（如 `DBInfo` 表中的开始时间，`Name2ID` 表中的映射关系）。将这部分 SQL 逻辑从路由逻辑中剥离，使得路由器的代码更加纯粹。

**Q: 如何处理并发读取？**
A: SQLite 默认开启了 `cache=shared` 和 `mode=ro` (只读模式)。`ConnectionPool` 内部维护了连接池，并使用 `sync.RWMutex` 保护，确保多协程下的安全访问。

**Q: 为什么 Query 参数要封装成结构体？**
A: 随着过滤条件（发送者、类型、关键词、分页）的增加，函数签名会变得非常长。封装成结构体（如 `MessageQuery`）具有更好的扩展性，未来增加参数不需要破坏接口兼容性。
