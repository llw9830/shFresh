package controllers

import (
	"github.com/astaxie/beego"
	"strconv"
	"github.com/astaxie/beego/orm"
	"beegodemo.com/shFresh/models"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"github.com/smartwalle/alipay"
	"fmt"
)

type OrderController struct {
	beego.Controller
}

// 展示订单详情页面
func (this *OrderController) ShowOrder () {
	// 获取数据
	skuids := this.GetStrings("skuid")
	// 校验数据
	if len(skuids) == 0 {
		beego.Info("确切数据错误！")
		this.Redirect("/user/cart", 302)
		return
	}


	// 处理数据
	o := orm.NewOrm()
	userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	totalPrice := 0
	totalCount := 0
	goodsBuffer := make([]map[string]interface{}, len(skuids))
	for i, skuid := range skuids {
		temp := make(map[string]interface{})

		// 商品id
		id, _ := strconv.Atoi(skuid)
		var goodsSKU models.GoodsSKU
		goodsSKU.Id = id
		o.Read(&goodsSKU)
		temp["goods"] = goodsSKU
		// redius获取商品数量
		conn, _:= redis.Dial("tcp", "127.0.0.1:6379")
		defer conn.Close()
		count, _ := redis.Int(conn.Do("hget", "cart_" + strconv.Itoa(user.Id), id))
		temp["count"] = count
		// 小计
		amount := count * goodsSKU.Price
		temp["amount"] = amount

		// 计算总金额和总件数
		totalCount += count
		totalPrice += amount

		goodsBuffer[i] = temp
	}

	// 传到商品数据
	this.Data["goodsBuffer"] = goodsBuffer
	// 传递计算总金额和总件数
	this.Data["totalCount"] = totalCount
	this.Data["totalPrice"] = totalPrice
	transferPrice := 10
	this.Data["transferPrice"] = transferPrice
	realPrice := totalPrice + transferPrice

	this.Data["realPrice"] = realPrice



	// 地址数据
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Id", user.Id).All(&addrs)
	this.Data["addrs"] = addrs

	this.Data["skuids"] = skuids

	// 返回视图
	this.TplName = "place_order.html"
}


// 添加订单
func (this *OrderController) AddOrder () {
	// 获取数据
	addrId, _ := this.GetInt("addrId")
	payId, _ := this.GetInt("payId")
	skuid := this.GetString("skuids") // 这里拿到的是一个类似[1, 2, 3]的字符串
	skuid = skuid[1: len(skuid)-1] // 去掉首位的[]
	skuids := strings.Split(skuid, " ") // 用空格分割


	totalCount, _ := this.GetInt("totalCount")
	//totalPrice, _ := this.GetInt("totalPrice")
	transferPrice, _ := this.GetInt("transferPrice")
	realPrice, _ := this.GetInt("realPrice")


	resp := make(map[string]interface{})
	defer this.ServeJSON()

	// 校验数据
	if len(skuids) == 0 {
		beego.Info("获取数据错误！")
		resp["code"] = 1
		resp["errmsg"] = "数据库连接错误！"
		this.Data["json"] = resp
		return
	}

	// 处理数据
	// 1、向订单表插入数据
	o := orm.NewOrm()

	// 数据库事务开始
	o.Begin()

	userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var order models.OrderInfo
	order.OrderId = time.Now().Format("2006010215030405") + strconv.Itoa(user.Id)
	order.User = &user
	order.Orderstatus = 1 // 1未支付
	order.PayMethod = payId
	order.TotalCount = totalCount
	order.TotalPrice = realPrice
	order.TransitPrice = transferPrice

	var addr models.Address
	addr.Id = addrId
	o.Read(&addr)

	order.Address = &addr

	// 执行插入操作
	o.Insert(&order)


	// 1、向订单商品表插入数据
	conn, _:= redis.Dial("tcp", "127.0.0.1:6379")
	defer conn.Close()
	for _, skuId := range skuids {
		id, _ := strconv.Atoi(skuId)

		var goods models.GoodsSKU
		goods.Id = id

		// 循环三次
		i := 3
		for i > 0 {
			o.Read(&goods)

			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goods
			orderGoods.OrderInfo = &order

			// 获取商品数量
			count, _ :=	redis.Int(conn.Do("hget", "cart_" + strconv.Itoa(user.Id), id))

			//
			if count > goods.Stock {
				resp["code"] = 2
				resp["errmsg"] = "商品库存不足！"
				this.Data["json"] = resp
				// 事务回滚
				o.Rollback()
				return
			}

			// 原来的库存,提交更新数据库时与这个相比
			preCount := goods.Stock

			//time.Sleep(5 * time.Second)

			orderGoods.Count = count

			orderGoods.Price = count * goods.Price

			o.Insert(&orderGoods)

			//减少库存  增加销量
			goods.Stock -= count
			goods.Sales += count

			// 通过原来的商品库存查询更新，如果库存与现在不一致，回退数据库操作。
			updateCount, _ := o.QueryTable("GoodsSKU").Filter("Id", goods.Id).Filter("Stock", preCount).
				Update(orm.Params{"Stock": goods.Stock, "Sales": goods.Sales})
			if updateCount == 0 {
				if i > 0 {
					i -= 1
					continue
				}
				resp["code"] = 3
				resp["errmsg"] = "商品库存改变，提交订单失败！"
				this.Data["json"] = resp
				// 事务回滚
				o.Rollback()
				return
			} else {
				// 更新redis
				conn.Do("hdel", "cart_" + strconv.Itoa(user.Id), goods.Id)
				break
			}
		}
	}

	// 事务提交
	o.Commit()

	// 返回数据
	resp["code"] = 5
	resp["errmsg"] = "OK"
	this.Data["json"] = resp
}


// 处理支付
func (this *OrderController) HandlePay () {
	var aliPublicKey = "" // 可选，支付宝提供给我们用于签名验证的公钥，通过支付宝管理后台获取
	var privateKey =""// 必须，上一步中使用 RSA签名验签工具 生成的私钥
	var client, err1 = alipay.New("", aliPublicKey, privateKey, false)

	// 将 key 的验证调整到初始化阶段
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	orderId := this.GetString("orderId")
	totalPrice := this.GetString("totalPrice")

	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://xxx"
	p.ReturnURL = "http://192.168.31.213:8080/user/payok"
	p.Subject = "天天生鲜购物平台"
	p.OutTradeNo = orderId
	p.TotalAmount = totalPrice
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()
	this.Redirect(payURL, 302)
	// 这个 payURL 即是用于支付的 URL，可将输出的内容复制，到浏览器中访问该 URL 即可打开支付页面。
}


// 支付成功
func (this *OrderController) PayOk () {
	orderId := this.GetString("out_trade_no")
	// 校验
	if orderId == ""{
		beego.Info("支付返回错误！")
		this.Redirect("/user/userCenterOrder", 302)
		return
	}

	// 更新数据库
	o := orm.NewOrm()
	// 更新支付状态
	count, _ := o.QueryTable("OrderInfo").Filter("OrderId", orderId).Update(orm.Params{"Orderstatus": 2})
	if count == 0 {
		beego.Info("更新数据错误！")
		this.Redirect("/user/userCenterOrder", 302)
		return
	}

	this.Redirect("/user/userCenterOrder", 302)
}