package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"beegodemo.com/shFresh/models"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"math"
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
func ShowLayout (this*beego.Controller) {
	var goodsTypes []*models.GoodsType
	o := orm.NewOrm()
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes
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

	// 添加历史浏览记录
	// 判断用户是否登录
	uname := this.GetSession("userName")
	if uname != nil{
		// 查询用户信息
		o := orm.NewOrm()
		var user models.User
		user.Name = uname.(string)
		o.Read(&user, "Name")

		// 添加历史记录, redis
		conn, err := redis.Dial("tcp", "127.0.0.1:6379")
		defer conn.Close()
		if err != nil {
			beego.Info("redis链接错误！")
		}
		// 先删除数据
		conn.Do("lrem", "history_" + strconv.Itoa(user.Id), 0, id)
		// 插入数据
		conn.Do("lpush", "history_" + strconv.Itoa(user.Id), id)


	}



	ShowLayout(&this.Controller)
	this.Data["cartCount"] = GetCartCount(&this.Controller)
	this.TplName = "detail.html"
}

// 分页
func PageTool(pageCount, pageIndex int) []int {
	// 这里默认显示5页下标
	pages := make([]int, pageCount)
	// 总页数小于5时
	if pageCount <= 5 {
		for i, _ := range pages {
			pages[i] = i+1
		}
		//  总页数大于5且当前下标效于3
	} else if pageIndex <= 3 {
		pages = []int{1, 2, 3, 4, 5}
		// 当前页在最后三页中时显示最后5页
	} else if pageIndex > pageCount - 3 {
		pages = []int{pageCount-4, pageCount-3, pageCount-2, pageCount-1, pageCount}
	} else {
		// 页数在中间时
		pages = []int {pageIndex-2, pageIndex-1, pageIndex, pageIndex+1, pageIndex+2}
	}
	return pages
}

// 展示商品列表页
func (this *GoodsController)  ShowList () {
	// 获取数据
	id, err := this.GetInt("typeId")
	// 校验数据
	if err != nil{
		beego.Info("请求路径错误！")
		this.Redirect("/", 302)
		return
	}

	// 处理数据
	ShowLayout(&this.Controller)
	// 获取新品
	o := orm.NewOrm()
	var goodsNew []models.GoodsSKU
	// 拿到该类型下的新品数据， 按时间排序前两条
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
		Filter("GoodsType__Id", id).OrderBy("Time").Limit(2, 0).All(&goodsNew)
	this.Data["goodsNew"] = goodsNew

	// 获取商品
	var goods []models.GoodsSKU
	/*o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
		Filter("GoodsType__Id", id).All(&goods)
	this.Data["goods"] = goods*/

	// 分页
	// 获取pageCount
	// 总页数
	count, _ := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
		Filter("GoodsType__Id", id).Count()
	// 每页数量
	pageSize := 3
	// 总页数
	pageCount := math.Ceil(float64(count)/float64(pageSize))
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil{
		//beego.Info("获取pageIndex错误：%v.", err)
		pageIndex = 1
	}
	// 分页
	//pageCount = 3
	pages := PageTool(int(pageCount), pageIndex)
	this.Data["pages"] = pages // 显示也页面数
	this.Data["typeId"] = id // 类型ID
	this.Data["pageIndex"] = pageIndex
	// 上一页
	preIndex := pageIndex - 1
	if preIndex <= 1 {
		preIndex = 1
	}
	this.Data["preIndex"] = preIndex
	// 下一页
	nextIndex := pageIndex + 1
	if nextIndex >= int(pageCount) {
		nextIndex = int(pageCount)
	}
	this.Data["nextIndex"] = nextIndex

	// 返回当前页数据
	start := pageSize * (pageIndex - 1)


	// 按一定顺序获取商品
	sort := this.GetString("sort")
	if sort == "" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
			Filter("GoodsType__Id", id).Limit(pageSize, start).All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = ""
	}else if sort == "price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
			Filter("GoodsType__Id", id).OrderBy("Price").Limit(pageSize, start).All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = "price"
	}else if sort == "sale" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
			Filter("GoodsType__Id", id).OrderBy("Sales").Limit(pageSize, start).All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = "sale"
	}

	// 返回数据
	this.TplName = "list.html"
}


// 搜索
func (this *GoodsController) HandleSearch () {
	// 获取数据
	goodsName := this.GetString("goodsName")

	o := orm.NewOrm()
	var goods []models.GoodsSKU
	// 校验数据
	if goodsName == "" {
		o.QueryTable("GoodsSKU").All(&goods)
		this.Data["goods"] = goods
		ShowLayout(&this.Controller)
		this.TplName = "search.html"
		return
	}
	// 处理数据
	o.QueryTable("GoodsSKU").Filter("Name__icontains", goodsName).All(&goods)
	this.Data["goods"] = goods
	ShowLayout(&this.Controller)
	this.TplName = "search.html"
}