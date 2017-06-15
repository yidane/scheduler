package main

import (
	"log"
	"runtime"
	"strconv"

	"strings"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yidane/scheduler/controller"
	"github.com/yidane/scheduler/entity"
	"github.com/yidane/scheduler/job"
)

func init() {
	mysqluser := strings.TrimSpace(beego.AppConfig.String("mysqluser"))
	mysqlpass := strings.TrimSpace(beego.AppConfig.String("mysqlpass"))
	mysqlurls := strings.TrimSpace(beego.AppConfig.String("mysqlurls"))
	mysqlport := strings.TrimSpace(beego.AppConfig.String("mysqlport"))
	mysqldb := strings.TrimSpace(beego.AppConfig.String("mysqldb"))
	if len(mysqldb) == 0 || len(mysqlpass) == 0 || len(mysqlurls) == 0 || len(mysqluser) == 0 || len(mysqlport) == 0 {
		log.Println("mysql配置不合法")
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
