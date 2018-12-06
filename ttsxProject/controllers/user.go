package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/KenmyZhang/aliyun-communicate"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/alipay"
	"math"
	"regexp"
	"strconv"
	"ttsxProject/models"
	_ "ttsxProject/models"
)

type UserController struct {
	beego.Controller
}

//获取用户名函数 并 使用视图模板
func ShowLayout(this *UserController) {
	//获取userName
	userName := this.GetSession("userName")
	this.Data["userName"] = userName.(string)

	this.Layout = "user_center_layout.html"
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//处理注册数据
func (this *UserController) HandleRegister() {
	//获取数据
	userName := this.GetString("user_name")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")
	//校验数据
	if userName == "" || pwd == "" || cpwd == "" || email == "" {
		this.Data["errmsg"] = "数据不能为空，请重新注册!"
		this.TplName = "register.html"
		return
	}
	//邮箱格式校验
	reg, _ := regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res := reg.FindString(email)
	if res == "" {
		this.Data["errmsg"] = "邮箱格式不正确，请重新注册!"
		this.TplName = "register.html"
		return
	}
	//两次输入密码是否一直
	if pwd != cpwd {
		this.Data["errmsg"] = "两次密码输入不正确，请重新注册！"
		this.TplName = "register.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	//获取插入对象
	var user models.User
	//给插入对象赋值
	user.Name = userName
	user.PassWord = pwd
	user.Email = email
	_, err := o.Insert(&user)
	if err != nil {
		this.Data["errmsg"] = "注册失败，请重新注册！"
		this.TplName = "register.html"
		return
	}

	//邮箱激活
	emailConfig := `{"username":"915440609@qq.com","password":"kxylzrtexgflbeaj","host":"smtp.qq.com","port":587}`
	ems := utils.NewEMail(emailConfig)
	ems.From = "915440609@qq.com"
	ems.To = []string{email}
	ems.Subject = "天天生鲜用户验证"
	ems.HTML = "<a href=\"http://120.0.0.1:8080/active?id=" + strconv.Itoa(user.Id) + "\">点击该链接，天天生鲜用户激活</a>"

	err = ems.Send()
	beego.Error(err)

	//返回数据
	this.Redirect("/login", 302)
	//this.Ctx.WriteString("注册成功！")

}

//处理激活业务
func (this *UserController) HandleActive() {
	//获取数据
	id, err := this.GetInt("id")
	if err != nil {
		this.Data["errmsg"] = "激活失败，清重新注册"
		this.TplName = "register.html"
		return
	}
	//查询数据
	o := orm.NewOrm()
	var user models.User
	user.Id = id
	err = o.Read(&user)
	if err != nil {
		this.Data["errmsg"] = "激活失败，清重新注册"
		this.TplName = "register.html"
		return
	}
	//更新字段
	user.Active = true
	_, err = o.Update(&user)
	if err != nil {
		this.Data["errmsg"] = "激活失败，清重新注册"
		this.TplName = "register.html"
		return
	}

	//返回数据
	this.Redirect("/login", 302)
}

//展示登陆页面
func (this *UserController) ShowLogin() {
	dec := this.Ctx.GetCookie("userName")
	userName, _ := base64.StdEncoding.DecodeString(dec)
	if string(userName) != "" {
		this.Data["userName"] = string(userName)
		this.Data["checked"] = "checked"
	} else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}

	this.TplName = "login.html"
}

//处理登陆业务
func (this *UserController) HandleLogin() {
	userName := this.GetString("username")
	passWord := this.GetString("pwd")
	if userName == "" || passWord == "" {
		this.Data["errmsg"] = "用户名或密码为空"
		this.TplName = "login.html"
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		this.Data["errmsg"] = "用户名不存在"
		this.TplName = "login.html"
		return
	}
	if user.PassWord != passWord {
		this.Data["errmsg"] = "密码错误，请重新输入"
		this.TplName = "register.html"
		return
	}
	if !user.Active {
		this.Data["errmsg"] = "邮箱没有验证，请验证后登陆"
		this.TplName = "register.html"
		return
	}

	//获取是否记住用户名
	remember := this.GetString("check")
	enc := base64.StdEncoding.EncodeToString([]byte(userName))
	if remember == "on" {
		this.Ctx.SetCookie("userName", enc, 3600*1)
	} else {
		this.Ctx.SetCookie("userName", userName, -1)
	}

	//设置Session
	this.SetSession("userName", userName)

	//返回数据
	this.Redirect("/", 302)
}

//处理退出登陆业务
func (this *UserController) Logout() {
	//删除session
	this.DelSession("userName")
	//跳转登陆界面
	this.Redirect("/login", 302)
}

//展示用户中心信息页
func (this *UserController) ShowUserCenterInfo() {
	//获取用户名函数 并 使用视图模板
	ShowLayout(this)
	//查询当前用户的默认地址
	userName := this.GetSession("userName").(string)
	o := orm.NewOrm()
	var address models.Address
	o.QueryTable("Address").Filter("Isdefault", true).RelatedSel("User").Filter("User__Name", userName).One(&address)
	/*//获取当前用户的默认地址
	o := orm.NewOrm()
	//指定要查询的表
	qs := o.QueryTable("Address")
	//关联用户表
	qs = qs.RelatedSel("User")
	//判断当前用户
	userName := this.GetSession("userName")
	qs = qs.Filter("User__Name",userName.(string))
	//获取默认地址
	qs = qs.Filter("Isdefault",true)
	//把查询到的数据，放到容器里面
	var address models.Address
	err := qs.One(&address)
	if err != nil{
		this.Data["address"] = ""
	}else {
		this.Data["address"] = address
	}*/
	conn, err := redis.Dial("tcp", "192.168.201.129:6379")
	if err != nil {
		beego.Error("redis链接失败", err)
	}
	defer conn.Close()

	//userName := this.GetSession("userName").(string)
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")
	resp, err := conn.Do("lrange", "history_"+strconv.Itoa(user.Id), 0, 4)
	//回复助手函数
	goodsId, err := redis.Ints(resp, err)
	if err != nil {
		beego.Error("redis获取商品错误", err)
	}
	var goodsSkus []models.GoodsSKU
	for _, id := range goodsId {
		var goods models.GoodsSKU
		goods.Id = id
		o.Read(&goods)
		goodsSkus = append(goodsSkus, goods)
	}

	this.Data["goodsSkus"] = goodsSkus
	this.Data["address"] = address
	this.TplName = "user_center_info.html"
}

//展示全部订单信息页
func (this *UserController) ShowUserCenterOrder() {
	o := orm.NewOrm()
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")

	//实现分页
	count, _ := o.QueryTable("OrderInfo").RelatedSel("User").Filter("User", user).Count()
	pageSize := 2   //每页显示个数
	//总页数
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//当前所在页面页码
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	var n = 5 //一个屏幕显示页码个数
	var pages []int
	if pageCount <= n {
		pages = make([]int, int(pageCount))
		for i := 0; i < int(pageCount); i++ {
			pages[i] = i + 1
		}
	} else if pageIndex <= n/2+1 {
		pages = make([]int, n)
		for i := 0; i < n; i++ {
			pages[i] = i + 1
		}
	} else if pageIndex >= int(pageCount)-n/2 {
		pages = make([]int, n)
		//给后三页赋值
		for i := 0; i < n; i++ {
			pages[i] = int(pageCount) - 4 + i
		}

	} else {
		pages = make([]int, n)
		for i := 0; i < n; i++ {
			pages[i] = pageIndex - 2 + i
		}
	}
	//上一页
	prePage := pageIndex - 1
	if prePage < 1 {
		prePage = 1
	}
	//下一页
	nextPage := pageIndex + 1
	if nextPage > int(pageCount) {
		nextPage = int(pageCount)
	}

	//分页后显示订单信息
	order := make([]map[string]interface{}, 0)
	start := (pageIndex - 1) * pageSize

	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User", user).Limit(pageSize, start).All(&orderInfos)

	for _, orderInfo := range orderInfos {
		temp := make(map[string]interface{})

		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo", "GoodsSKU").Filter("OrderInfo", orderInfo).All(&orderGoods)

		temp["oInfo"] = orderInfo
		temp["oGoods"] = orderGoods

		order = append(order, temp)
	}

	this.Data["order"] = order
	this.Data["pages"] = pages
	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage
	this.Data["pageIndex"] = pageIndex
	//获取用户名函数 并 使用视图模板
	ShowLayout(this)
	this.TplName = "user_center_order.html"
}

//展示收获地址信息页
func (this *UserController) ShowUserCenterSite() {
	//获取用户名函数 并 使用视图模板
	ShowLayout(this)
	//查询当前用户的默认地址
	userName := this.GetSession("userName").(string)
	o := orm.NewOrm()
	var address models.Address
	o.QueryTable("Address").Filter("Isdefault", true).RelatedSel("User").Filter("User__Name", userName).One(&address)

	this.Data["address"] = address
	this.TplName = "user_center_site.html"
}

//处理收获地址业务
func (this *UserController) HandleUserCenterSite() {
	//获取页面信息
	addRessee := this.GetString("addRessee")
	addr := this.GetString("addr")
	zipCode := this.GetString("zipCode")
	phone := this.GetString("phone")
	userName := this.GetSession("userName").(string)
	//校验获取信息
	if phone == "" || zipCode == "" || addr == "" || addRessee == "" {
		this.Data["errmsg"] = "内容有空，请重新输入！"
		this.Redirect("/ttsx/userCenterSite", 302)
		return
	}
	////数据操作--一对多插入////
	o := orm.NewOrm()

	var user models.User
	user.Name = userName

	o.Read(&user, "Name")

	var address models.Address
	address.Receiver = addRessee
	address.Addr = addr
	address.Zipcode = zipCode
	address.Phone = phone
	address.User = &user
	/*判断当前用户是否有默认地址，如果没有,则直接插入默认地址，
	如果有默认地址，把默认地址更新为非默认地址，把新插入的地址设置为默认地址
	*/
	var oldAddress models.Address
	err := o.QueryTable("Address").Filter("Isdefault", true).RelatedSel("User").Filter("User__Id", user.Id).One(&oldAddress)
	if err != nil {
		address.Isdefault = true
	} else {
		oldAddress.Isdefault = false
		o.Update(&oldAddress)
		address.Isdefault = true
	}

	_, err = o.Insert(&address)
	if err != nil {
		this.Ctx.WriteString("服务器遭受陨石攻击，暂时无法服务")
	}

	//返回数据
	this.Redirect("/ttsx/userCenterSite", 302)
}

//支付宝付款
func (this *UserController) PayAli() {
	var privateKey = ""

	var appId = "2016092200569649"
	var aliPublicKey = ""

	var client = alipay.New(appId, aliPublicKey, privateKey, false)

	//alipay.trade.page.pay
	var p = alipay.AliPayTradePagePay{}
	p.NotifyURL = "http://192.168.110.81:8080/user/payOk"
	p.ReturnURL = "http://192.168.110.81:8080/user/payOk"
	p.Subject = "天天生鲜"
	p.OutTradeNo = "987654321"
	p.TotalAmount = "1000.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()

	this.Redirect(payURL, 302)
}

//短信服务
func (this *UserController) SMS() {
	var (
		gatewayUrl      = "http://dysmsapi.aliyuncs.com/"
		accessKeyId     = "LTAIQ9aVPA8IEwCg"
		accessKeySecret = "EFwkulaxYhp4gFDP9IY4rvUVvf8NE0"
		phoneNumbers    = "18339985087"             //要发送的电话号码
		signName        = "天天生鲜"                    //签名名称
		templateCode    = "SMS_149101793"           //模板号
		templateParam   = "{\"code\":\"bj2qttsx\"}" //验证码
	)

	smsClient := aliyunsmsclient.New(gatewayUrl)
	result, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, templateCode, templateParam)
	fmt.Println("Got raw response from server:", string(result.RawResponse))
	if err != nil {
		beego.Info("配置有问题")
	}

	if result.IsSuccessful() {
		this.Data["result"] = "短信已经发送"
	} else {
		this.Data["result"] = "短信发送失败"
	}
	//this.TplName = "SMS.html"
}
