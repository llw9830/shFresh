package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"beegodemo.com/shFresh/models"
)

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

// 传递类型数据和username
func showLayout (this*beego.Controller) {
	var types []*models.GoodsType
	o := orm.NewOrm()
	o.QueryTable("GoodsType").All(&types)
	this.Data["types"] = types
	GetUser(this)
	this.Layout = "goodsLayout.html"
}


// 商品首页展示
func (this *GoodsController) ShowIndex ()  {
	GetUser(&this.Controller)
	o := orm.NewOrm()
	// --------------------获取类型数据
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	// ------------------获取轮播图片数据
	var indexGoodsBanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanner)
	this.Data["indexGoodsBanner"] = indexGoodsBanner

	// ----------------------获取促销商品数据
	var promotionBanner []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanner)
	this.Data["promotionBanner"] = promotionBanner

	// ---------------------获取展示商品数据
	goods := make([]map[string]interface{}, len(goodsTypes))
	// 向切片interface中插入类型数据
	for index, value := range goodsTypes {
		// 获取对应商品类型的首页展示商品
		temp := make(map[string]interface{})
		temp["type"] = value
		goods[index] = temp
	}
	// 商品数据
	for _, value := range goods {
		var textGoods []models.IndexTypeGoodsBanner
		var imgGoods []models.IndexTypeGoodsBanner
		// 获取文字商品数据
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").
			OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 0).All(&textGoods)
		// 获取图片商品数据
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").
			OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 1).All(&imgGoods)

		value["textGoods"] = textGoods
		value["imgGoods"] = imgGoods

		//beego.Info(textGoods)
	}

	this.Data["goods"] = goods
	this.TplName = "index.html"

}

// 展示商品详情
func (this *GoodsController) ShowGoodsDetail () {
	// 获取数据
	id, err := this.GetInt("id")

	// 校验数据
	if err != nil{
		beego.Error("浏览器请求错误！")
		this.Redirect("/", 302)
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var goodSKU models.GoodsSKU
	goodSKU.Id = id
	//o.Read(&goodSKU)
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType", "Goods").Filter("Id", id).One(&goodSKU)

	// 获取同类型靠前的两条商品
	var goodsNew []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType", goodSKU.GoodsType).
		OrderBy("Time").Limit(2, 0).All(&goodsNew)
	this.Data["goodsNew"] = goodsNew

	// 返回数据
	this.Data["goodSKU"] = goodSKU

	showLayout(&this.Controller)
	this.TplName = "detail.html"
}