package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Структура для ответа от anketas-service
type AnketaData struct {
	ID       string `json:"ID"`
	Username struct {
		Value string `json:"Value"`
	} `json:"Username"`
	Age      int `json:"Age"`
	Photos []struct {
		Url string `json:"Url"`
	} `json:"Photos"`
}

// Получение данных пользователя из anketas-service
func getUserDataFromAnketasService(userID string) (*AnketaData, error) {
	url := fmt.Sprintf("http://localhost:8004/anketa/%s", userID)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anketas-service returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var anketaData AnketaData
	
	if err := json.Unmarshal(body, &anketaData); err != nil {
		return nil, err
	}
	
	return &anketaData, nil
}

// Отправка сообщения
func sendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	message := Message{
		ID:         primitive.NewObjectID(),
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Text:       req.Text,
		Timestamp:  time.Now(),
		Read:       false,
	}

	_, err := messagesCollection.InsertOne(context.Background(), message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message sent successfully",
		"messageId": message.ID.Hex(),
	})
}

// Получение диалога между двумя пользователями
func getConversation(c *gin.Context) {
	senderId := c.Param("senderId")
	receiverId := c.Param("receiverId")

	filter := bson.M{
		"$or": []bson.M{
			{"senderId": senderId, "receiverId": receiverId},
			{"senderId": receiverId, "receiverId": senderId},
		},
	}

	opts := options.Find().SetSort(bson.D{{"timestamp", 1}})
	cursor, err := messagesCollection.Find(context.Background(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation"})
		return
	}
	defer cursor.Close(context.Background())

	var messages []Message
	if err = cursor.All(context.Background(), &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode messages"})
		return
	}

	c.JSON(http.StatusOK, ConversationResponse{
		Messages: messages,
	})
}

// Получение всех диалогов пользователя
func getUserConversations(c *gin.Context) {
	userId := c.Param("userId")

	// Получаем всех пользователей, с которыми есть диалоги
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"senderId": userId},
					{"receiverId": userId},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$senderId", userId}},
						"then": "$receiverId",
						"else": "$senderId",
					},
				},
				"lastMessage": bson.M{"$last": "$text"},
				"lastTimestamp": bson.M{"$last": "$timestamp"},
				"unreadCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$and": []bson.M{
									{"$eq": []interface{}{"$receiverId", userId}},
									{"$eq": []interface{}{"$read", false}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
			},
		},
		{
			"$sort": bson.M{"lastTimestamp": -1},
		},
	}

	cursor, err := messagesCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}
	defer cursor.Close(context.Background())

	var conversations []ConversationSummary
	for cursor.Next(context.Background()) {
		var result struct {
			ID            string    `bson:"_id"`
			LastMessage   string    `bson:"lastMessage"`
			LastTimestamp time.Time `bson:"lastTimestamp"`
			UnreadCount   int       `bson:"unreadCount"`
		}
		
		if err := cursor.Decode(&result); err != nil {
			continue
		}

	// Получаем данные пользователя из anketas-service
	userData, err := getUserDataFromAnketasService(result.ID)
	userName := "User " + result.ID
	userAge := 0
	userPhoto := ""
	
	if err == nil && userData != nil {
		userName = userData.Username.Value
		userAge = userData.Age
		if len(userData.Photos) > 0 {
			userPhoto = userData.Photos[0].Url
		}
	}

		conversations = append(conversations, ConversationSummary{
			UserID:      result.ID,
			UserName:   userName,
			UserAge:    userAge,
			UserPhoto:  userPhoto,
			LastMessage: result.LastMessage,
			Timestamp:  result.LastTimestamp,
			UnreadCount: result.UnreadCount,
		})
	}

	c.JSON(http.StatusOK, UserConversationsResponse{
		Conversations: conversations,
	})
}

// Отметить сообщение как прочитанное
func markAsRead(c *gin.Context) {
	messageId := c.Param("messageId")
	
	objectId, err := primitive.ObjectIDFromHex(messageId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$set": bson.M{"read": true}}

	_, err = messagesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message marked as read",
	})
}
