package ecbg

import (
	"github.com/astaxie/beego"
	//"github.com/astaxie/beego/context"
	"fmt"
	"github.com/astaxie/beego/session"
	"github.com/eaciit/database/base"
	"github.com/eaciit/database/proxy"
	"github.com/eaciit/errorlib"
	"github.com/eaciit/orm"
)

/*
type IController interface {
	Context() *context.Context
	SetSession(string, interface{})
	GetSession(string) interface{}
	ReleaseSession()
}
*/

type Controller struct {
	beego.Controller
	sstr session.SessionStore
	Orm  *orm.DataContext
	Db   base.IConnection
}

func (ec *Controller) Prepare() {
	host := beego.AppConfig.String("db_server")
	username := beego.AppConfig.String("db_username")
	password := beego.AppConfig.String("db_password")
	database := beego.AppConfig.String("db_name")
	dbtype := beego.AppConfig.String("db_type")

	if dbtype != "" {
		db, e := proxy.NewConnection(dbtype, host, username, password, database)
		if e != nil {
			beego.Error(e.Error())
		}
		ec.Db = db
	}

	fmt.Println("Prepare")
	if e := ec.Db.Connect(); e != nil {
		beego.Error("Unable to connect to database")
	}
	fmt.Println("Connected")
	ec.Orm = orm.New(ec.Db)
}

func (ec *Controller) prepareSession() error {
	ctx := ec.Ctx
	w := ctx.ResponseWriter
	r := ctx.Request
	store, err := SessionMgr.SessionStart(w, r)
	ec.sstr = store
	return err
}

func (ec *Controller) Finish() {
	if ec.sstr != nil {
		w := ec.Ctx.ResponseWriter
		ec.sstr.SessionRelease(w)
	}

	if ec.Orm != nil {
		ec.Orm.Close()
	}

	if ec.Db != nil {
		ec.Db.Close()
	}
}

func (ec *Controller) GetSession(key string, def interface{}) interface{} {
	if ec.sstr == nil {
		errSession := ec.prepareSession()
		if errSession != nil {
			return def
		}
	}

	val := ec.sstr.Get(key)
	if val == nil {
		return def
	}
	return val
}

func (ec *Controller) SetSession(key string, o interface{}) error {
	if ec.sstr == nil {
		errSession := ec.prepareSession()
		if errSession != nil {
			return errorlib.Error(packageName, modController, "SetSession", "Session Store could not be prepared = "+errSession.Error())
		}
	}

	return ec.sstr.Set(key, o)
}

func (ec *Controller) DeleteSession(key string, o interface{}) error {
	if ec.sstr == nil {
		errSession := ec.prepareSession()
		if errSession != nil {
			return errorlib.Error(packageName, modController, "DeleteSession", "Session Store could not be prepared = "+errSession.Error())
		}
	}

	return ec.sstr.Delete(key)
}

func (ec *Controller) Json(data interface{}) {
	ec.Data["json"] = data
	ec.ServeJson()
}
