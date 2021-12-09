package mysql

import (
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/mysql/global"
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/mysql/service"
)

func RunMySQLDemo() {
	global.InitStore()
	service.Preload()
	service.Isolation()
}
