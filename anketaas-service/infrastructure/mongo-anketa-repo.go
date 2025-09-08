package infrastructure

import "go.mongodb.org/mongo-driver/v2/mongo"

type AnketaRepo struct {
	collection *mongo.Collection
}

func NewAnketaRepo(db *mongo.Client) AnketaRepo {
	return AnketaRepo{
		db.Database("main").Collection("anketas"),
	}
}
