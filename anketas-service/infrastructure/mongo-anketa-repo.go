package infrastructure

import (
	"anketas-service/domain"
	errs "anketas-service/errors"
	"context"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoAnketaRepo struct {
	collection *mongo.Collection
}

func NewAnketaRepo(db *mongo.Client) *MongoAnketaRepo {
	return &MongoAnketaRepo{
		db.Database("main").Collection("anketas"),
	}
}

func (r *MongoAnketaRepo) Create(ctx context.Context, anketa domain.Anketa) error {

	doc := bson.M{
		"id":               anketa.ID,
		"username":         anketa.Username,
		"gender":           anketa.Gender,
		"preferred_gender": anketa.PreferredGender,
		"description":      anketa.Description,
		"tags":             anketa.Tags,
		"photos":           anketa.Photos,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		log.Println("Не удалось создать анкету", err)
		return err
	}

	return nil
}

func (r *MongoAnketaRepo) Update(ctx context.Context, id uuid.UUID, updateData map[string]any) error {

	filter := bson.M{"_id": id.String()}
	update := bson.M{"$set": updateData}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Не удалось обновить анкету", err)
		return err
	}

	if result.MatchedCount == 0 {
		return errs.InternalServerError
	}

	return nil
}

func (r *MongoAnketaRepo) Delete(ctx context.Context, id uuid.UUID) error {

	filter := bson.M{"id": id.String()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("Произошла ошибка при удалении анкеты", err)
		return err
	}

	if result.DeletedCount == 0 {
		return errs.InternalServerError
	}

	return nil
}

func (r *MongoAnketaRepo) FindByID(ctx context.Context, id uuid.UUID) (domain.Anketa, error) {

	filter := bson.M{"id": id.String()}

	var anketa domain.Anketa
	err := r.collection.FindOne(ctx, filter).Decode(&anketa)
	if err != nil {
		log.Println("Не удалось найти пользователя по айди", id, err)
		return domain.Anketa{}, err
	}

	return anketa, nil
}

func (r *MongoAnketaRepo) GetAnketas(ctx context.Context, pref domain.PreferredAnketaGender, limit int) ([]domain.Anketa, error) {

	cursor, err := r.collection.Find(ctx, bson.M{"preferred_gender": pref})
	if err != nil {
		return []domain.Anketa{}, err
	}

	var anketas []domain.Anketa
	count := 0
	for cursor.Next(ctx) && count <= 10 {
		count++

		var anketa domain.Anketa
		err := cursor.Decode(&anketa)
		if err != nil {
			return []domain.Anketa{}, err
		}

		anketas = append(anketas, anketa)
	}

	return anketas, nil
}
