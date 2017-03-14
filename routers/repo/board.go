// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"

	//"github.com/gogits/gogs/models"
	//"github.com/gogits/gogs/modules/auth"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	//"github.com/gogits/gogs/modules/log"
	//"github.com/gogits/gogs/modules/markdown"
)

const (
	BOARD    base.TplName = "repo/board/list"
)

func Board(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.release.board")
	ctx.Data["PageIsBoardList"] = true

	lists, err := ctx.Repo.Repository.GetLists()
	if err != nil {
		ctx.Handle(500, "Lists", err)
		return
	}
	for _, item := range lists {
		fmt.Println("%+v", item)
		for _, list := range item.Cards {
			fmt.Println("%+v", list)
		}
	}

	collaborators, err := ctx.Repo.Repository.GetCollaborators()
	if err != nil {
		fmt.Println("GetCollaborators: %v", err)
	}

	ctx.Data["Lists"] = lists
	ctx.Data["Collaborators"] = collaborators
	ctx.HTML(200, BOARD)
}
