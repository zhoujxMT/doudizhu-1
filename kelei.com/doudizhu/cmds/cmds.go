package cmds

import (
	"bytes"
	"fmt"

	eng "kelei.com/doudizhu/engine"
	"kelei.com/utils/common"
	"kelei.com/utils/delaymsg"
	"kelei.com/utils/frame/command"
	"kelei.com/utils/frame/config"
	"kelei.com/utils/logger"
)

var (
	engine *eng.Engine
)

func Inject(engine_ *eng.Engine) {
	engine = engine_
}

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法
	commands["usercount"] = usercount
	commands["roomcount"] = roomcount
	commands["roominfo"] = roominfo
	commands["cardcount"] = cardcount
	commands["tuoguan"] = tuoguan
	commands["loguserid"] = loguserid
	commands["task"] = task
	commands["cardmode"] = cardmode
	//end
	commands["mode"] = config.Mode
	command.CreateHelp(commands)
	return commands
}

func cardmode(cmdVal string) {
	if cmdVal == "" {
		logger.Infof("当前模式 : %d", eng.GetCardMode())
	} else {
		if cardMode, err := common.CheckNum(cmdVal); err == nil {
			eng.SetCardMode(cardMode)
		}
	}
}

func task(cmdVal string) {
	if cmdVal == "" {
		logger.Infof("当前任务 : %d", eng.GetFixedTaskID())
	} else {
		if fixedTaskID, err := common.CheckNum(cmdVal); err == nil {
			eng.SetFixedTaskID(fixedTaskID)
		}
	}
}

func loguserid(cmdVal string) {
	if cmdVal == "" {
		logger.Infof("loguserid : %s", delaymsg.GetLogUserID())
	} else {
		delaymsg.SetLogUserID(cmdVal)
	}
}

func usercount(cmdVal string) {
	logger.Infof("玩家数量:%d", eng.UserManage.GetUserCount())
}

func roomcount(cmdVal string) {
	logger.Infof("房间数量:%d", eng.RoomManage.GetRoomCount())
}

func roominfo(cmdVal string) {
	rooms := eng.RoomManage.GetRooms()
	buff := bytes.Buffer{}
	buff.WriteString("{\n")
	for _, room := range rooms {
		buff.WriteString(fmt.Sprintf("     id:%s,比赛类型:%d,房间类型:%d,房间人数:%d,观战人数:%d\n", *room.GetRoomID(), room.GetMatchID(), room.GetRoomType(), room.GetPCount(), room.GetIdlePCount()))
	}
	buff.WriteString("}")
	logger.Infof(buff.String())
	logger.Infof("房间数量:%d", eng.RoomManage.GetRoomCount())
}

func cardcount(commVal string) {
	if commVal == "" {
		logger.Infof("每个人的牌数量 : %d", eng.GetCardCount())
	} else {
		if cardcount, err := common.CheckNum(commVal); err == nil {
			eng.SetCardCount(cardcount)
		}
	}
}

func tuoguan(commVal string) {
	if commVal == "" {
		logger.Infof("请设置托管参数")
	} else {
		if v, err := common.CheckNum(commVal); err == nil {
			rooms := eng.RoomManage.GetRooms()
			for _, room := range rooms {
				room.SetAllUsersTrusteeshipStatus(v == 1)
			}
		}
	}
}
