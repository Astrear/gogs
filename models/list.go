// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"time"
	api "github.com/gogits/go-gogs-client"
	"github.com/go-xorm/xorm"

	//"github.com/gogits/gogs/modules/log"
)

// List represents a git repository.
type List struct {
	ID            int64  `xorm:"pk autoincr"`
	RepoID        int64  `xorm:"UNIQUE(s)"`
	Title         string `xorm:"UNIQUE(s) INDEX NOT NULL"`
	Position	  int64  `xorm:"UNIQUE(s)"`

	Cards         []*Card `xorm:"-"`
	NumCards 	  int `xorm:"DEFAULT 0"`

	Created       time.Time `xorm:"-"`
	CreatedUnix   int64
	Updated       time.Time `xorm:"-"`
	UpdatedUnix   int64
}

func (list *List) APIFormat() *api.List {
	apiList := &api.List{
		ID:      list.ID,
		Title:   list.Title,
		Index:   list.Position,
	}
	return apiList
}

func (list *List) BeforeInsert() {
	list.CreatedUnix = time.Now().Unix()
	list.UpdatedUnix = list.CreatedUnix
}

func (list *List) BeforeUpdate() {
	list.UpdatedUnix = time.Now().Unix()
}

func (list *List) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
		case "created_unix":
			list.Created = time.Unix(list.CreatedUnix, 0).Local()
		case "updated_unix":
			list.Updated = time.Unix(list.UpdatedUnix, 0)
	}
}

func countRepoLists(e Engine, repoID int64) int64 {
	count, _ := e.Where("repo_id=?", repoID).Count(new(List))
	return count
}

// CountRepoMilestones returns number of milestones in given repository.
func CountRepoLists(repoID int64) int64 {
	return countRepoLists(x, repoID)
}

func (list *List) loadCards(e Engine) (err error) {
	if list.Cards == nil {
		if err := list.getCards(e); err != nil {
			return fmt.Errorf("getCards [%d]: %v", list.ID, err)
		}
	}

	for _, card := range list.Cards {
		if err := card.LoadAttributes(); err != nil {
			return err
		}
	}

	return nil
}

func (list *List) LoadCards() error {
	return list.loadCards(x)
}

func isListExist(e Engine, repo *Repository, list *List) (bool, error) {
	has, err := e.Get(&List{
		RepoID:   repo.ID,
		Title:    list.Title,
	})
	return has, err
}

// IsListExist returns true if the repository with given name under user has already existed.
func IsListExist(repo *Repository, list *List) (bool, error) {
	return isListExist(x, repo, list)
}

func newList(e *xorm.Session, repo *Repository, list *List) (err error) {
	has, err := isListExist(e, repo, list)
	if err != nil {
		return fmt.Errorf("isListExist: %v", err)
	} else if has {
		return ErrListAlreadyExist{repo.ID, list.Title}
	}

	if _, err = e.Insert(list); err != nil {
		return err
	}

	if _, err = e.Exec("UPDATE `repository` SET num_lists = num_lists + 1 WHERE id = ?", repo.ID); err != nil {
		return err
	}

	return nil
}

// CreateRepository creates a repository for given user or organization.
func NewList(repo *Repository, list *List) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = newList(sess, repo, list); err != nil {
		return fmt.Errorf("newList: %v", err)
	}

	if err = sess.Commit(); err != nil {
		return fmt.Errorf("Commit: %v", err)
	}
	return nil
}


func (repo *Repository) getLists(e Engine) (_ []*List, err error) {
	rawLists := make([]*List, 0, repo.NumLists)
	lists := make([]*List, 0, repo.NumLists)
	if repo.NumLists> 0 {
		if err = e.Where("repo_id = ?", repo.ID).OrderBy("position").Find(&rawLists); err != nil {
			return nil, err
		}
		for _, raw := range rawLists {
			if list, err := GetListByID(raw.ID); err != nil {
				return make([]*List, 0, 0), err
			} else {
				lists = append(lists, list)
			}
		}
	}

	return lists, nil
}

func (repo *Repository) GetLists() (_ []*List, err error) {
	return repo.getLists(x)
}

func getListByRepoID(e Engine, id int64, repoID int64) (*List, error) {
	list := &List{
		ID 		:id,	
		RepoID	:repoID,
	}
	has, err := e.Get(list)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrRepoNotExist{0, repoID, ""}
	}
	return list, err
}

// GetRepositoryByID returns the repository by given id if exists.
func GetListByRepoID(id int64, repoID int64) (*List, error) {
	return getListByRepoID(x, id, repoID)
}

func updateList(e Engine, repo *Repository, list *List) error {
	_, err := e.Id(list.ID).AllCols().Update(list)
	return err
}

// UpdateMilestone updates information of given milestone.
func UpdateList(repo *Repository, list *List) error {
	return updateList(x, repo, list)
}

func getListByID(e Engine, id int64) (*List, error) {
	list := new(List)
	has, err := e.Id(id).Get(list)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrListNotExist{id, 0, 0}
	}
	return list, list.LoadCards()
}

// GetIssueByID returns an issue by given ID.
func GetListByID(id int64) (*List, error) {
	return getListByID(x, id)
}

// GetIssueByIndex returns raw issue without loading attributes by index in a repository.
func GetRawListByIndex(repoID, index int64) (*List, error) {
	list := &List{
		RepoID: 	repoID,
		Position:   index,
	}
	has, err := x.Get(list)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrListNotExist{0, repoID, index}
	}
	return list, nil
}

// GetIssueByIndex returns raw issue without loading attributes by index in a repository.
func GetRawListByID(listID int64) (*List, error) {
	list := &List{
		ID: listID,
	}
	has, err := x.Get(list)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrListNotExist{0, 0, listID}
	}
	return list, nil
}

// GetIssueByIndex returns issue by index in a repository.
func getListByIndex(e Engine, repoID, ListIndex int64) (*List, error) {
	list, err := GetRawListByIndex(repoID, ListIndex)
	if err != nil {
		return nil, err
	}
	return list, list.LoadCards()
}

// GetIssueByID returns an issue by given ID.
func GetListByIndex(repoID, ListIndex int64) (*List, error) {
	return getListByIndex(x, repoID, ListIndex)
}

// DeleteMilestoneByRepoID deletes a milestone from a repository.
func DeleteListByRepoID(repoID, ListIndex int64) error {
	list, err := GetListByIndex(repoID, ListIndex)
	if err != nil {
		if IsErrListNotExist(err) {
			return nil
		}
		return err
	}

	if list.NumCards > 0{
		if err := list.Empty(); err != nil{
			return err
		}
	}

	repo, err := GetRepositoryByID(list.RepoID)
	if err != nil {
		return err
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Id(list.ID).Delete(new(List)); err != nil {
		return err
	}

	repo.NumLists = int(countRepoLists(sess, repo.ID))
	if _, err = sess.Id(repo.ID).AllCols().Update(repo); err != nil {
		return err
	}
	return sess.Commit()
}




//  _________                  .___
//  \_   ___ \_____ _______  __| _/
//  /    \  \/\__  \\_  __ \/ __ | 
//  \     \____/ __ \|  | \/ /_/ | 
//   \______  (____  /__|  \____ | 
//          \/     \/           \/ 

type Card struct {
	ID            int64  `xorm:"pk autoincr"`
	ListID        int64  
	AssigneeID    int64  
	Assignee      *User  `xorm:"-"`
	Description   string `xorm:"TEXT"`
	Position	  int64  
	State 		  int
	LimitDate 	  time.Time `xorm:"-"`
	LimitDateUnix int64
	Created       time.Time `xorm:"-"`
	CreatedUnix   int64
	Updated       time.Time `xorm:"-"`
	UpdatedUnix   int64
}

func (card *Card) BeforeInsert() {
	card.CreatedUnix = time.Now().Unix()
	card.UpdatedUnix = card.CreatedUnix
}

func (card *Card) BeforeUpdate() {
	card.UpdatedUnix = time.Now().Unix()
}

func (card *Card) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
		case "created_unix":
			card.Created = time.Unix(card.CreatedUnix, 0).Local()
		case "updated_unix":
			card.Updated = time.Unix(card.UpdatedUnix, 0)
	}
}

func (card *Card) APIFormat() *api.Card {
	apiCard := &api.Card {
		ID:       card.ID,
		List:     card.ListID,
		Index:    card.Position,
		Body:     card.Description,
		State:    card.State,
		Limit: 	  card.LimitDateUnix,
		Created:  card.Created,
		Updated:  card.Updated,
	}

	if card.Assignee != nil {
		apiCard.Assignee = card.Assignee.APIFormat()
	}

	return apiCard
}


func countListCards(e Engine, listID int64) int64 {
	count, _ := e.Where("list_id = ?", listID).Count(new(Card))
	return count
}

// CountRepoMilestones returns number of milestones in given repository.
func CountListCards(listID int64) int64 {
	return countListCards(x, listID)
}

func (card *Card) loadAttributes(e Engine) (err error) {
	if card.Assignee == nil && card.AssigneeID > 0 {
		card.Assignee, err = getUserByID(e, card.AssigneeID)
		if err != nil {
			return fmt.Errorf("getUserByID.(assignee) [%d]: %v", card.AssigneeID, err)
		}
	}
	return nil
}

func (card *Card) LoadAttributes() error {
	return card.loadAttributes(x)
}

func newCard(e *xorm.Session, repo *Repository, card *Card) (err error) {

	if card.AssigneeID > 0 {
		assignee, err := getUserByID(e, card.AssigneeID)
		if err != nil && !IsErrUserNotExist(err) {
			return fmt.Errorf("getUserByID: %v", err)
		}

		// Assume assignee is invalid and drop silently.
		card.AssigneeID = 0
		if assignee != nil {
			valid, err := hasAccess(e, assignee, repo, ACCESS_MODE_WRITE)
			if err != nil {
				return fmt.Errorf("hasAccess [user_id: %d, repo_id: %d]: %v", assignee.ID, repo.ID, err)
			}
			if valid {
				card.AssigneeID = assignee.ID
				card.Assignee = assignee
			}
		}
	}

	if _, err = e.Insert(card); err != nil {
		return err
	}

	if _, err = e.Exec("UPDATE `list` SET num_cards = num_cards + 1 WHERE id = ?", card.ListID); err != nil {
		return err
	}

	return card.loadAttributes(e)
}

// Newcard creates new card with labels for repository.
func NewCard(repo *Repository, card *Card) (err error) {
	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if err = newCard(sess, repo, card); err != nil {
		return fmt.Errorf("newCard: %v", err)
	}

	if err = sess.Commit(); err != nil {
		return fmt.Errorf("Commit: %v", err)
	}

	/*if err = card.MailParticipants(); err != nil {
		log.Error(4, "MailParticipants: %v", err)
	}*/

	return nil
}

func getCardByListID(e Engine, id int64, listID int64) (*Card, error) {
	card := &Card{
		ID 		:id,	
		ListID	:listID,
	}
	has, err := e.Get(card)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrCardNotExist{0, listID, 0}
	}
	return card, err
}

// GetRepositoryByID returns the repository by given id if exists.
func GetCardByListID(id int64, listID int64) (*Card, error) {
	return getCardByListID(x, id, listID)
}

func updateCard(e Engine, c *Card) error {
	_, err := e.Id(c.ID).AllCols().Update(c)
	return err
}

// UpdateMilestone updates information of given milestone.
func UpdateCard(c *Card) error {
	return updateCard(x, c)
}

// DeleteMilestoneByRepoID deletes a milestone from a repository.
func (list *List) empty(e Engine) error {
	if list.Cards != nil {
		for _, card := range list.Cards {
			if _, err := e.Delete(&Card{ID: card.ID, ListID: card.ListID}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (list *List) Empty() error {
	return list.empty(x)
}

// GetIssueByIndex returns raw issue without loading attributes by index in a repository.
func GetRawCardByIndex(listID, index int64) (*Card, error) {
	card := &Card{
		ListID: 	listID,
		Position:   index,
	}
	has, err := x.Get(card)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrCardNotExist{0, listID, index}
	}
	return card, nil
}

// GetIssueByIndex returns issue by index in a repository.
func getCardByIndex(e Engine, listID, CardIndex int64) (*Card, error) {
	card, err := GetRawCardByIndex(listID, CardIndex)
	if err != nil {
		return nil, err
	}
	return card, card.LoadAttributes()
}

// GetIssueByID returns an issue by given ID.
func GetCardByIndex(listID, CardIndex int64) (*Card, error) {
	return getCardByIndex(x, listID, CardIndex)
}

// DeleteMilestoneByRepoID deletes a milestone from a repository.
func DeleteCardByListID(listID int64, id int64, repoID int64) error {
	card, err := GetCardByID(id)
	if err != nil {
		if IsErrListNotExist(err) {
			return nil
		}
		return err
	}

	list, err := GetListByRepoID(listID, repoID)
	if err != nil {
		return err
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Id(card.ID).Delete(new(Card)); err != nil {
		return err
	}

	list.NumCards = int(countListCards(sess, list.ID))
	if _, err = sess.Id(list.ID).AllCols().Update(list); err != nil {
		return err
	}
	return sess.Commit()
}

func getCardByID(e Engine, id int64) (*Card, error) {
	card := new(Card)
	has, err := e.Id(id).Get(card)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrCardNotExist{id, 0, 0}
	}
	return card, card.LoadAttributes()
}

// GetIssueByID returns an issue by given ID.
func GetCardByID(id int64) (*Card, error) {
	return getCardByID(x, id)
}

func (list *List) getCards(e Engine) (err error) {
	list.Cards = make([]*Card, 0, list.NumCards)
	return e.Where("list_id=?", list.ID).Asc("position").Find(&list.Cards)
}

// GetLabelsByIssueID returns all labels that belong to given issue by ID.
func (list *List) GetCards() error {
	return list.getCards(x)
}

func (card *Card) moveCard(e Engine, newListID int64) error {
	_, err := GetRawListByID(card.ListID)
	if err != nil {
		if IsErrListNotExist(err) {
			return nil
		}
		return err
	}

	_, err = GetRawListByID(newListID)
	if err != nil {
		if IsErrListNotExist(err) {
			return nil
		}
		return err
	}

	if _, err = e.Exec("UPDATE `list` SET num_cards = num_cards - 1 WHERE id = ?", card.ListID); err != nil {
		return err
	}

	if _, err = e.Exec("UPDATE `list` SET num_cards = num_cards + 1 WHERE id = ?", newListID); err != nil {
		return err
	}

	card.ListID = newListID
	return nil

}

func (card *Card) MoveCard(newListID int64) error {
	return card.moveCard(x, newListID)
}