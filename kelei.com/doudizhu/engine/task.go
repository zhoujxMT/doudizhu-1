package engine

import (
	"kelei.com/utils/common"
)

var (
	FixedTaskID = -1 //固定的任务ID为了测试
)

func GetFixedTaskID() int {
	return FixedTaskID
}

func SetFixedTaskID(fixedTaskID int) {
	FixedTaskID = fixedTaskID
}

type Task struct {
	ID      int
	Content string
	Award   int
}

func (this *Task) getID() int {
	return this.ID
}

func (this *Task) getContent() string {
	return this.Content
}

func (this *Task) getAward() int {
	return this.Award
}

type TaskSystem struct {
	Tasks []*Task
}

func (this *TaskSystem) getTasks() []*Task {
	return this.Tasks
}

func (this *TaskSystem) addTasks(task *Task) {
	this.Tasks = append(this.Tasks, task)
}

func (this *TaskSystem) randomGetTask() *Task {
	taskID := common.Random(0, len(this.getTasks()))
	if FixedTaskID >= 0 {
		taskID = FixedTaskID
	}
	return this.getTasks()[taskID]
}

var taskSystem *TaskSystem

func init() {
	taskSystem = &TaskSystem{[]*Task{}}
	taskSystem.addTasks(&Task{0, "自己打出两幅三带一并获得胜利", 1500})
	taskSystem.addTasks(&Task{1, "自己打出两幅三带二并获得胜利", 1500})
	taskSystem.addTasks(&Task{2, "自己打出一副炸弹并获得胜利", 1000})
	taskSystem.addTasks(&Task{3, "最后一手牌打出飞机并获得胜利", 2000})
	taskSystem.addTasks(&Task{4, "最后一手牌打出小王并获得胜利", 2000})
	taskSystem.addTasks(&Task{5, "自己打出一副飞机并获得胜利", 1000})
	taskSystem.addTasks(&Task{6, "第一手牌打出一副顺子并获得胜利", 1000})
	taskSystem.addTasks(&Task{7, "自己打出两幅顺子并获得胜利", 1500})
	taskSystem.addTasks(&Task{8, "最后一手牌打出梅花Q并获得胜利", 2000})
	taskSystem.addTasks(&Task{9, "最后一手牌打出大王并获得胜利", 1000})
	taskSystem.addTasks(&Task{10, "最后一手牌打出一副顺子并获得胜利", 1500})
	taskSystem.addTasks(&Task{11, "自己打出一副王炸并获得胜利", 1500})
}
