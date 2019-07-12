package main

import (
	_ "beegodemo.com/shFresh/routers"
	"github.com/astaxie/beego"
	_ "beegodemo.com/shFresh/models"
)

func main() {
	beego.Run()
}

