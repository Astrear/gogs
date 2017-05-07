// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package board

import (

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func CreateList(ctx *context.APIContext, form api.CreateListOption) {
	
	list := &models.List{
		RepoID: 	 ctx.Repo.Repository.ID,
		Title: 		 form.Title,
		Position: 	 form.Index,
	}

	if ctx.Repo.IsWriter() {
		if err := models.NewList(ctx.Repo.Repository, list); err != nil {
			ctx.Error(500, "NewCard", err)
			return
		}
	}

	var err error
	list, err = models.GetListByID(list.ID)
	if err != nil {
		ctx.Error(500, "GetListByID", err)
		return
	}
	ctx.JSON(201, list.APIFormat())
}

func EditList(ctx *context.APIContext, form api.EditListOption) {
	list, err := models.GetListByRepoID(ctx.Repo.Repository.ID, ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrListNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetListByIndex", err)
		}
		return
	}

	if len(form.Title) > 0 {
		if list.Title != form.Title{
			list.Title = form.Title
		}
	}

	if form.Index >= 0 {
		if list.Position != form.Index{
			list.Position = form.Index
		}	
	}

	if err := models.UpdateList(ctx.Repo.Repository, list); err != nil {
		ctx.Handle(500, "UpdateList", err)
		return
	}

	ctx.JSON(200, list.APIFormat())
}

func DeleteList(ctx *context.APIContext) {
	if err := models.DeleteListByRepoID(ctx.Repo.Repository.ID, ctx.ParamsInt64(":id")); err != nil {
		ctx.Error(500, "DeleteListByRepoID", err)
		return
	}
	ctx.Status(204)
}