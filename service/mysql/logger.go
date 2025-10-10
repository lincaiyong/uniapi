package mysql

import (
	"context"
	"github.com/lincaiyong/log"
	"gorm.io/gorm/logger"
	"time"
)

var traceFn func(string)

func SetTraceFn(fn func(sql string)) {
	traceFn = fn
}

type Logger struct {
}

func (l *Logger) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

func (l *Logger) Info(_ context.Context, s string, i ...interface{}) {
	log.InfoLog(s, i)
}

func (l *Logger) Warn(_ context.Context, s string, i ...interface{}) {
	log.WarnLog(s, i)
}

func (l *Logger) Error(_ context.Context, s string, i ...interface{}) {
	log.ErrorLog(s, i)
}

func (l *Logger) Trace(_ context.Context, _ time.Time, f func() (sql string, rowsAffected int64), _ error) {
	if traceFn != nil {
		sql, _ := f()
		traceFn(sql)
	}
}
