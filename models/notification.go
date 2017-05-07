// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"time"
	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/base"
)

// Course represents an subject-user relation.
type Notification struct {
	ID         		int64 `xorm:"pk autoincr"`
	UserID     		int64 `xorm:"NOT NULL INDEX"`
	Watched    		bool  `xorm:"NOT NULL DEFAULT false"`
	Description		string `xorm:"VARCHAR(200) NOT NULL"`
	Type 			int
	Link			string
	CreatedUnix 	int64
}

func CreateNotification(userID int64, description string, typeNot int, link string) (err error) {
	
	n := &Notification{
		UserID: userID,
		Watched: false,
		Description: description,
		Type: typeNot,
		Link: link,
		CreatedUnix: time.Now().Unix(),
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(n); err != nil {
		return err
	} 

	return sess.Commit()
}

func UpdateWatched(nid int64) (err error){
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Exec("UPDATE `notification` SET watched=TRUE WHERE id=?", nid); err != nil {
		return err
	}

	return sess.Commit()
}

func UpdateAllNotificationsWatched(userID int64) (err error){
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Exec("UPDATE `notification` SET watched=TRUE WHERE user_id=?", userID); err != nil {
		return err
	}

	return sess.Commit()
}

func GetUserNotifications(uid int64) ([]*Notification, error) {
	notifications := make([]*Notification, 0)
	return notifications, x.OrderBy("id DESC").Find(&notifications, &Notification{UserID: uid})
}

func GetUserUnWatchedNotifications(uid int64) ([]*Notification, error) {
	notifications := make([]*Notification, 0)
	return notifications, x.OrderBy("id DESC").Where("watched = FALSE").Find(&notifications, &Notification{UserID: uid})
}

func GetLastUserNotifications(uid int64) ([]*Notification, error) {
	notifications := make([]*Notification, 0)
	return notifications, x.OrderBy("id DESC").Limit(5,0).Find(&notifications, &Notification{UserID: uid})
}

func CountUnWatchedUserNotifications(uid int64) int64 {
	sess := x.Where("user_id = ?", uid).And("watched = FALSE")

	count, err := sess.Count(new(Notification))
	if err != nil {
		log.Error(4, "CountUnWatchedUserNotifications: %v", err)
	}
	return count
}

func CountUserNotifications(uid int64) int64 {
	sess := x.Where("user_id = ?", uid)

	count, err := sess.Count(new(Notification))
	if err != nil {
		log.Error(4, "CountUserNotifications: %v", err)
	}
	return count
}

func TimeSinceNotification(timeNotification int64, lang string) string{
	return base.RawTimeSince(time.Unix(timeNotification,0) , lang)
}