package routers

import (
	"beegodemo.com/shFresh/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/register", &controllers.UserController{}, "get:ShowReg;post:HandleReg")
    beego.Router("/active", &controllers.UserController{}, "get:ActiveUser")

    beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")

}
