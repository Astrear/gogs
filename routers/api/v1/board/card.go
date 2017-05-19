// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package board

import (
	"time"
	"fmt"
	"strings"

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func CreateCard(ctx *context.APIContext, form api.CreateCardOption) {

	card := &models.Card{
		RepoID: 		ctx.Repo.Repository.ID,
		ListID:   		form.List,
		Position: 		form.Index,
		Description:    form.Body,
		State: 			models.CARD_STATE_PLANNED,
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
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByID", err)
		}
		return
	}

	if ctx.User.CanEditCard(card.ID) || ctx.Repo.IsOwner() {
		if !card.IsClosed() { 
			if len(form.Body) > 0 {
				card.Description = form.Body
			}
			card.Priority = form.Priority
			if len(form.Assignee) > 0 && (card.Assignee == nil || card.Assignee.LowerName != strings.ToLower(form.Assignee)) {
				assignee, err := models.GetUserByName(form.Assignee)
				if err != nil {
					if models.IsErrUserNotExist(err) {
						ctx.Error(422, "", fmt.Sprintf("assignee does not exist: [name: %s]", form.Assignee))
					} else {
						ctx.Error(500, "GetUserByName", err)
					}
					return
				}

				repoName := ctx.Repo.Repository.Owner.Name + "/" + ctx.Repo.Repository.Name

				if card.Assignee != nil && card.Assignee.LowerName != strings.ToLower(form.Assignee) {
					if err := models.CreateNotification(card.AssigneeID, "Se te ha reemplazado como reponsable de una tarjeta en " + repoName, 9, ctx.Repo.Repository.HTMLURL()); err != nil{
						fmt.Errorf("Error at CreateNotification in EditCard: %v", err)
					}else{
						models.SendEmailCardReplaced(card.AssigneeID, repoName, ctx.Repo.Repository.HTMLURL(), card.Description)
					}
				}

				if err := models.CreateNotification(assignee.ID, "Se te ha asignado una tarjeta en " + repoName, 8, ctx.Repo.Repository.HTMLURL()); err != nil{
					fmt.Errorf("Error at CreateNotification in EditCard: %v", err)
				}
				models.SendEmailCardAsigned(assignee.ID, repoName, ctx.Repo.Repository.HTMLURL(), card.Description)

				card.AssigneeID = assignee.ID
			}
		}
	} else {
		ctx.Status(403)
		return
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

func EditCardDuration(ctx *context.APIContext, form api.EditCardDurationOption) {
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByID", err)
		}
		return
	}

	if ctx.User.CanEditCard(card.ID) || ctx.Repo.IsOwner() {
		if !card.IsClosed() {
			card.Duration 		= form.Duration
			card.ActivatedUnix 	= time.Now().Unix()
		}
	} else {
		ctx.Status(403)
		return
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

func EditCardState(ctx *context.APIContext) {
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByListID", err)
		}
		return
	}

	if ctx.User.CanEditCard(card.ID) || ctx.Repo.IsOwner() {
		if card.IsActive() {
			card.State 		= models.CARD_STATE_CLOSED
			card.Duration 	= 0
		} else if card.IsExpired() {
			if card.Duration > 0 {
				card.State 			= models.CARD_STATE_ACTIVE
				card.ActivatedUnix 	= time.Now().Unix()
			} else {
				card.State 		= models.CARD_STATE_CLOSED
				card.Duration 	= 0
				if card.TimeElapsed > 0 {
					card.TimeElapsed += time.Now().Unix() - card.ActivatedUnix
				} else {
					card.TimeElapsed = time.Now().Unix()
				}
			}
		} else {
			card.State 			= models.CARD_STATE_ACTIVE
			card.ActivatedUnix 	= time.Now().Unix()
		}
	} else {
		ctx.Status(403)
		return
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

func EditCardIndex(ctx *context.APIContext, form api.EditCardIndexOption) {
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByID", err)
		}
		return
	}

	if !card.IsClosed() {
		card.Position = form.Index
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

func TransferCard(ctx *context.APIContext, form api.TransferCardOption){
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByID", err)
		}
		return
	}

	if !card.IsClosed() {
		if err := card.UpdateCardList(form.List); err != nil {
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

func ExpireCard(ctx *context.APIContext) {
	card, err := models.GetCardByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrCardNotExist(err) {
			ctx.Status(404)
		} else {
			ctx.Error(500, "GetCardByID", err)
		}
		return
	}

	card.State = models.CARD_STATE_EXPIRED
	card.Duration = 0

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

	//SEND NOTIFICATION
	repoName := ctx.Repo.Repository.Owner.Name + "/" + ctx.Repo.Repository.Name;
	if card.HasAssignee() {
		if err := models.CreateNotification(card.AssigneeID, "Ha caducado una de tus tarjetas en " + repoName, 11, ctx.Repo.Repository.HTMLURL()); err != nil{
			fmt.Errorf("Error at CreateNotification in ExpireCard: %v", err)
		}else{
			models.SendEmailCardExpired(card.AssigneeID, repoName, ctx.Repo.Repository.HTMLURL(), card.Description)
		}
	}

	if(ctx.Repo.Owner.ID != card.AssigneeID){
		if err := models.CreateNotification(ctx.Repo.Owner.ID, "Ha caducado una tarjeta en " + ctx.Repo.Repository.Name, 11, ctx.Repo.Repository.HTMLURL()); err != nil{
			fmt.Errorf("Error at CreateNotification in ExpireCard: %v", err)
		}
	}
	//SEND NOTIFICATION

	ctx.JSON(201, card.APIFormat())
}

func DeleteCard(ctx *context.APIContext) {
	
	if ctx.User.CanEditCard(ctx.ParamsInt64(":id")) || ctx.Repo.IsOwner() {

		card, errCard := models.GetCardByID(ctx.ParamsInt64(":id"))
		if errCard != nil{
			fmt.Errorf("Error at GetCardByID in DeleteCard: %v", errCard)
		}

		//SEND NOTIFICATION
		repoName := ctx.Repo.Repository.Owner.Name + "/" + ctx.Repo.Repository.Name;
		if err := models.CreateNotification(card.AssigneeID, "Se ha eliminado una de tus tarjetas de " + repoName, 10, ctx.Repo.Repository.HTMLURL()); err != nil{
			fmt.Errorf("Error at CreateNotification in DeleteCard: %v", err)
		}else{
			//NOTIFICACION POR CORREO
			models.SendEmailCardDeleted(card.AssigneeID, repoName, ctx.Repo.Repository.HTMLURL(), card.Description)
		}

		if err := models.CreateNotification(ctx.Repo.Repository.OwnerID, "Se ha eliminado una tarjeta de " + ctx.Repo.Repository.Name, 10, ctx.Repo.Repository.HTMLURL()); err != nil{
			fmt.Errorf("Error at CreateNotification in DeleteCard: %v", err)
		}
		//SEND NOTIFICATION

		if err := models.DeleteCard(ctx.ParamsInt64(":id")); err != nil {
			ctx.Error(500, "DeleteCardByListID", err)
			return
		}

		ctx.Status(204)
		return
	}
	ctx.Status(403)
}

func GetCardsUser(ctx *context.APIContext){
	userID:= ctx.ParamsInt64(":userID")
	repoID:= ctx.ParamsInt64(":repoID")

	cardsUser, err := models.GetCardsbyUser(userID, repoID)
	if err != nil{
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	results := make([]*api.Card, len(cardsUser))
	for i := range cardsUser {
		results[i] = cardsUser[i].APIFormat()
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
	})
}