package repo

import (
	"github.com/afumu/wetrace/store/bind"
	"github.com/afumu/wetrace/store/core"
)

// Repository 是数据访问层的入口，聚合了路由和连接池
type Repository struct {
	router *bind.TimelineRouter
	pool   *core.ConnectionPool
}

// New 创建一个新的 Repository
func New(router *bind.TimelineRouter, pool *core.ConnectionPool) *Repository {
	return &Repository{
		router: router,
		pool:   pool,
	}
}
