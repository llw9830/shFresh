package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"beegodemo.com/shFresh/models"
	"strconv"
)

type CartController struct {
	beego.Controller
}

// 添加购物车
func (this *CartController) HandleAddCart() {
	// 获取数据
	skuid, err1 := this.GetInt("skuid")
	count, err2 := this.GetInt("count")
	resp := make(map[string]interface{})
	defer 	this.ServeJSON()

	// 校验数据
	if err1 != nil || err2 != nil {
		resp["code"] = 1
		resp["msg"] = "传递的数据不正确"
		this.Data["json"] = resp
		return
	}
	userName := this.GetSession("userName")
	if userName == nil {
		resp["code"] = 1
		resp["msg"] = "未登录状态！"
		this.Data["json"] = resp
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")


	// 处理数据
	// 购物车数据存储redis，使用hash
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		beego.Info("redis 连接错误！")
		return
	}

	// 先获取原来的数量，然后数量加起来
	preCount, err := redis.Int(conn.Do("hget", "cart_"+strconv.Itoa(user.Id)))
	conn.Do("hset", "cart_"+strconv.Itoa(user.Id), skuid, count + preCount)

	rep, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	// 回复助手函数
	cartCount, _ := redis.Int(rep, err)

	resp["code"] = 5
	resp["msg"] = "OK"
	resp["cartCount"] = cartCount


	this.Data["json"] = resp

	// 返回json数据
}

// 获取购物车数量函数
func GetCartCount(this*beego.Controller) int {
	// 从redis中获取
	userName := this.GetSession("userName")
	if userName == nil {
		return 0
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		return 0
	}
	defer conn.Close()

	rep, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	// 回复助手函数
	cartCount, _ := redis.Int(rep, err)

	return cartCount

}


// 展示购物车页面
func (this *CartController) ShowCart () {
	// 用户信息
	userName := this.GetSession("userName")
	this.Data["userName"] = userName

	// redis获取数据
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		beego.Info("redis连接失败！")
		return
	}
	defer conn.Close()

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")
	goodsMap, _ := redis.IntMap(conn.Do("hgetall", "cart_" + strconv.Itoa(user.Id)) )// 返回map[string]int

	goods := make([]map[string]interface{}, len(goodsMap))
	index := 0
	totalPrice := 0
	totalCount := 0
	for i, v := range goodsMap {
		id,_ := strconv.Atoi(i)
		var goodsSKU models.GoodsSKU
		goodsSKU.Id = id
		o.Read(&goodsSKU)

		temp := make(map[string]interface{})
		temp["goodsSKU"] = goodsSKU
		temp["count"] = v

		totalCount += v
		totalPrice += goodsSKU.Price * v

		temp["addPrice"] = goodsSKU.Price * v
		goods[index] = temp
		index++
	}

	this.Data["goods"] = goods
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount

	this.TplName = "cart.html"
}


// 更新购物车数量
func (this *CartController) HandleUpdateCart () {
	skuid, err1 := this.GetInt("skuid")
	count, err2 := this.GetInt("count")
	resp := make(map[string]interface{})
	defer this.ServeJSON() // 发送数据

	// 校验数据
	if err1 != nil || err2 != nil {
		resp["code"] = 1
		resp["errmsg"] = "请求数据不正确"
		this.Data["json"] = resp
		return
	}


	userName := this.GetSession("userName")
	if userName == nil {
		resp["code"] = 3
		resp["errmsg"] = "用户未登录！"
		this.Data["json"] = resp
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")



	// 处理数据
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "redis连接不正确"
		this.Data["json"] = resp
		return
	}
	defer  conn.Close()

	conn.Do("hset", "cart_" + strconv.Itoa(user.Id), skuid, count)

	resp["code"] = 5
	resp["errmsg"] = "OK"
	this.Data["json"] = resp


}


// 删除购物车
func (this *CartController) DeleteCart () {
		// 获取数据
	skuid, err := this.GetInt("skuid")

	resp := make(map[string] interface{})
	defer this.ServeJSON()
	// 校验数据
	if err != nil{
		resp["errcode"] = 1
		resp["errmsg"] = "确切数据不正确！"

		this.Data["json"] = resp
		return
	}

	// 处理数据
	// 处理数据
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	defer  conn.Close()
	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "redis连接不正确"
		this.Data["json"] = resp
		return
	}

	userName := this.GetSession("userName")
	o := orm.NewOrm()
	user := models.User{}
	user.Name = userName.(string)
	o.Read(&user, "Name")
	conn.Do("hdel", "cart_" + strconv.Itoa(user.Id), skuid)

	// 返回数据
	resp["code"] = 5
	resp["errmsg"] = "OK"
	this.Data["json"] = resp

}