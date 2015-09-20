package ecbg

import (
	"fmt"
	_ "github.com/astaxie/beego"
	dbs "github.com/eaciit/database/base"
	"github.com/eaciit/orm"
	"github.com/eaciit/toolkit"
	"strings"
)

/*
skip: skipped record
page: page index
sort[0][field]:
sort[0][dir]
filer[logic]
filter[filters][0][logic]
filter[filters][0][field]
filter[filters][0][operator]
filter[filters][0][value]
*/
var e error

func (a *Controller) KendoGridSettings(ins toolkit.M) toolkit.M {
	if ins == nil {
		ins = toolkit.M{}
	}
	s := toolkit.M{}
	q_skip := a.Ctx.Input.Query("skip")
	q_page := a.Ctx.Input.Query("page")
	q_size := a.Ctx.Input.Query("pageSize")

	if q_skip != "" {
		s.Set("skip", toolkit.ToInt(q_skip))
	}

	if q_page != "" {
		s.Set("page", toolkit.ToInt(q_page))
	}

	if q_size != "" {
		s.Set("limit", toolkit.ToInt(q_size))
	}

	sortField := strings.ToLower(a.Ctx.Input.Query("sort[0][field]"))
	sortDir := a.Ctx.Input.Query("sort[0][dir]")

	if sortField != "" {
		if sortField == "id" {
			sortField = "_id"
		}
		if sortDir == "" || sortDir == "asc" {
			s.Set("order", []string{sortField})
		} else {
			s.Set("order", []string{"-" + sortField})
		}
	}

	if fqe := a.KendoGridFilter("where"); fqe != nil {
		if ins.Has("filter") {
			fqe = dbs.And(fqe, ins.Get("where").(*dbs.QE))
		}
		s.Set("where", fqe)
	}

	return s
}

func (a *Controller) KendoGridFilter(parent string) *dbs.QE {
	input := a.Ctx.Input
	logic := input.Query(parent + "[logic]")
	//fmt.Printf("Filter for %s logic: %s \n", parent, logic)
	//--- has subfilters
	if logic == "" {
		field := strings.ToLower(input.Query(parent + "[field]"))
		if field == "id" {
			field = "_id"
		}
		op := input.Query(parent + "[operator]")
		value := input.Query(parent + "[value]")
		//fmt.Printf("Op: %v Value: %v\n", op, value)
		if op == "eq" {
			return dbs.Eq(field, value)
		} else if op == "contains" {
			return dbs.Contains(field, value)
		} else if op == "notcontains" {
			return dbs.Contains(field, value)
		} else if op == "startswith" {
			return dbs.StartWith(field, value)
		} else if op == "endswith" {
			return dbs.EndWith(field, value)
		} else {
			return nil
		}
	} else {
		filters := []*dbs.QE{}
		iChild := 0
		var qeFilter *dbs.QE
		filterOk := true
		for valid := filterOk; valid == true; valid = filterOk {
			qeFilter = a.KendoGridFilter(fmt.Sprintf("%s[filters][%d]", parent, iChild))
			if qeFilter != nil {
				filters = append(filters, qeFilter)
			} else {
				filterOk = false
			}
			iChild++
			//fmt.Printf("Filter for %s qe: %s  valid:%v \n", parent, toolkit.JsonString(qeFilter), filterOk)
		}

		//fmt.Printf("Filter done %v\n", toolkit.JsonString(filters))
		if logic == "or" {
			return dbs.Or(filters...)
		} else {
			return dbs.And(filters...)
		}
	}

	return nil
}

func (c *Controller) KendoGridData(obj orm.IModel, objs interface{}, ins toolkit.M) *toolkit.Result {
	result := toolkit.NewResult()
	s := c.KendoGridSettings(ins)
	cursor := c.Orm.Find(obj, s)
	e = cursor.FetchAll(objs, true)

	s.Unset("limit")
	s.Unset("skip")
	cursor = c.Orm.Find(obj, s)
	defer cursor.Close()
	count := cursor.Count()

	if e != nil {
		result.Status = toolkit.Status_NOK
		result.Message = e.Error()
	} else {
		//result.Data = toolkit.M{"Data": dsUsers.Data,
		result.Data = toolkit.M{"Data": objs,
			"Count": count}
		result.Status = toolkit.Status_OK
	}

	return result
}
