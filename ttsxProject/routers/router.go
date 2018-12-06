package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"ttsxProject/controllers"
	_ "ttsxProject/models"
)

func init() {
	//路由过滤器
	beego.InsertFilter("/ttsx/*", beego.BeforeExec, funcFilter)

	//主页
	beego.Router("/", &controllers.GoodsController{}, "get:ShowIdex")
	//注册
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	//邮箱验证
	beego.Router("/active", &controllers.UserController{}, "get:HandleActive")
	//登陆
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//退出登陆
	beego.Router("/ttsx/logout", &controllers.UserController{}, "get:Logout")
	//用户中心信息页
	beego.Router("/ttsx/userCenterInfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	//用户中心订单信息页
	beego.Router("/ttsx/userCenterOrder", &controllers.UserController{}, "get:ShowUserCenterOrder")
	//用户中心地址信息页
	beego.Router("/ttsx/userCenterSite", &controllers.UserController{}, "get:ShowUserCenterSite;post:HandleUserCenterSite")
	//展示商品详情页
	beego.Router("goodsDetail", &controllers.GoodsController{}, "get:ShowGoodsDetail")
	//展示商品列表页
	beego.Router("goodsList", &controllers.GoodsController{}, "get:ShowGoodsList")
	//搜索商品
	beego.Router("/searchGoods", &controllers.GoodsController{}, "post:HandleSearch")
	//抽奖页面
	beego.Router("/ball", &controllers.GoodsController{}, "get:ShowBall")
	//添加购物车
	beego.Router("/ttsx/addCart", &controllers.CartController{}, "post:AjaxAddCart")
	//购物车页面
	beego.Router("/ttsx/showCart", &controllers.CartController{}, "get:ShowCart")
	//删除购物车商品
	beego.Router("/ttsx/deleteCart", &controllers.CartController{}, "get:DeleteCart;post:AjaxDeleteCart")
	//结算订单
	beego.Router("/ttsx/showOrder", &controllers.OrderController{}, "post:ShowOrder")
	//提交订单
	beego.Router("/ttsx/orderInfo", &controllers.OrderController{}, "post:AjaxOrderInfo")
	//支付宝付款
	beego.Router("/user/PayAli",&controllers.UserController{},"get:PayAli")
	//短信服务
	beego.Router("/sms",&controllers.UserController{},"get:SMS")
}

var funcFilter = func(ctx *context.Context) {
	//登陆校验
	userName := ctx.Input.Session("userName")
	if userName == nil {
		ctx.Redirect(302, "/login")
	}
}
