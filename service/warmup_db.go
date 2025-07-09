package service

import (
	"fmt"
	"time"

	"seanmcapp/repository"
)

type WarmupDBService interface {
	Run()
}

type WarmupDBServiceImpl struct {
	PeopleRepo repository.PeopleRepo
}

func (w *WarmupDBServiceImpl) Run() {
	fmt.Println("=== warmup database ===")
	day := time.Now().Day()
	month := int(time.Now().Month())

	res, err := w.PeopleRepo.Get(day, month)
	if err != nil {
		fmt.Println("Error warming up database:", err)
		return
	}
	fmt.Println("warmup result:", res)
}
