package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)

// Models
type (
	UserInfoModel struct {
		ID        int    `db:"id" json:"id"`
		Email     string `db:"email" json:"email"`
		FirstName string `db:"first_name" json:"firstName"`
		LastName  string `db:"last_name" json:"lastName"`
	}
	ResponseData struct {
		Users []UserInfoModel `json:"users"`
	}
)

func (u *UserInfoModel) GetTable() string {
	return "userinfo"
}

// Global Vars
var (
	seq        = 1
	dsn        string
	connection *dbr.Connection
	session    *dbr.Session
)

// Connect to MySQL
func connect() {
	// config/db.yamlにDB接続情報を設定する
	// 例)
	// host: 'testmaster.xxxxxxxxxxxx.ap-northeast-1.rds.amazonaws.com'
	// port: 3306
	// database: 'test'
	// user: 'dev'
	// password: 'password'
	//
	buf, err := ioutil.ReadFile("config/db.yaml")
	if err != nil {
		panic(err)
	}
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)
	if err != nil {
		panic(err)
	}
	fmt.Println("map:", m)
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		m["user"], m["password"], m["host"], m["port"], m["database"])
	fmt.Println("DataSourceName:", dsn)
	connection, _ = dbr.Open("mysql", dsn, nil)
	session = connection.NewSession(nil)
}

func insertUser(c echo.Context) error {
	fmt.Println("insertUser POST:/users/ + [JSON: Fields other than ID]")
	u := new(UserInfoModel)
	if err := c.Bind(u); err != nil {
		fmt.Println("err:", err)
		return err
	}
	fmt.Println("model:", u)
	session.InsertInto(u.GetTable()).Columns("id", "email", "first_name", "last_name").Values(u.ID, u.Email, u.FirstName, u.LastName).Exec()

	return c.NoContent(http.StatusOK)
}

func selectUsers(c echo.Context) error {
	fmt.Println("selectUsers GET:/users/")
	u := new(UserInfoModel)
	var ua []UserInfoModel

	session.Select("*").From(u.GetTable()).Load(&ua)
	response := new(ResponseData)
	response.Users = ua
	return c.JSON(http.StatusOK, response)
}
func selectUser(c echo.Context) error {
	fmt.Println("selectUser GET:/user/[id]")
	var u UserInfoModel
	id := c.Param("id")
	fmt.Println("id:", id)
	session.Select("*").From(u.GetTable()).Where("id = ?", id).Load(&u)
	return c.JSON(http.StatusOK, u)
}

func updateUser(c echo.Context) error {
	fmt.Println("updateUser PUT:/users/ + [JSON: ID and update fields]")
	u := new(UserInfoModel)
	if err := c.Bind(u); err != nil {
		return err
	}
	attrsMap := map[string]interface{}{"id": u.ID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName}
	session.Update(u.GetTable()).SetMap(attrsMap).Where("id = ?", u.ID).Exec()
	return c.NoContent(http.StatusOK)
}

func deleteUser(c echo.Context) error {
	fmt.Println("deleteUser DELETE:/users/[id]")
	u := new(UserInfoModel)
	id := c.Param("id")
	fmt.Println("id:", id)
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Println("err:", err)
	} else {
		session.DeleteFrom(u.GetTable()).
			Where("id = ?", id).
			Exec()
	}
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Database
	connect()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/users/", insertUser)
	e.GET("/user/:id", selectUser)
	e.GET("/users/", selectUsers)
	e.PUT("/users/", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":1234"))
}
