package routers

import (
	"beegodemo.com/shFresh/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*", beego.BeforeExec, filterFunc)

    // beego.Router("/", &controllers.MainController{})
    // 注册
    beego.Router("/register", &controllers.UserController{}, "get:ShowReg;post:HandleReg")
    // 激活用户
	beego.Router("/active", &controllers.UserController{}, "get:ActiveUser")
	// 登录首页
    beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	// 跳转首页
	beego.Router("/", &controllers.GoodsController{}, "get:ShowIndex")
	// 退出登录
	beego.Router("/user/logout", &controllers.UserController{}, "get:Logout")
	// 用户中心信息页
	beego.Router("/user/userCenterInfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	// 用户中心-全部订单
	beego.Router("/user/userCenterOrder", &controllers.UserController{}, "get:ShowUserCenterOrder")
	// 用户中心-收货地址
	beego.Router("/user/userCenterSite", &controllers.UserController{}, "get:ShowUserCenterSite;post:HandleUserCenterSite")
	// 商品详情页
	beego.Router("/goodsDetail", &controllers.GoodsController{}, "get:ShowGoodsDetail")
	// 获取商品列表页
	beego.Router("/goodsList", &controllers.GoodsController{}, "get:ShowList")

}

var filterFunc = func(ctx *context.Context) {
	userName := ctx.Input.Session("userName")
	// 获取不到session的name返回登录页面
	if userName == nil {
		ctx.Redirect(302, "/login")
		return
	}


}