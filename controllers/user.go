package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"beegodemo.com/shFresh/models"
	"log"
	"github.com/astaxie/beego/utils"
	"strconv"
	"encoding/base64"
)

type UserController struct {
	beego.Controller
}

// 显示注册页面
func (this *UserController) ShowReg () {
	this.TplName = "register.html"
}

// 处理注册数据
func (this *UserController) HandleReg () {
	// 1、 获取数据
	user_name := this.GetString("user_name")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")
	// 2、 校验数据
	if user_name == "" || pwd == "" || cpwd == "" || email == "" {
		this.Data["errmsg"] = "数据不完整，请重新注册！"
		this.TplName = "register.html"
		return
	}
	if pwd != cpwd {
		this.Data["errmsg"] = "两次输入密码不一致，请重新注册！"
		this.TplName = "register.html"
		return
	}
	// 验证邮箱格式
	reg, _ := regexp.Compile(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`)
	res := reg.FindString(email)
	if res == "" {
		log.Println("邮箱格式不正确，请重新注册！")
		this.Data["errmsg"] = "邮箱格式不正确，请重新注册！"
		this.TplName = "register.html"
		return
	}
	// 3、 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = user_name
	user.PassWord = pwd
	user.Email = email
	_, err := o.Insert(&user)
	if err != nil{
		log.Println("注册邮箱格式不正确！")
		this.Data["errmsg"] = "注册失败，请更换数据注册！"
		this.TplName = "register.html"
		return
	}

	// 发送邮件
	emailConfig := `{"username":"llw9830@foxmail.com","password":"***********","host":"smtp.qq.com","port":587}`
	emailConn := utils.NewEMail(emailConfig)
	emailConn.From = "llw9830@foxmail.com"
	//emailConn.To = []string{email}
	emailConn.To = []string{email,}
	emailConn.Subject = "天天生鲜用户注册！"
	// 这里发的是激活地址
	emailConn.Text = "192.168.1.109:8080/active?id=" + strconv.Itoa(user.Id)
	//emailConn.Send()
	err = emailConn.Send()
	if err != nil{
		log.Printf("发送邮件失败：%v", err)
		this.Data["errmsg"] = "发送激活邮件失败，请重新注册！"
		this.TplName = "register.html"
		return
	}

	// 4、 返回视图
	this.Ctx.WriteString("注册成功，请去相应邮箱激活！")

}

// 激活用户
func (this *UserController) ActiveUser() {
	// 获取数据
	id, err := this.GetInt("id")
	if err != nil{
		this.Data["errmsg"] = "要激活的用户不存在！"
		this.TplName = "register.html"
		return
	}

	// 数据处理
	// 更新操作
	o := orm.NewOrm()
	var user models.User
	user.Id = id
	err = o.Read(&user)
	if err != nil{
		this.Data["errmsg"] = "要激活的用户不存在！"
		this.TplName = "register.html"
		return
	}
	user.Active = true
	o.Update(&user)

	// 返回视图
	this.Redirect("/login", 302)


}

// 展示登录页面
func (this *UserController) ShowLogin () {
	uname := this.Ctx.GetCookie("userName")
	temp, _  := base64.StdEncoding.DecodeString(uname)
	if string(temp) == "" {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	} else {
		this.Data["userName"] = string(temp)
		this.Data["checked"] = "checked"
	}

	this.TplName = "login.html"
}

// 处理登录业务
func (this *UserController) HandleLogin () {
	// 获取数据
	username := this.GetString("username")
	pwd := this.GetString("pwd")

	// 校验数据
	if username == "" || pwd == "" {
		this.Data["errmsg"] = "登录数据不完整，请重新输入！"
		this.TplName = "login.html"
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = username
	err := o.Read(&user, "Name")
	if err != nil {
		this.Data["errmsg"] = ">用户名或密码错误，请重新输入！"
		this.TplName = "login.html"
		return
	}

	if user.PassWord != pwd {
		this.Data["errmsg"] = "用户名或>密码错误，请重新输入！"
		this.TplName = "login.html"
		return
	}
	if user.Active != true {
		this.Data["errmsg"] = "用户未激活，请在邮箱激活！"
		this.TplName = "login.html"
		return
	}

	// 4、 返回视图
	// 记住密码操作
	remember :=  this.GetString("remember")
	if remember == "on" {
		// cookie加密
		tempuname  := base64.StdEncoding.EncodeToString([]byte(username))
		this.Ctx.SetCookie("userName", tempuname, 24 * 3600 * 30)
	} else {
		this.Ctx.SetCookie("userName", "", -1)
	}

	this.SetSession("userName", username)
	//this.Ctx.WriteString("登录成功！")
	this.Redirect("/", 302)
}

// 退出登录
func (this *UserController) Logout () {
	this.DelSession("userName")
	// 跳转
	this.Redirect("/login", 302)
}

// 展示用户中心页面
func (this *UserController) ShowUserCenterInfo (){
	userName := GetUser(&this.Controller)
	this.Data["userName"] = userName
	// 查询其他内容
	// 获取地址表
	o := orm.NewOrm()
	// 高级查询 表关联
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("Receiver", userName).Filter("Isdefault", true).One(&addr)
	this.Data["addr"] = addr
	// 传递参数
	if addr.Id == 0 {
		this.Data["addr"] = ""
	}else {
		this.Data["addr"] = addr
	}

	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_info.html"
}

// 展示用户中心订单页
func (this *UserController)  ShowUserCenterOrder () {
	GetUser( &this.Controller)

	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_order.html"
}

// 用户中心-收货地址页面展示
func (this *UserController) ShowUserCenterSite () {
	GetUser(&this.Controller)
	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_site.html"
}

// 添加收获地址
func (this *UserController) HandleUserCenterSite () {
	GetUser(&this.Controller)
	//this.Layout = "userCenterLayout.html"
	//this.TplName = "user_center_site.html"

	// 获取数据
	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	zipCode := this.GetString("zipCode")
	phone := this.GetString("phone")
	// 校验数据
	if receiver==""||addr==""||zipCode==""||phone==""{
		beego.Info("数据填写不完整！")
		this.Redirect("/user/userCenterSite", 302)
		return
	}

	o := orm.NewOrm()
	var addrUser models.Address

	addrUser.Isdefault = true
	err := o.Read(&addrUser, "Isdefault")
	// 如果原来有默认地址将老的默认地址改为非默认
	if err == nil {
		addrUser.Isdefault = false
		o.Update(&addrUser)
	}

	// 更新地址时如果原来的覅在对象赋值了,还用原来的地址对象插入,也就是用原来的id插入会报错
	// 关联的user表
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var addrUserNew models.Address
	addrUserNew.Receiver = receiver
	addrUserNew.Addr = addr
	addrUserNew.Zipcode = zipCode
	addrUserNew.Phone = phone
	addrUserNew.Isdefault = true
	addrUserNew.User = &user
	o.Insert(&addrUserNew)


	// 返回视图
	this.Redirect("/user/userCenterSite", 302)

}