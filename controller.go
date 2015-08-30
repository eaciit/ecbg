package ecbg

import (
	"github.com/astaxie/beego"
	//"github.com/astaxie/beego/context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/session"
	"github.com/eaciit/database/base"
	"github.com/eaciit/database/proxy"
	"github.com/eaciit/errorlib"
	"github.com/eaciit/orm"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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

//-- name is for later purpose
func (p *Controller) SaveFiles(key string, folder string, filenamePattern string) error {
	hs, eGet := p.GetFiles(key)
	if eGet != nil {
		return errors.New("Unable to get file: " + eGet.Error())
	}
	for _, h := range hs {
		f, eOpen := h.Open()
		if eOpen != nil {
			return errors.New("Unable to open: " + eOpen.Error())
		}
		defer f.Close()

		newFileName := filepath.Join(folder, h.Filename)
		if eWrite := copyFile(f, newFileName); eWrite != nil {
			return errors.New("Unable to write file " + h.Filename + " : " + eWrite.Error())
		}
	}
	return nil
}

func (p *Controller) SaveFile(key string, folder string, filename string) error {
	f, h, e := p.GetFile(key)
	if e != nil {
		return errors.New("Unable to open file: " + e.Error())
	}
	defer f.Close()
	if filename == "" {
		filename = h.Filename
	}

	filename = filepath.Join(folder, filename)
	return copyFile(f, filename)
}

func copyFile(source io.Reader, tofile string) error {
	fnew, errWrite := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if errWrite != nil {
		return errors.New("Unable to write file: " + errWrite.Error())
	}
	defer fnew.Close()
	io.Copy(fnew, source)
	return nil
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
			return errorlib.Error(packageName, modController,
				"SetSession", "Session Store could not be prepared = "+errSession.Error())
		}
	}

	return ec.sstr.Set(key, o)
}

func (ec *Controller) DeleteSession(key string, o interface{}) error {
	if ec.sstr == nil {
		errSession := ec.prepareSession()
		if errSession != nil {
			return errorlib.Error(packageName, modController,
				"DeleteSession", "Session Store could not be prepared = "+errSession.Error())
		}
	}

	return ec.sstr.Delete(key)
}

func (p *Controller) GetPayload(result interface{}) error {
	body := p.Ctx.Request.Body
	decoder := json.NewDecoder(body)
	return decoder.Decode(result)
}

func (p *Controller) GetPayloadMultipart(result interface{}) (map[string][]*multipart.FileHeader,
	map[string][]string, error) {
	var e error
	e = p.Ctx.Request.ParseMultipartForm(1024 * 1024 * 1024 * 1024)
	if e != nil {
		return nil, nil, fmt.Errorf("Unable to parse: %s", e.Error())
	}
	m := p.Ctx.Request.MultipartForm
	return m.File, m.Value, nil
}

func (ec *Controller) Json(data interface{}) {
	ec.Data["json"] = data
	ec.ServeJson()
}
