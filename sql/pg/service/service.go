package service

import (
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/global"
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/store"
	"strconv"
)

func Preload() {
	err := global.GpgDb.Migrator().DropTable(&store.Account{})
	if err != nil {
		return
	}
	err = global.GpgDb.Migrator().CreateTable(&store.Account{})
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

	// 2. READ COMMITTED pg的默认级别
	/**
	Read Committed is the default isolation level in PostgreSQL.
	When a transaction uses this isolation level,
	1. SELECT query (without a FOR UPDATE/SHARE clause) sees only data committed before the query began;
	   In effect, a SELECT query sees a snapshot of the database as of the instant the query begins to run.
	2. SELECT does see the effects of previous updates executed within its own transaction, even though they are not yet committed.
	3. two successive SELECT commands can see different data, even though they are within a single transaction,
	   if other transactions commit changes after the first SELECT starts and before the second SELECT starts.
	*/

	ReadCommitSELECT() // 模拟一个事务中两次SELECT结果不一样
	//ReadCommitUPDATE() // 模拟一个事务中两次UPDATE，因为其他事务改变UPDATE条件，导致两次UPDATE未达到期望。（！！！存在无法解释的现象，不准确！！！！！）

	// 3. REPEATABLE READ
	/**
	When a transaction uses this isolation level,
	1. The Repeatable Read isolation level only sees data committed before the transaction began;
	   it never sees either uncommitted data or changes committed during transaction execution by concurrent transactions.
	2. the query does see the effects of previous updates executed within its own transaction, even though they are not yet committed.
	*/

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
