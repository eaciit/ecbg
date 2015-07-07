package ecbg

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/session"
	"github.com/eaciit/errorlib"
)

type IController interface {
	Context() *context.Context
	SetSession(string, interface{})
	GetSession(string) interface{}
	ReleaseSession()
}

type Controller struct {
	beego.Controller
	sstr session.SessionStore
}

func (ec *Controller) Context() *context.Context {
	return ec.Ctx
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

func (ec *Controller) Json(data interface{}) {
	ec.Data["json"] = data
	ec.ServeJson()
}
