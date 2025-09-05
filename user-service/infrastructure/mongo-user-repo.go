package infrastructure

import (
	"context"
	"time"
	"user-service/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoUserRepo struct {
	collection *mongo.Collection
}

func NewMongoRepo(db *mongo.Client) *MongoUserRepo {
	return &MongoUserRepo{
		db.Database("main").Collection("users"),
	}
}

func (m *MongoUserRepo) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*10)
}

func (m *MongoUserRepo) Create(user domain.User) error {
	ctx, cancel := m.GetContext()
	defer cancel()

	userDoc := bson.M{
		"id":            user.ID.String(),
		"login":         user.Login.String(),
		"email":         user.Email.String(),
		"phone_number":  user.PhoneNumber.String(),
		"password_hash": user.PasswordHash.String(),
	}

	_, err := m.collection.InsertOne(ctx, userDoc)
	return err
}

func (m *MongoUserRepo) Update(id uuid.UUID, update domain.UserUpdate) error {
	ctx, cancel := m.GetContext()
	defer cancel()

	_, err := m.collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update.FieldsToUpdate})
	return err
}

func (m *MongoUserRepo) Delete(id uuid.UUID) error {
	ctx, cancel := m.GetContext()
	defer cancel()

	_, err := m.collection.DeleteOne(ctx, bson.M{"id": id.String()})
	return err
}

func (m *MongoUserRepo) FindByID(id uuid.UUID) (domain.User, error) {
	ctx, cancel := m.GetContext()
	defer cancel()

	var user domain.User
	err := m.collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	return user, err
}

func (m *MongoUserRepo) FindByLogin(login string) (domain.User, error) {
	ctx, cancel := m.GetContext()
	defer cancel()

	var user domain.User
	err := m.collection.FindOne(ctx, bson.M{"login": login}).Decode(&user)
	return user, err
}

func (m *MongoUserRepo) ExistsByEmail(email string) (bool, error) {
	return m.existsByField("email", email)
}

func (m *MongoUserRepo) ExistsByLogin(login string) (bool, error) {
	return m.existsByField("login", login)
}

func (m *MongoUserRepo) ExistsByPhone(phone string) (bool, error) {
	return m.existsByField("phone_number", phone)
}

func (m *MongoUserRepo) existsByField(field, value string) (bool, error) {
	ctx, cancel := m.GetContext()
	defer cancel()

	count, err := m.collection.CountDocuments(ctx, bson.M{field: value})
	return count > 0, err
}
