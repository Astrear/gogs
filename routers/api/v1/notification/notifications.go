// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package notification

import (
	api "github.com/gogits/go-gogs-client"
	"time"
	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/base"
)

func GetUserNotifications(ctx *context.APIContext) {
	userID := ctx.User.ID
	notifications, err := models.GetLastUserNotifications(userID)
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	results := make([]*api.Notification, len(notifications))
	for i := range notifications {
		results[i] = &api.Notification{
			ID:       notifications[i].ID,
			UserID:   notifications[i].UserID,
			Watched:   notifications[i].Watched, 
			Description:   notifications[i].Description,
			CreatedUnix:   notifications[i].CreatedUnix,
			TimeSince: 	   base.RawTimeSince(time.Unix(notifications[i].CreatedUnix,0) ,"es-ES"),
		}
	}

	count:= models.CountUnWatchedUserNotifications(userID)

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": results,
		"count": count,
	})
}

func UpdateWatchNotifications(ctx *context.APIContext) {
	userID := ctx.User.ID
	err := models.UpdateAllNotificationsWatched(userID)
	if err != nil {
		ctx.JSON(500, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
	})
}
