package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"ttsxProject/models"
)

type CartController struct {
	beego.Controller
}

//处理 ajax 添加购物车请求
func (this *CartController) AjaxAddCart() {
	resp := make(map[string]interface{})
	defer this.ServeJSON()

	goodsId, err1 := this.GetInt("goodsId")
	count, err2 := this.GetInt("count")
	temp := this.GetString("temp")
	//beego.Error(temp)
	if err1 != nil || err2 != nil {
		beego.Error("ajax获取数据失败", err1, "22", err2)
		resp["code"] = 1
		resp["errmsg"] = "ajax获取数据失败"
		this.Data["json"] = resp
		return
	}

	userName := this.GetSession("userName").(string)
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "192.168.201.129:6379")
	if err != nil {
		beego.Error("redis链接错误", err)
		resp["code"] = 2
		resp["errmsg"] = "redis链接错误"
		this.Data["json"] = resp
		return
	}
	defer conn.Close()
	if temp == "" {
		res, err := conn.Do("hget", "cart_"+strconv.Itoa(user.Id), goodsId)
		preCount, _ := redis.Int(res, err)
		conn.Do("hset", "cart_"+strconv.Itoa(user.Id), goodsId, preCount+count)
	} else {
		conn.Do("hset", "cart_"+strconv.Itoa(user.Id), goodsId, count)
	}

	num, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	carCount, _ := redis.Int(num, err)

	resp["code"] = 5
	resp["count"] = carCount
	this.Data["json"] = resp

}

//展示购物车页面
func (this *CartController) ShowCart() {
	//获取用户名 并 查询用户信息
	userName := this.GetSession("userName").(string)
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	//连接redis服务器
	conn, err := redis.Dial("tcp", "192.168.201.129:6379")
	if err != nil {
		beego.Error("redis链接错误", err)
		return
	}
	defer conn.Close()
	//查询所有信息
	resp, err := conn.Do("hgetall", "cart_"+strconv.Itoa(user.Id))

	goodsMap, _ := redis.IntMap(resp, err)

	var goods = make([]map[string]interface{}, 0)
	for goodsId, count := range goodsMap {
		id, _ := strconv.Atoi(goodsId)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		cliMap := make(map[string]interface{})
		cliMap["goodsSku"] = goodsSku
		cliMap["count"] = count
		cliMap["sumPrice"] = goodsSku.Price * count

		goods = append(goods, cliMap)
	}
	len := len(goods)

	this.Data["CartCount"] = len
	this.Data["goods"] = goods
	this.Data["userName"] = userName
	this.TplName = "cart.html"
}

//处理删除购物车商品请求
//<a>标签 路由控制器 后台实现
func (this *CartController) DeleteCart() {
	goodsId, err := this.GetInt("goodsId")
	if err != nil {
		beego.Error("ajax获取数据失败，无效商品id", err)
		return
	}
	userName := this.GetSession("userName").(string)
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "192.168.201.129:6379")
	if err != nil {
		beego.Error("redis链接错误", err)
		return
	}
	defer conn.Close()

	_, err = conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), goodsId)
	if err != nil {
		beego.Error("删除商品失败", err)
		return
	}
	this.Redirect("/ttsx/showCart",302)
}
//ajax实现	前台实现
func (this *CartController) AjaxDeleteCart() {
		resp := make(map[string]interface{})
		resp["code"] = 5
		this.Data["json"] = resp

		defer this.ServeJSON()

		goodsId, err := this.GetInt("goodsId")
		if err != nil {
			beego.Error("ajax获取数据失败", err)
			resp["code"] = 1
			resp["errmsg"] = "ajax获取数据失败，无效商品id"
			this.Data["json"] = resp
			return
		}
		userName := this.GetSession("userName").(string)
		o := orm.NewOrm()
		var user models.User
		user.Name = userName
		o.Read(&user, "Name")

		conn, err := redis.Dial("tcp", "192.168.201.129:6379")
		if err != nil {
			beego.Error("redis链接错误", err)
			resp["code"] = 2
			resp["errmsg"] = "redis链接错误"
			this.Data["json"] = resp
			return
		}
		defer conn.Close()

		_, err = conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), goodsId)
		if err != nil {
			beego.Error("删除商品失败", err)
			resp["code"] = 3
			resp["errmsg"] = "删除商品失败"
			this.Data["json"] = resp
			return
		}
}