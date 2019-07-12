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

	this.Ctx.WriteString("登录成功！")
}
