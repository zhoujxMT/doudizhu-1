/*
玩家管理
*/

package engine

import (
	"net"
	"sync"

	"kelei.com/utils/delaymsg"
)

type userManage struct {
	users map[string]*User //比赛中的玩家列表
	lock  sync.Mutex
}

func userManage_init() *userManage {
	UserManage := userManage{}
	UserManage.users = map[string]*User{}
	return &UserManage
}

var (
	UserManage = *userManage_init()
)

//重置玩家
func (u *userManage) Reset() {
	u.lock.Lock()
	defer u.lock.Unlock()
	for _, user := range u.users {
		user.deleteUserInfo()
		user.reset()
	}
	u.users = map[string]*User{}
}

//获取玩家
func (u *userManage) GetAllUsers() map[string]*User {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.users
}

//获取玩家
func (u *userManage) GetUser(userid *string) *User {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.users[*userid]
}

//获取玩家根据UID
func (u *userManage) GetUserByUID(uid *string) *User {
	u.lock.Lock()
	defer u.lock.Unlock()
	for _, user := range u.users {
		if *user.getUID() == *uid {
			return user
		}
	}
	return nil
}

//获取比赛中的玩家数量
func (u *userManage) GetUserCount() int {
	u.lock.Lock()
	defer u.lock.Unlock()
	return len(u.users)
}

//创建一个玩家
func (u *userManage) createUser() *User {
	user := User{}
	user.cards = []Card{}
	user.cardTypeRecord = map[int]int{}
	return &user
}

//根据userid创建玩家
func (u *userManage) newUser(userid *string) *User {
	user := UserManage.createUser()
	user.setUserID(userid)
	return user
}

//添加玩家(比赛中的玩家,不包含功能中的玩家)
func (u *userManage) AddUser(uid *string, userid *string, conn net.Conn) *User {
	user := u.users[*userid]
	if user == nil {
		user = u.createUser()
		user.uid = uid
		user.userid = userid
		user.conn = conn
		user.dm = delaymsg.NewDelayMessage()
		user.dm.SetUserID(*userid)
		user.dm.Start()
		user.online = true
		u.lock.Lock()
		defer u.lock.Unlock()
		u.users[*userid] = user
	} else {
		user.conn = conn
	}
	return user
}

//添加玩家
func (u *userManage) AddTheUser(user *User) {
	userid := *user.getUserID()
	if u.users[userid] != nil {
		return
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	u.users[userid] = user
}

//删除一个玩家
func (u *userManage) RemoveUser(user *User) {
	delete(u.users, *user.getUserID())
}
