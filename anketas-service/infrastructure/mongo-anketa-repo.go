package infrastructure

import (
	"anketas-service/domain"
	errs "anketas-service/errors"
	"anketas-service/valueObjects"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"

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

type anketaDTO struct {
	ID              string   `bson:"id"`
	Username        string   `bson:"username"`
	Age             int      `bson:"age"`
	Gender          string   `bson:"gender"`
	PreferredGender string   `bson:"preferred_gender"`
	Description     string   `bson:"description"`
	Tags            []string `bson:"tags"`
	Photos          []string `bson:"photos"`
}

const AGE_DIFFERENCE = 2

func (r *MongoAnketaRepo) Create(ctx context.Context, anketa domain.Anketa) error {

	log.Println("Репозиторий начал создание анкеты")

	tags := make([]string, 0, 10)
	photos := make([]string, 0, 3)
	for _, photo := range anketa.Photos {
		photos = append(photos, photo.Url)
	}
	for _, tag := range anketa.Tags {
		tags = append(tags, tag.Value)
	}
	doc := bson.M{
		"id":               anketa.ID.String(),
		"username":         anketa.Username.Value,
		"age":              anketa.Age.Int(),
		"gender":           anketa.Gender.Value,
		"preferred_gender": anketa.PreferredGender.Value,
		"description":      anketa.Description,
		"tags":             tags,
		"photos":           photos,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		log.Println("Не удалось создать анкету")
		return errors.New("Произошла внутренняя ошибка сервера при создании анкеты. Приносим свои извинения")
	}

	log.Println("Репозиторий успешно создал анкету")

	return nil
}

func (r *MongoAnketaRepo) Update(ctx context.Context, id uuid.UUID, updateData map[string]any) error {

	log.Println("Репозиторий начал обновление анкеты")
	filter := bson.M{"id": id.String()}
	update := bson.M{"$set": updateData}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Не удалось обновить анкету", err)
		return err
	}

	if result.MatchedCount == 0 {
		log.Println("Не найдено ни одной анкеты с таким айди:", id.String())
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

	var anketaDTO anketaDTO

	err := r.collection.FindOne(ctx, filter).Decode(&anketaDTO)
	if err != nil {
		log.Println("Не удалось найти пользователя по айди", id.String(), err)
		return domain.Anketa{}, errs.InternalServerError
	}

	anketa, err := anketaDTOtoDomainAnketa(anketaDTO)
	if err != nil {
		return domain.Anketa{}, err
	}

	return anketa, nil
}

func (r *MongoAnketaRepo) GetAnketas(ctx context.Context, pref domain.PreferredAnketaGender, id uuid.UUID) ([]domain.Anketa, error) {

	anketa, err := r.FindByID(ctx, id)
	if err != nil {
		log.Println("Произошла ошибка при получении пользователя по айди", id, "|", err)
		return []domain.Anketa{}, err
	}

	var cursor *mongo.Cursor

	if pref.Value == domain.PreferredBoth {
		cursor, err = r.collection.Find(ctx, bson.M{})
		if err != nil {
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}
	} else {
		cursor, err = r.collection.Find(ctx,
			bson.M{"preferred_gender": anketa.Gender.Value[:len(anketa.Gender.Value)-2], /*Это тоже пиздец позор*/
				"gender": pref.Value + "а" /*Это позорище ебать*/})
		// log.Printf("id: %s\nИщем гендер %s\nПредпочитаемый гендер %s", id.String(), pref.Value+"a", anketa.Gender.Value[:len(anketa.Gender.Value)-2])
		if err != nil {
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}
	}

	var anketas []domain.Anketa
	for cursor.Next(ctx) {

		var anketaDTO anketaDTO
		err := cursor.Decode(&anketaDTO)
		if err != nil {
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}

		anketa, err := anketaDTOtoDomainAnketa(anketaDTO)
		if err != nil {
			return []domain.Anketa{}, err
		}

		anketas = append(anketas, anketa)
	}

	log.Println("Анкеты до метчинга:", anketas)

	afterMatch := matchAnketas(anketa, anketas)

	log.Println("Анкеты после метчинга:", afterMatch)

	return afterMatch, nil
}

func anketaDTOtoDomainAnketa(a anketaDTO) (domain.Anketa, error) {

	a.Username = a.Username[1:]

	tagsArray := make([]domain.Tag, 0, 10)
	photosArray := make([]domain.Photo, 0, 3)
	for _, tag := range a.Tags {
		valueObjectTag, err := domain.NewTag(tag)
		if err != nil {
			log.Println("Неверный формат тэга")
			return domain.Anketa{}, err
		}
		tagsArray = append(tagsArray, valueObjectTag)
	}
	for _, photo := range a.Photos {
		valueObjectPhoto, err := domain.NewPhoto(photo)
		if err != nil {
			log.Println("Неверный формат ссылки на фото")
			return domain.Anketa{}, err
		}
		photosArray = append(photosArray, valueObjectPhoto)
	}
	username, err := valueObjects.NewUsername(a.Username)
	if err != nil {
		log.Println("Неверный формат юзернейма")
		return domain.Anketa{}, err
	}
	gender, err := domain.NewAnketaGender(a.Gender)
	if err != nil {
		log.Println("Неверный формат пола")
		return domain.Anketa{}, err
	}
	preferredGender, err := domain.NewPreferredAnketaGender(a.PreferredGender)
	if err != nil {
		log.Println("Неверный формат предпочитаемого пола в поиске")
		return domain.Anketa{}, err
	}
	anketaAge, err := domain.NewAge(a.Age)
	if err != nil {
		log.Println("Неверный возраст")
		return domain.Anketa{}, err
	}
	id, err := uuid.Parse(a.ID)
	if err != nil {
		log.Println("Ошибка с uuid", err)
		return domain.Anketa{}, err
	}

	return domain.Anketa{
		ID:              id,
		Username:        username,
		Age:             anketaAge,
		Gender:          gender,
		PreferredGender: preferredGender,
		Description:     a.Description,
		Tags:            tagsArray,
		Photos:          photosArray,
	}, nil
}

func matchAnketas(userAnketa domain.Anketa, anketas []domain.Anketa) []domain.Anketa {

	tagMap := make(map[domain.Tag]struct{})
	for _, tag := range userAnketa.Tags {
		tagMap[tag] = struct{}{}
	}

	type anketaCountPair struct {
		anketa *domain.Anketa
		count  int
	}

	var pairs []anketaCountPair

	// переписать на горутины
	for _, anketa := range anketas {
		var count int

		if anketa.Age.Int() < userAnketa.Age.Int()-AGE_DIFFERENCE ||
			anketa.Age.Int() > userAnketa.Age.Int()+AGE_DIFFERENCE {
			continue
		}

		for _, tag := range anketa.Tags {
			_, ok := tagMap[tag]
			if ok == true {
				count++
			}
		}

		pairs = append(pairs, anketaCountPair{&anketa, count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	var result []domain.Anketa
	for _, pair := range pairs {
		result = append(result, *pair.anketa)
	}

	return result
}
