package main

import (
	"net/http"
	"strconv"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
)


type (
	UserInfoModel struct {
		ID        int    `db:"id" json:"id"`
		Email     string `db:"email" json:"email"`
		FirstName string `db:"first_name" json:"firstName"`
		LastName  string `db:"last_name" json:"lastName"`
	}
	responseData struct {
		Users []UserInfoModel `json:"users"`
	}
)
func (u *UserInfoModel) GetTable() string {
	return "userinfo"
}

var (
	seq           = 1
	dsn           string
	connection    *dbr.Connection
	session       *dbr.Session
)

func connectionect() {
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
	dsn =  fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		m["user"], m["password"], m["host"], m["port"], m["database"] )
	fmt.Println("DataSourceName:", dsn)
	connection, _ = dbr.Open("mysql", dsn, nil)
	session = connection.NewSession(nil)
}

func insertUser(c echo.Context) error {
	fmt.Println("insertUser")
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
	u := new(UserInfoModel)
    var ua []UserInfoModel

    session.Select("*").From(u.GetTable()).Load(&ua)
    response := new(responseData)
    response.Users = ua
    return c.JSON(http.StatusOK,response)
}
func selectUser(c echo.Context) error {
    var u UserInfoModel
    id := c.Param("id")
	fmt.Println("id:", id)
    session.Select("*").From(u.GetTable()).Where("id = ?", id).Load(&u)
    //idはPrimary Key属性のため重複はありえない。そのため一件のみ取得できる。そのため配列である必要はない
    return c.JSON(http.StatusOK, u)
}


func updateUser(c echo.Context) error {
    u := new(UserInfoModel)
    if err := c.Bind(u); err != nil {
        return err
    }
    attrsMap := map[string]interface{}{"id": u.ID, "email": u.Email , "first_name" : u.FirstName , "last_name" : u.LastName}
    session.Update(u.GetTable()).SetMap(attrsMap).Where("id = ?", u.ID).Exec()
    return c.NoContent(http.StatusOK)
}

func deleteUser(c echo.Context) error {
    u := new(UserInfoModel)
    id,_ := strconv.Atoi(c.Param("id"))

    session.DeleteFrom(u.GetTable()).
    Where("id = ?", id).
    Exec()
    return c.NoContent(http.StatusNoContent)
}


func main() {
    e := echo.New()

	// Database
	connectionect()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Routes
    e.POST("/users/", insertUser)
    e.GET("/user/:id", selectUser)
    e.GET("/users/",selectUsers)
    e.PUT("/users/", updateUser)
    e.DELETE("/users/:id", deleteUser)

    // Start server
	e.Logger.Fatal(e.Start(":1234"))
}

