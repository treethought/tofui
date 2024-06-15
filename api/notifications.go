package api

import (
	"context"
	"fmt"
	"time"
)

type NotificationsType string
type CastReactionObjType string

const (
	NotificationsTypeFollows NotificationsType = "follows"
	NotificationsTypeLikes   NotificationsType = "likes"
	NotificationsTypeRecasts NotificationsType = "recasts"
	NotificationsTypeMention NotificationsType = "mention"
	NotificationsTypeReply   NotificationsType = "reply"

	CastReactionObjTypeLikes   CastReactionObjType = "likes"
	CastReactionObjTypeRecasts CastReactionObjType = "recasts"
)

type NotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
	Next          struct {
		Cursor *string `json:"cursor"`
	}
}

type Notification struct {
	Object              string                 `json:"object"`
	MostRecentTimestamp time.Time              `json:"most_recent_timestamp"`
	Type                NotificationsType      `json:"type"`
	Cast                *Cast                   `json:"cast"`
	Follows             []FollowNotification   `json:"follows"`
	Reactions           []ReactionNotification `json:"reactions"`
}

type FollowNotification struct {
	Object string `json:"object"`
	User   User   `json:"user"`
}

type ReactionNotification struct {
	Object CastReactionObjType `json:"object"`
	Cast   NotificationCast    `json:"cast"`
	User   User                `json:"user"`
}

type NotificationCast struct {
	Cast // may be cast_dehydrated which only has hash, specifically for reactions
}

func (c *Client) GetNotifications(fid uint64, opts ...RequestOption) (*NotificationsResponse, error) {
	path := fmt.Sprintf("/notifications")

	opts = append(opts, WithFID(fid))

	var resp NotificationsResponse
	if err := c.doRequestInto(context.TODO(), path, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
