package service

import (
	"database/sql"
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/global"
	"time"
)

func ReadCommitSELECT()  {
	// 事务1
	go func() {
		var money int
		tx:= global.GpgDb.Begin()
		tx.Raw("SELECT money from accounts where name = ?", "person1").Scan(&money)
		println(money)

		time.Sleep(5*time.Second)

		tx.Raw("SELECT money from accounts where name = ?", "person1").Scan(&money)
		println(money)
		tx.Commit()
	}()

	// 事务2
	go func() {
		time.Sleep(1*time.Second)
		tx:= global.GpgDb.Begin()
		tx.Exec("UPDATE accounts SET money = ? where name = ?", 100, "person1" )
		tx.Commit()
	}()
	// main 协程等待,否则打不出信息
	time.Sleep(10*time.Second)
}

// ReadCommitUPDATE
// 事务1 对 money+100 再 +100
// 事务2 将money置为1
// 现象：期望是事务1中最后money只有101，然而真正的现象是事务2的执行被卡住，类似于事务上锁？
func ReadCommitUPDATE(){
	// 事务1
	go func() {
		var money int
		tx:= global.GpgDb.Begin()
		// +100
		tx.Exec("UPDATE accounts SET money = money + 100 where name = ?", "person1")

		// 模拟一些耗时操作
		time.Sleep(10*time.Second)

		// 再+100
		tx.Exec("UPDATE accounts SET money = money + 100 where name = ?", "person1")

		// 按照道理，money应改为201
		tx.Raw("SELECT money from accounts where name = ?", "person1").Scan(&money)
		println(money)
		tx.Commit()
	}()

	// 事务2
	go func() {
		time.Sleep(100*time.Millisecond)
		tx:= global.GpgDb.Begin()
		tx.Exec("UPDATE accounts SET money = ? where name = ? ", 1 ,"person1" )
		tx.Commit()
	}()
	// main 协程等待,否则打不出信息
	time.Sleep(20*time.Second)

}

func RepeatableReadNoPhantom()  {

	go func() {
		var count int
		tx:= global.GpgDb.Begin(&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
			ReadOnly: false,
		})
		tx.Raw("SELECT count(*) from accounts").Scan(&count)
		println(count)

		time.Sleep(5*time.Second)

		tx.Raw("SELECT count(*) from accounts").Scan(&count)
		println(count)
		tx.Commit()
	}()

	go func() {
		time.Sleep(1*time.Second)
		tx:= global.GpgDb.Begin()
		tx.Exec("INSERT into accounts (name,money) values (?,?)",  "person101",100 )
		tx.Commit()
	}()

	time.Sleep(10*time.Second)

}

func RepeatableReadNoUnRepeatableRead()  {
	// 事务1
	go func() {
		var money int
		tx:= global.GpgDb.Begin(&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
			ReadOnly: false,
		})
		tx.Raw("SELECT money from accounts where name = ?", "person1").Scan(&money)
		println(money)

		time.Sleep(5*time.Second)

		tx.Raw("SELECT money from accounts where name = ?", "person1").Scan(&money)
		println(money)
		tx.Commit()
	}()

	// 事务2
	go func() {
		time.Sleep(1*time.Second)
		tx:= global.GpgDb.Begin()
		tx.Exec("UPDATE accounts SET money = ? where name = ?", 100, "person1" )
		tx.Commit()
	}()
	// main 协程等待,否则打不出信息
	time.Sleep(10*time.Second)
}
