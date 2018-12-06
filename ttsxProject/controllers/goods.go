package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"
	"ttsxProject/models"
	_ "ttsxProject/models"
)

type GoodsController struct {
	beego.Controller
}

//展示主页
func (this *GoodsController) ShowIdex() {
	//获取用户名和购物车数量
	userName, cartCount := NameCarCount(this)
	this.Data["CartCount"] = cartCount
	this.Data["userName"] = userName

	o := orm.NewOrm()
	var goodstypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodstypes)

	//查詢
	var indexGoodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanners)

	//
	var indexPromotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexPromotionBanners)
	//查詢首頁展示商品
	var goodsSkus = make([]map[string]interface{}, len(goodstypes))

	//把類型對象放入我們的map容器中
	for index, _ := range goodsSkus {
		temp := make(map[string]interface{})
		temp["types"] = goodstypes[index]
		goodsSkus[index] = temp
	}
	//存商品數據
	for _, goodsMap := range goodsSkus {
		var goodsImage []models.IndexTypeGoodsBanner
		var goodsText []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSku").Filter("GoodsType", goodsMap["types"]).Filter("DisplayType", 0).All(&goodsText)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSku").Filter("GoodsType", goodsMap["types"]).Filter("DisplayType", 1).All(&goodsImage)

		goodsMap["goodsImage"] = goodsImage
		goodsMap["goodsText"] = goodsText
	}

	this.Data["goodsSkus"] = goodsSkus
	this.Data["goodsTypes"] = goodstypes
	this.Data["indexGoodsBanners"] = indexGoodsBanners
	this.Data["indexPromotionBanners"] = indexPromotionBanners
	this.TplName = "index.html"
}

//展示商品详情页
func (this *GoodsController) ShowGoodsDetail() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("请求路径错误")
	}
	//处理数据
	//查询
	o := orm.NewOrm()
	//获取查询对象
	var goodsSku models.GoodsSKU
	//给查询条件赋值
	goodsSku.Id = id
	//查询
	err = o.QueryTable("GoodsSKU").RelatedSel("Goods", "GoodsType").Filter("Id", id).One(&goodsSku)
	if err != nil {
		beego.Error("查询商品数据错误")
	}
	//goodsSku.GoodsType.Id

	//添加历史浏览记录
	//判断是否是登录状体
	userName := this.GetSession("userName")
	if userName != nil {
		//需要获取存储的信息  用户id 和商品id
		//1.获取当前用户信息
		var user models.User
		user.Name = userName.(string)
		o.Read(&user, "Name")

		//存储
		//链接，获取redis操作对象
		conn, err := redis.Dial("tcp", "192.168.201.129:6379")
		if err != nil {
			beego.Error("redis链接失败")
		}
		defer conn.Close()
		conn.Do("lrem", "history_"+strconv.Itoa(user.Id), 0, id)
		conn.Do("lpush", "history_"+strconv.Itoa(user.Id), id)

	}

	//返回数据
	this.Data["goodsSku"] = goodsSku
	GoodsLayout(this, goodsSku.GoodsType.Id)
	this.Data["title"] = "天天生鲜-商品详情"
	this.TplName = "detail.html"
}

//展示商品列表页
func (this *GoodsController) ShowGoodsList() {
	//获取数据
	typeId, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取商品类型错误")
	}
	//处理数据
	//获取和传递过来的类型一直的商品
	o := orm.NewOrm()
	//获取查询对象
	var goodsSkus []models.GoodsSKU
	//查询  默认排序
	//o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",typeId).All(&goodsSkus)
	//this.Data["goodsSkus"] = goodsSkus
	//按照相应的排序方式获取数据
	sort := this.GetString("sort")

	if sort == "price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Price").All(&goodsSkus)
	} else if sort == "sale" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Sales").All(&goodsSkus)
	} else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).All(&goodsSkus)
	}

	//实现分页
	//分析讨论，现在应该展示哪些页码
	//获取查询到总个数
	count, _ := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).Count()
	//每页显示个数
	pageSize := 3
	//总页数
	pageCount := math.Ceil(float64(count) / float64(pageSize))
	//当前所在页面页码
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//调用函数
	pages := PageEdior(pageCount, pageIndex)

	start := (pageIndex - 1) * pageSize
	if sort == "price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Price").Limit(pageSize, start).All(&goodsSkus)
	} else if sort == "sale" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Sales").Limit(pageSize, start).All(&goodsSkus)
	} else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).Limit(pageSize, start).All(&goodsSkus)
	}
	prePage := pageIndex - 1
	if prePage < 1 {
		prePage = 1
	}
	nextPage := pageIndex + 1
	if nextPage > int(pageCount) {
		nextPage = int(pageCount)
	}

	//查找列表类型类型
	var goodsType models.GoodsType
	o.QueryTable("GoodsType").Filter("Id", typeId).All(&goodsType)
	//返回数据
	GoodsLayout(this, typeId)
	this.Data["sort"] = sort
	this.Data["GoodsType"] = goodsType
	this.Data["pages"] = pages
	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage
	this.Data["pageIndex"] = pageIndex
	this.Data["goodsSkus"] = goodsSkus
	this.Data["typeId"] = typeId
	this.Data["title"] = "天天生鲜-商品列表"
	this.TplName = "list.html"
}

//搜索商品
func (this *GoodsController) HandleSearch() {
	search := this.GetString("searchName")
	o := orm.NewOrm()
	var goodsSkus []models.GoodsSKU
	if search == "" {
		o.QueryTable("GoodsSKU").All(&goodsSkus)
	} else {
		o.QueryTable("GoodsSKU").Filter("Name__icontains", search).All(&goodsSkus)
	}

	//返回数据
	this.Data["search"] = goodsSkus
	this.Layout = "goods_layout.html"
	this.TplName = "search.html"
}

//抽奖页面
func (this *GoodsController) ShowBall() {
	this.TplName = "双色球.html"
}

//展示Layout页面
func GoodsLayout(this *GoodsController, typeId int) {
	//获取用户名和购物车数量
	userName, cartCount := NameCarCount(this)
	this.Data["CartCount"] = cartCount
	this.Data["userName"] = userName

	//获取类型数据
	//查询类型
	o := orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes
	//获取新品数据
	//获取同一类型的新品数据
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Time").Limit(2, 0).All(&newGoods)
	this.Data["newGoods"] = newGoods
	this.Layout = "goods_layout.html"
}

//实现分页
func PageEdior(pageCount float64, pageIndex int) []int {
	//判断显示哪些页码
	var pages []int
	if pageCount <= 5 {
		pages = make([]int, int(pageCount))
		for i := 0; i < int(pageCount); i++ {
			pages[i] = i + 1
		}
		//i := 1
		//for pageCount > 0 {
		//	pages[i-1] = i
		//	pageCount -= 1
		//	i += 1
		//}
	} else if pageIndex <= 3 {
		pages = make([]int, 5)
		pages = []int{1, 2, 3, 4, 5}
		//当前页码等于
		//for i:=1;i<=5 ;i++  {
		//	pages[i-1]=i
		//}

		//i := 1
		//var temp = 5
		//for temp > 0 {
		//	pages[i-1] = i
		//	temp -= 1
		//	i += 1
		//}
	} else if pageIndex >= int(pageCount)-2 {
		pages = make([]int, 5)
		//给后三页赋值
		for i := 0; i < 5; i++ {
			pages[i] = int(pageCount) - 4 + i
		}
		//temp := 5
		//i := 1
		//for temp > 0 {
		//	pages[i-1] = int(pageCount) - temp + 1
		//	temp -= 1
		//	i += 1
		//}
	} else {
		pages = make([]int, 5)
		for i := 0; i < 5; i++ {
			pages[i] = pageIndex - 2 + i
		}
		//temp := 2
		//i := 1
		//for temp > -3 {
		//	pages[i-1] = pageIndex - temp
		//	temp --
		//	i ++
		//}
	}
	return pages
}

//传递用户名与购物车数量
func NameCarCount(this *GoodsController) (string, string) {
	//当登录的时候显示欢迎你，username,并显示购物车数量,当没有登录显示登录注册
	userName := this.GetSession("userName")
	var CartCount string
	var UserName string
	if userName == nil {
		CartCount = "登陆"
		userName = ""
	} else {
		o := orm.NewOrm()
		var user models.User
		user.Name = userName.(string)
		o.Read(&user, "Name")
		//连接redis服务器
		conn, err := redis.Dial("tcp", "192.168.201.129:6379")
		if err != nil {
			beego.Error("redis链接错误", err)
		}
		defer conn.Close()
		//查询购物车数量   获取redis有多少属性
		num, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
		carCount, _ := redis.Int(num, err)

		UserName=userName.(string)
		CartCount = strconv.Itoa(carCount)
	}
	//beego.Info("UserName=",UserName, "CartCount=",CartCount)
	return UserName, CartCount
}
