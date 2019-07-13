package controllers

import "github.com/astaxie/beego"

type GoodsController struct {
	beego.Controller
}

func GetUser(this *beego.Controller) string {
	// 获取登录信息
	userName := this.GetSession("userName")
	if userName == nil {
		this.Data["userName"] = ""
	} else {
		this.Data["userName"] = userName
		return userName.(string)
	}

	return ""
}

// 商品首页
func (this *GoodsController) ShowIndex ()  {
	GetUser(&this.Controller)

	this.TplName = "index.html"

}