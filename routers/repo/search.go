// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (

	//"fmt"
	git "github.com/gogits/git-module"

	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"

)

const (
	THIS_REPO_SEARCH    base.TplName = "repo/search"
)

func ThisRepoSearch(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("search_this_repo")
	ctx.Data["PageIsSearch"] = true
	ctx.Data["PageIsSearchThisRepo"] = true

	RenderThisRepoSearch(ctx, &ThisRepoSearchOptions{
		Counter:  0,
		Ranger:   0,
		PageSize: 0,
		OrderBy:  "",
		TplName:  THIS_REPO_SEARCH,
	})
}

type ThisRepoSearchOptions struct {
	Counter  int64
	Ranger   int64
	PageSize int
	OrderBy  string
	TplName  base.TplName
}

func RenderThisRepoSearch(ctx *context.Context, opts *ThisRepoSearchOptions) {
	

	page := ctx.QueryInt("page")
	if page <= 0 {
		page = 1
	}

	order := ctx.Query("order")
	if order != "reverse" {
		order = "--date-order"
	} else {
		order = "--date-order --reverse"
	}

	var (
		matches *git.MatchesResults
		count int64
		err   error
	)

	keyword := ctx.Query("q")
	if len(keyword) == 0 {
		//matches = nil
		err = nil
		count = opts.Counter
	} else {
		matches, err = ctx.Repo.GitRepo.ShearchMatchesThisRepo(&git.RepoSearchOptions{
			Keyword:  keyword,
			OrderBy:  order,
			Page:     page,
			PageSize: 10,
		},)

		if err != nil {
			ctx.Handle(500, "ShearchMatchesThisRepo", err)
			return
		}
	}

	/*for _, element := range matches.Results{
		fmt.Println(element.CommitID)
	}*/

	ctx.Data["Keyword"] = keyword
	ctx.Data["Total"] = count
	ctx.Data["Matches"] = matches


	ctx.HTML(200, opts.TplName)
}


