package agent

import (
	"cron-s/internal/task"
	"github.com/gin-gonic/gin"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"time"
)

type router struct {
	taskScheduler *task.Scheduler
}

func newRouter(ts *task.Scheduler) *router {
	return &router{
		taskScheduler: ts,
	}
}

func (r *router) Tasks(c *gin.Context) {
	c.JSON(200, r.taskScheduler.Heap)
}

func (r *router) TaskSave(c *gin.Context) {
	t := task.New()
	err := c.BindJSON(t)
	if err != nil {
		log.Error("schedule: http.handleTaskSave Unmarshal err", err)
		return
	}
	t.CronExpression, err = cronexpr.Parse(t.CronLine)
	if err != nil {
		log.Error("schedule: http.handleTaskSave Parse err", err)
		return
	}
	t.PlanExecTime = t.CronExpression.Next(time.Now())

	r.taskScheduler.Heap.Push(t)
	r.taskScheduler.Renew()

	c.String(200, "ok")
}

func (r *router) TaskDel(c *gin.Context) {
	t := task.New()
	err := c.BindJSON(t)
	if err != nil {
		log.Error("schedule: http.handleTaskSave ReadAll err", err)
		return
	}

	r.taskScheduler.Heap.Remove(t)
	r.taskScheduler.Renew()

	c.String(200, "ok")
}

func (r *router) Join(c *gin.Context) {
	nodeID := c.Query("nodeID")
	peerAddress := c.Query("peerAddress")

	index := r.taskScheduler.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(peerAddress), 0, 3*time.Second)
	if err := index.Error(); err != nil {
		log.Error("router Join err, ", err)
		return
	}

	c.String(200, "ok")
}
