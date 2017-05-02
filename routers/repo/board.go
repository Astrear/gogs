// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"strings"
	"github.com/gogits/gogs/models"
	//"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	//"github.com/gogits/gogs/modules/log"
	//"github.com/gogits/gogs/modules/markdown"
)

const (
	BOARD    base.TplName = "repo/board/list"
	CARDSUSER base.TplName = "repo/cardsuser"
)

func Board(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("board")
	ctx.Data["PageIsBoard"] = true

	lists, err := ctx.Repo.Repository.GetLists()
	if err != nil {
		ctx.Handle(500, "Lists", err)
		return
	}
	var noncoll []string
	for _, list := range lists {
		for _, card := range list.Cards {
			if models.IsCollaboratorOfRepo(card.AssigneeID, ctx.Repo.Repository.ID) == false {
				if user, err := models.GetUserByID(card.AssigneeID); err == nil{
					if(user.ID != ctx.Repo.Repository.OwnerID){
						if !isInSlice(user.Name, noncoll) {
							noncoll = append(noncoll, user.Name)
						}
					}
				} 
			}
		}
	}

	collaborators, err := ctx.Repo.Repository.GetCollaborators()
	if err != nil {
		fmt.Println("GetCollaborators: %v", err)
	}
	
	ctx.Data["Lists"] = lists
	ctx.Data["Collaborators"] = collaborators
	ctx.Data["NonColl"] = strings.Join(noncoll,",")
	ctx.HTML(200, BOARD)
}

func isInSlice(item string, slice []string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

func CardsByUser(ctx *context.Context) {
	ctx.Data["PageIsContributions"] = true
	collaborators, err := ctx.Repo.Repository.GetCollaborators()
	if err != nil {
		fmt.Println("GetCollaborators: %v", err)
	}

	cardsOwner,err := models.GetCardsbyUser(ctx.Repo.Repository.OwnerID, ctx.Repo.Repository.ID)
	if err != nil {
		fmt.Println("GetCardsbyUser: %v", err)
	}

	ctx.Data["Collaborators"] = collaborators
	ctx.Data["CardsOwner"] = cardsOwner
	
	ctx.HTML(200, CARDSUSER)
}