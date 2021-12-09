package pg

import (
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/global"
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/service"
)

func RunPgDemo() {
	global.InitStore()
	service.Preload()   // 插入数据，进行一些预处理
	service.Isolation() // 模拟不同隔离级别
}
