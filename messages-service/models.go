package main

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID   string             `bson:"senderId" json:"senderId"`
	ReceiverID string             `bson:"receiverId" json:"receiverId"`
	Text       string             `bson:"text" json:"text"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	Read       bool               `bson:"read" json:"read"`
}

type SendMessageRequest struct {
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	Text       string `json:"text"`
}

type ConversationResponse struct {
	Messages []Message `json:"messages"`
}

type UserConversationsResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}

type ConversationSummary struct {
	UserID      string    `json:"userId"`
	UserName    string    `json:"userName"`
	UserAge     int       `json:"userAge"`
	UserPhoto   string    `json:"userPhoto"`
	LastMessage string    `json:"lastMessage"`
	Timestamp   time.Time `json:"timestamp"`
	UnreadCount int       `json:"unreadCount"`
}
