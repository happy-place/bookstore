package svc

import (
	"bookstore/rpc/check/internal/config"
	"bookstore/rpc/model"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	c config.Config
	Model model.BookModel   // 手动代码
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		c: c,
		Model: model.NewBookModel(sqlx.NewMysql(c.DataSource), c.Cache, c.Table), // 手动代码
	}
}
