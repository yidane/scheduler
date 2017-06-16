package main

import (
	"log"
	"runtime"
	"strconv"

	"strings"

	"fmt"

	"errors"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yidane/scheduler/controller"
	"github.com/yidane/scheduler/entity"
	"github.com/yidane/scheduler/job"
)

func init() {
	checkConfig := func(config, msg string, err *error) string {
		if *err != nil {
			return ""
		}
		str := strings.TrimSpace(config)
		if len(str) == 0 {
			*err = errors.New(msg)
			return ""
		}
		return str
	}
	var err error
	mysqluser := checkConfig(beego.AppConfig.String("mysqluser"), "数据库访问账户不可为空", &err)
	mysqlpass := checkConfig(beego.AppConfig.String("mysqlpass"), "数据库访问密码不可为空", &err)
	mysqlurls := checkConfig(beego.AppConfig.String("mysqlurls"), "数据库访问链接不可为空", &err)
	mysqlport := checkConfig(beego.AppConfig.String("mysqlport"), "数据库访问端口不可为空", &err)
	mysqldb := checkConfig(beego.AppConfig.String("mysqldb"), "数据库名称不可为空", &err)

	if err != nil {
		log.Println("mysql配置不合法:", err)
		return
	}
	port, err := strconv.Atoi(mysqlport)
	if err != nil {
		fmt.Println(err)
		return
	}
	orm.RegisterDriver("mysql", orm.DRMySQL)
	err = orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8&loc=Local",
		mysqluser,
		mysqlpass,
		mysqlurls,
		port,
		mysqldb))
	if err != nil {
		log.Println(err)
		return
	}
	orm.SetMaxIdleConns("default", 30)
	orm.SetMaxOpenConns("default", 30)
	orm.RegisterModel(&entity.JobInfo{}, &entity.JobInfoHistory{}, &entity.JobSnapshot{})
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		log.Println(err)
	}
}
func main() {
	// set CPU
	runtime.GOMAXPROCS(runtime.NumCPU())
	orm.Debug = true
	jobManager := job.NewJobMnager()
	jobManager.PushAllJob()
	// TODO Init jobList

	// set home  path
	beego.Router("/", &controller.IndexController{}, "get:Index")

	// jobinfo
	beego.Router("/jobinfo/list", &controller.JobInfoManagerController{}, "*:List")
	beego.Router("/jobinfo/add", &controller.JobInfoManagerController{}, "get:ToAdd")
	beego.Router("/jobinfo/add", &controller.JobInfoManagerController{}, "post:Add")
	beego.Router("/jobinfo/edit", &controller.JobInfoManagerController{}, "get:ToEdit")
	beego.Router("/jobinfo/edit", &controller.JobInfoManagerController{}, "post:Edit")
	beego.Router("/jobinfo/delete", &controller.JobInfoManagerController{}, "post:Delete")
	beego.Router("/jobinfo/info", &controller.JobInfoManagerController{}, "get:Info")
	beego.Router("/jobinfo/active", &controller.JobInfoManagerController{}, "*:Active")
	// jobsnapshot
	beego.Router("/jobsnapshot/list", &controller.JobSnapshotController{}, "*:List")
	beego.Router("/jobsnapshot/info", &controller.JobSnapshotController{}, "get:Info")

	// jobinfohistory
	beego.Router("/jobinfohistory/list", &controller.JobInfoHistoryController{}, "*:List")

	//about
	beego.Router("/about", &controller.AboutController{}, "*:Index")

	//monitor
	beego.Router("/monitor/", &controller.MonitorController{}, "*:Index")

	// set static resource
	beego.SetStaticPath("static", "static")
	beego.SetStaticPath("public", "static")

	// start web app
	beego.Run()
}
