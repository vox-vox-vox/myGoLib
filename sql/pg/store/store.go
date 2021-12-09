package store

import (
	"gitlab.momenta.works/hdmap-workflow/hx_golib/myGoLib/sql/pg/global"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Name  string
	Money int
}

func create(account Account) {
	global.GpgDb.Create(&account)
}

func BatchCreate(accounts []Account) error {
	if err := global.GpgDb.Create(&accounts).Error; err != nil {
		return err
	}
	return nil
}

func get() {

}

func save() {

}

func list() {

}
