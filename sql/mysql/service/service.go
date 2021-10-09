package service

import (
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/mysql/global"
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/mysql/store"
	"strconv"
)

func Preload() {
	err := global.GMySQLDb.Migrator().DropTable(&store.Account{})
	if err != nil {
		return
	}
	err = global.GMySQLDb.Migrator().CreateTable(&store.Account{})
	if err != nil {
		return
	}
	// 创建表
	accounts := make([]store.Account, 0)
	for i := 0; i < 100; i++ {
		accounts = append(accounts, store.Account{
			Name:  "person" + strconv.Itoa(i),
			Money: i,
		})
	}
	err = store.BatchCreate(accounts)
	if err != nil {
		return
	}
}

func Isolation() {
	// 1. READ UNCOMMITTED pg不支持 所以在pg中不会出现dirty-read

	// ReadUnCommit() // 模拟脏读

	// 2. READ COMMITTED
	// ReadCommitSELECT() // 模拟一个事务中两次SELECT结果不一样，即不可重复读
	// ReadCommitUPDATE() // 模拟一个事务中两次UPDATE，因为其他事务改变UPDATE条件，导致两次UPDATE未达到期望。（！！！存在无法解释的现象，不准确！！！！！）

	// 3. REPEATABLE READ mysql的默认级别
	// RepeatableReadNoPhantom() // 验证RepeatableRead级别不会出现幻读
	// RepeatableReadNoUnRepeatableRead() // 验证RepeatableRead级别不会出现可重复读

	// 4. SERIALIZABLE


}

func Lock() {

}

func Index() {

}

func Select() {

}
