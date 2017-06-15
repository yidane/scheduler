package controller

import (
	"github.com/yidane/scheduler/job"
	"github.com/yidane/scheduler/common"
)

type MonitorController struct  {
	 BaseController

}

func (this *MonitorController)Index()  {

	jobManger := job.NewJobMnager()
	jobList,err := jobManger.GetJobSnapshotList()
	if err != nil {
		common.PanicIf(err)

	}

	this.TplName = "monitor/index.html"
	this.Data["jobList"] = jobList
	this.Render()
}


