package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
	"ttsxProject/models"
)

type OrderController struct {
	beego.Controller
}

//处理结算订单请求
func (this *OrderController) ShowOrder() {
	userName := this.GetSession("userName")

	goodsIds := this.GetStrings("goodsId")
	if len(goodsIds) == 0 {
		beego.Error("请求路径错误.")
	}
	//连接redis服务器
	conn, err := redis.Dial("tcp", "192.168.201.129:6379")
	if err != nil {
		beego.Error("redis链接错误", err)
	}
	defer conn.Close()
	//连接数据库
	o := orm.NewOrm()
	var addrs []models.Address
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")
	o.QueryTable("Address").RelatedSel("User").Filter("User", user).All(&addrs)

	var totalPrice int //总金额
	var totalCount = 0 //商品总数
	var carriage = 10  //运费
	var orderInfo = make([]map[string]interface{}, 0)
	for _, value := range goodsIds {
		var orderGoods = make(map[string]interface{}, 0)
		goodsId, _ := strconv.Atoi(value)

		//查询商品数量
		resp, err := conn.Do("hget", "cart_"+strconv.Itoa(user.Id), goodsId)
		count, _ := redis.Int(resp, err)

		//查询商品信息
		var goodsSku models.GoodsSKU
		goodsSku.Id = goodsId
		o.Read(&goodsSku, "Id")

		//小计
		sumPrice := goodsSku.Price * count
		//总金额
		totalPrice += sumPrice
		//总商品数
		totalCount += 1

		orderGoods["goodsSku"] = goodsSku
		orderGoods["count"] = count
		orderGoods["sumPrice"] = sumPrice

		orderInfo = append(orderInfo, orderGoods)
	}

	this.Data["orderInfo"] = orderInfo
	this.Data["totalCount"] = totalCount
	this.Data["totalPrice"] = totalPrice
	this.Data["carriage"] = carriage
	this.Data["turePrice"] = totalPrice + carriage
	this.Data["addrs"] = addrs
	this.Data["goodsIds"] = goodsIds
	if userName == nil {
		this.Data["userName"] = ""
	} else {
		this.Data["userName"] = userName.(string)
	}
	this.TplName = "place_order.html"
}

//处理 ajax 提交订单请求
func (this *OrderController) AjaxOrderInfo() {
	re := make(map[string]interface{})

	addId, err1 := this.GetInt("addId")
	payId, err2 := this.GetInt("payId")
	ids := this.GetString("goodsIds")
	totalPrice, err3 := this.GetInt("totalPrice")
	totalCount, err4 := this.GetInt("totalCount")

	beego.Error(addId, payId, ids, totalPrice, totalCount)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || len(ids) == 0 {
		beego.Error("ajax获取数据失败")
		re["code"] = 1
		re["errmsg"] = "ajax获取数据失败"
		this.Data["json"] = re
		return
	}
	goodsIds := strings.Split(ids[1:len(ids)-1], " ")
	//beego.Error(addId, payId, goodsIds,ids, totalPrice, totalCount)

	o := orm.NewOrm()
	//获取用户表(user)数据
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")
	//获取地址表(Address)信息
	var addr models.Address
	addr.Id = addId
	o.Read(&addr)
	//向订单表(OrderInfo)插入数据
	var order models.OrderInfo
	order.OrderId = time.Now().Format("20060102150405") + strconv.Itoa(user.Id)
	order.User = &user
	order.Address = &addr
	order.PayMethod = payId
	order.TotalPrice = totalPrice
	order.TotalCount = totalCount
	order.TransitPrice = 10
	//**事务开始**//
	o.Begin()
	//插入操作
	o.Insert(&order)

	//插入数据到订单商品表
	conn, _ := redis.Dial("tcp", "192.168.201.129:6379")
	for _, value := range goodsIds {
		goodsId, _ := strconv.Atoi(value)
		for i := 0; i < 3; i++ {
			//获取商品信息
			var goodsSku models.GoodsSKU
			goodsSku.Id = goodsId
			o.Read(&goodsSku)
			//获取商品数量
			resp, err := conn.Do("hget", "cart_"+strconv.Itoa(user.Id), goodsId)
			count, _ := redis.Int(resp, err)
			//向订单商品表(OrderGoods)插入数据
			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goodsSku
			orderGoods.Price = goodsSku.Price * count
			orderGoods.OrderInfo = &order
			orderGoods.Count = count
			//插入操作
			o.Insert(&orderGoods)

			//更新商品SKU(goodsSKU)表 的库存和销量
			if count > goodsSku.Stock {
				beego.Error("商品库存不足,订单提交失败。")
				re["code"] = 2
				re["errmsg"] = "商品库存不足,订单提交失败。"
				this.Data["json"] = re
				//**事务回滚**//
				o.Rollback()
				return
			}

			preStock:=goodsSku.Stock
			//goodsSku.Sales += count
			//o.Update(&goodsSku)

			_, err = o.QueryTable("GoodsSKU").Filter("Id", goodsSku.Id).Filter("Stock", preStock).Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})
			if err != nil {

				beego.Error("库存不足")
				re["code"] = 2
				re["errmsg"] = "商品库存不足，订单提交失败"
				this.Data["json"] = re
				//**事务回滚**//
				o.Rollback()
				continue
			} else {
				break
			}
		}
		//订单提交成功 删除购物车信息
		conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), goodsId)
	}
	re["code"] = 5
	re["errmsg"] = "OK"
	this.Data["json"] = re
	defer this.ServeJSON()
	//**事务结束**//
	o.Commit()
}
