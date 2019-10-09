package sys

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/syyongx/php2go"
	"go-cms/controllers"
	"go-cms/models"
	"go-cms/pkg/e"
	"go-cms/pkg/util"
	"go-cms/services"
	"go-cms/validations/backend"
	"log"
	"strings"
)

type UserController struct {
	controllers.BaseController
}

func (c *UserController) Prepare() {

}

func (c *UserController) Index() {
	if c.Ctx.Input.IsAjax() {
		page, _ := c.GetInt("page")
		limit, _ := c.GetInt("limit")
		key := c.GetString("key", "")

		result, count := models.NewUser().Pagination((page-1)*limit, limit, key)
		c.JsonResult(e.SUCCESS, "ok", result, count)
	}
}
func (c *UserController) UserInfo() {
	token := c.Ctx.Input.Header(beego.AppConfig.String("jwt::token_name"))
	kv := strings.Split(token, " ")
	uid := util.GetUserIdByToken(kv[1])
	userInfo, err := models.NewUser().FindById(uid)
	if err != nil {
		c.JsonResult(e.ERROR, e.ResponseMap[e.ERROR])
	}
	c.JsonResult(e.SUCCESS, "success", userInfo)
}
func (c *UserController) Create() {
	if c.Ctx.Input.IsPost() {
		userModel := models.NewUser()
		//1.压入数据
		if err := c.ParseForm(userModel); err != nil {
			c.JsonResult(e.ERROR, "赋值失败")
		}

		salt := util.Krand(5, 2)
		userModel.Salt = salt
		userModel.LoginName = userModel.UserName
		userModel.Email = userModel.Email
		userModel.Password = php2go.Md5(userModel.Password + salt)
		userModel.CreatedAt = php2go.Time()
		userModel.UpdatedAt = php2go.Time()

		//2.验证
		valid := validation.Validation{}
		b, err := valid.Valid(userModel)
		if err != nil {
			log.Printf("%v\n%v", err, valid.Errors)
			c.JsonResult(e.ERROR, "验证失败")
		}
		if !b {
			for _, err := range valid.Errors {
				log.Println(err.Key, err.Message)
			}
		}

		//3.插入数据
		if _, err := userModel.Create(); err != nil {
			c.JsonResult(e.ERROR, "创建失败")
		}
		c.JsonResult(e.SUCCESS, "添加成功")
	}
}

func (c *UserController) Update() {
	id, _ := c.GetInt("id")
	user, _ := models.NewUser().FindById(id)

	if c.Ctx.Input.IsPost() {
		//1
		if err := c.ParseForm(&user); err != nil {
			c.JsonResult(e.ERROR, "赋值失败")
		}
		//2
		valid := validation.Validation{}
		if b, _ := valid.Valid(user); !b {
			for _, err := range valid.Errors {
				log.Println(err.Key, err.Message)
			}
			c.JsonResult(e.ERROR, "验证失败")
		}
		//3
		if _, err := user.Update(); err != nil {
			c.JsonResult(e.ERROR, "修改失败")
		}
		c.JsonResult(e.SUCCESS, "修改成功")
	}

	c.Data["vo"] = user
	c.TplName = c.ADMIN_TPL + "user/update.html"
}

func (c *UserController) Delete() {
	userModel := models.NewUser()
	id, _ := c.GetInt("id")
	userModel.Id = id
	if err := userModel.Delete(); err != nil {
		c.JsonResult(e.ERROR, "删除失败")
	}
	c.JsonResult(e.SUCCESS, "删除成功")
}

func (c *UserController) BatchDelete() {
	var ids []int
	if err := c.Ctx.Input.Bind(&ids, "ids"); err != nil {
		c.JsonResult(e.ERROR, "赋值失败")
	}

	userModel := models.NewUser()
	if err := userModel.DelBatch(ids); err != nil {
		c.JsonResult(e.ERROR, "删除失败")
	}
	c.JsonResult(e.SUCCESS, "删除成功")
}

//用户登录
func (c *UserController) Login() {
	if c.Ctx.Input.IsPost() {
		user_name := c.GetString("user_name")
		password := c.GetString("password")

		//数据校验
		loginData := backend.UserLoginValidation{}
		loginData.UserName = user_name
		loginData.Password = password
		c.ValidatorAuto(&loginData)

		//通过service查询
		user := services.FindByUserName(user_name)

		if php2go.Empty(user) {
			c.JsonResult(e.ERROR, "User Not Exist")
		}

		has := php2go.Md5(password + user.Salt)

		if (user.Password == has) {
			token := util.CreateToken(user)
			jsonData := make(map[string]interface{}, 1)
			jsonData["token"] = token
			c.JsonResult(e.SUCCESS, "登录成功!", jsonData)
		}

		c.JsonResult(e.ERROR, has)
	}
}

func (c *UserController) CheckToken() {

	token := c.Ctx.Input.Header("Authorization")

	b, _ := util.CheckToken(token)

	if !b {
		c.JsonResult(e.ERROR, "验证失败!")
	}

	c.JsonResult(e.SUCCESS, "success")
}
