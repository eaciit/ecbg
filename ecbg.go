package ecbg

import (
	//"github.com/astaxie/beego"
	"github.com/astaxie/beego/session"
)

var SessionMgr *session.Manager

const (
	packageName   = "eaciit.ecbg"
	modController = "Controller"
)

func init() {
	if SessionMgr == nil {
		SessionMgr, _ = session.NewManager("memory", `{"cookieName":"ecsessionid", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 3600, "providerConfig": ""}`)
		go SessionMgr.GC()
	}
}
