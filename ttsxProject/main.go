package main

import (
	"github.com/astaxie/beego"
	_ "ttsxProject/models"
	_ "ttsxProject/routers"
)

func main() {
	beego.AddFuncMap("ChangeStock", ChangeStock)
	beego.AddFuncMap("CalStock", CalStock)
	beego.AddFuncMap("AddIndex", AddIndex)
	beego.Run()
}

func AddIndex(Index int) int {
	return Index + 1
}

func ChangeStock(Stock int) int {
	return Stock - 1
}

func CalStock(Stock,count int)int{
	return Stock - count
}