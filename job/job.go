package job

import (
	"errors"
	"log"

	"github.com/yidane/scheduler/common"
	"github.com/yidane/scheduler/entity"
	invoke "github.com/yidane/scheduler/invoker"
	"github.com/yidane/scheduler/quartz"

	"time"
)

var (
	Error_quartz_null = errors.New("the quartz is null")
)

var jm *JobManager
var QZ *quartz.Quartz

type JobManager struct {
	qz *quartz.Quartz
}

func NewJobMnager() *JobManager {
	if jm == nil {
		QZ = quartz.New()
		QZ.BootStrap()
		jm = &JobManager{qz: QZ}
	}

	return jm
}

func (this *JobManager) PushAllJob() {

	jobInfo := &entity.JobInfo{}
	jobs, err := jobInfo.FindAllJobInfo()
	common.PanicIf(err)
	if len(jobs) > 0 {
		for _, job := range jobs {
			j := &quartz.Job{
				ID:         job.ID,
				Name:       job.Name,
				Group:      job.Group,
				Expression: job.Cron,
				Params:     job.Param,
				Active:     job.IsActivity,
				JobFunc:    this.InvokeJob,
				URL:        job.Urls,
			}
			err := this.qz.AddJob(j)
			log.Println("add job error ", err)

		}
	}
}

//execute job
func (this *JobManager) InvokeJob(jobId int, targetUrl, params string, nextTime time.Time) {
	log.Println("jobId=", jobId, " targetUrl=", targetUrl, " params=", params)

	jobInfo := &entity.JobInfo{ID: jobId}
	err := jobInfo.GetJobInfo()
	if err != nil {
		err = this.qz.RemoveJob(jobId)
		log.Println("remove the job error ", err)
		return
	}
	invoker := &invoke.Invoker{}
	err = invoker.Execute(jobInfo, nextTime, params)
	if err != nil {
		log.Println("err = ", err)
	}
}

// add job
func (this *JobManager) AddJob(jobInfo *entity.JobInfo) error {
	j := &quartz.Job{
		ID:         jobInfo.ID,
		Name:       jobInfo.Name,
		Group:      jobInfo.Group,
		Expression: jobInfo.Cron,
		Params:     jobInfo.Param,
		Active:     jobInfo.IsActivity,
		JobFunc:    this.InvokeJob,
		URL:        jobInfo.Urls,
	}
	return this.qz.AddJob(j)
}

// modify a job
func (this *JobManager) ModifyJob(jobInfo *entity.JobInfo) error {
	j := &quartz.Job{
		ID:         jobInfo.ID,
		Name:       jobInfo.Name,
		Group:      jobInfo.Group,
		Expression: jobInfo.Cron,
		Params:     jobInfo.Param,
		Active:     jobInfo.IsActivity,
		JobFunc:    this.InvokeJob,
		URL:        jobInfo.Urls,
	}
	return this.qz.ModifyJob(j)
}

// delete job
func (this *JobManager) DeleteJob(jobId int) error {
	return this.qz.RemoveJob(jobId)
}

func (this *JobManager) GetJobSnapshotList() ([]*quartz.Job, error) {
	return this.qz.SnapshotJob()
}
