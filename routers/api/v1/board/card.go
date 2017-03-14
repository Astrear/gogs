// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package board

import (
	"fmt"
	"strings"

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func CreateCard(ctx *context.APIContext, form api.CreateCardOption) {

	card := &models.Card{
		ListID:   		form.List,
		Position: 		form.Index,
		Description:    form.Body,
		State: 			4,
	}

	if ctx.Repo.IsWriter() {
		if len(form.Assignee) > 0 {
			assignee, err := models.GetUserByName(form.Assignee)
			if err != nil {
				if models.IsErrUserNotExist(err) {
					ctx.Error(422, "", fmt.Sprintf("Assignee does not exist: [name: %s]", form.Assignee))
				} else {
					ctx.Error(500, "GetUserByName", err)
				}
				return
			}
			card.AssigneeID = assignee.ID
		}
	}

	if err := models.NewCard(ctx.Repo.Repository, card); err != nil {
		ctx.Error(500, "NewCard", err)
		return
	}

	var err error
	card, err = models.GetCardByID(card.ID)
	if err != nil {
		ctx.Error(500, "GetIssueByID", err)
		return
	}
	ctx.JSON(201, card.APIFormat())
}

func EditCard(ctx *context.APIContext, form api.EditCardOption) {
	card, err := models.GetCardByIndex(ctx.ParamsInt64(":list"), ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByIndex", err)
		}
		return
	}

	if form.Index >= 0 && form.Index != card.Position {
		card.Position = form.Index
	}

	if len(form.Body) > 0 {
		card.Description = form.Body
	}

	if ctx.Repo.IsWriter() && len(form.Assignee) > 0 &&
		(card.Assignee == nil || card.Assignee.LowerName != strings.ToLower(form.Assignee)) {
		if len(form.Assignee) == 0 {
			card.AssigneeID = 0
		} else {
			assignee, err := models.GetUserByName(form.Assignee)
			if err != nil {
				if models.IsErrUserNotExist(err) {
					ctx.Error(422, "", fmt.Sprintf("assignee does not exist: [name: %s]", form.Assignee))
				} else {
					ctx.Error(500, "GetUserByName", err)
				}
				return
			}
			card.AssigneeID = assignee.ID
		}
	}

	if form.State > 0  && form.State <= 4 {
		if form.State != card.State {
			card.State = form.State
		}
	}

	if form.Limit > 0 {
	    card.LimitDateUnix = form.Limit
	}

	if form.List != card.ListID {
		if err := card.MoveCard(form.List); err != nil{
			ctx.Error(500, "MoveCard", err)
			return
		}
	}

	if err = models.UpdateCard(card); err != nil {
		ctx.Error(500, "UpdateCard", err)
		return
	}

	// Refetch from database to assign some automatic values
	card, err = models.GetCardByID(card.ID)
	if err != nil {
		ctx.Error(500, "GetCardByID", err)
		return
	}
	ctx.JSON(201, card.APIFormat())
}

func DeleteCard(ctx *context.APIContext) {
	if err := models.DeleteCardByListID(ctx.ParamsInt64(":list"), ctx.ParamsInt64(":id"), ctx.Repo.Repository.ID); err != nil {
		ctx.Error(500, "DeleteCardByListID", err)
		return
	}
	ctx.Status(204)
}