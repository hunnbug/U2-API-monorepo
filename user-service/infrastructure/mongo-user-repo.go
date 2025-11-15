package infrastructure

import (
	"context"
	"log"
	"time"
	"user-service/domain"
	"user-service/valueObjects"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UserDTO - DTO для MongoDB
type UserDTO struct {
	ID           string `bson:"id"`
	Login        string `bson:"login"`
	Email        string `bson:"email"`
	PhoneNumber  string `bson:"phone_number"`
	PasswordHash string `bson:"password_hash"`
}

// convertDTOToUser - конвертирует UserDTO в domain.User
func convertDTOToUser(dto UserDTO) (domain.User, error) {
	// Парсим UUID
	userID, err := uuid.Parse(dto.ID)
	if err != nil {
		return domain.User{}, err
	}

	// Создаем value objects
	loginVO, err := valueObjects.NewLogin(dto.Login)
	if err != nil {
		return domain.User{}, err
	}

	emailVO, err := valueObjects.NewEmail(dto.Email)
	if err != nil {
		return domain.User{}, err
	}

	phoneVO, err := valueObjects.NewPhone(dto.PhoneNumber)
	if err != nil {
		return domain.User{}, err
	}

	passwordVO, err := valueObjects.NewPassword(dto.PasswordHash)
	if err != nil {
		return domain.User{}, err
	}

	// Создаем domain.User
	user := domain.User{
		ID:           userID,
		Login:        loginVO,
		Email:        emailVO,
		PhoneNumber:  phoneVO,
		PasswordHash: passwordVO,
	}

	return user, nil
}

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

	changed := bson.M{}
	for k, v := range update.FieldsToUpdate {
		log.Printf("обновляется поле %s на знеачение %s\n", k, v)
		changed[k] = v
	}
	updateBson := bson.M{"$set": changed}

	for k, v := range changed {
		log.Printf("key %s value %s\n", k, v)
	}

	for k, v := range updateBson {
		log.Printf("ket %s value %s", k, v)
	}

	log.Println("обновляем документ с id:", id)

	result, err := m.collection.UpdateOne(ctx, bson.M{"id": id.String()}, updateBson)
	log.Printf("Найдено документов %d, обновлено %d\n", result.MatchedCount, result.ModifiedCount)
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

	var userDTO UserDTO
	err := m.collection.FindOne(ctx, bson.M{"id": id.String()}).Decode(&userDTO)
	if err != nil {
		return domain.User{}, err
	}

	log.Printf("Найден пользователь DTO: %+v\n", userDTO)

	// Конвертируем DTO в domain.User
	user, err := convertDTOToUser(userDTO)
	if err != nil {
		return domain.User{}, err
	}

	log.Println("\n\nКонвертированный пользователь:", user)

	return user, nil
}

func (m *MongoUserRepo) FindByLogin(login string) (domain.User, error) {
	ctx, cancel := m.GetContext()
	defer cancel()

	var userDTO UserDTO
	err := m.collection.FindOne(ctx, bson.M{"login": login}).Decode(&userDTO)
	if err != nil {
		return domain.User{}, err
	}

	// Конвертируем DTO в domain.User
	user, err := convertDTOToUser(userDTO)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
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
