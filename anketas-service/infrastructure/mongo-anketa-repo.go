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
	LikedBy         []string `bson:"liked_by"`
}

const AGE_DIFFERENCE = 2

func (r *MongoAnketaRepo) Create(ctx context.Context, anketa domain.Anketa) error {

	log.Println("Репозиторий начал создание анкеты")

	tags := make([]string, 0, 10)
	photos := make([]string, 0, 3)
	var likedBy []string
	for _, photo := range anketa.Photos {
		photos = append(photos, photo.Url)
	}
	for _, tag := range anketa.Tags {
		tags = append(tags, tag.Value)
	}
	for _, tag := range anketa.LikedBy {
		likedBy = append(likedBy, tag.String())
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
		"liked_by":         likedBy,
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
	log.Printf("=== GetAnketas: ищем анкету пользователя по ID: %s ===", id.String())
	
	anketa, err := r.FindByID(ctx, id)
	if err != nil {
		log.Printf("ОШИБКА: не удалось найти анкету пользователя по ID %s: %v", id.String(), err)
		log.Printf("Это может означать, что анкета еще не была сохранена в MongoDB")
		return []domain.Anketa{}, err
	}
	
	log.Printf("Анкета пользователя найдена: ID=%s, Gender=%s, PreferredGender=%s", 
		anketa.ID.String(), anketa.Gender.Value, anketa.PreferredGender.Value)

	var cursor *mongo.Cursor

	log.Printf("=== ПОИСК АНКЕТ ===")
	log.Printf("Анкета пользователя: ID=%s, Gender=%s", id.String(), anketa.Gender.Value)
	log.Printf("PreferredGender: %s", pref.Value)

	if pref.Value == domain.PreferredBoth {
		log.Printf("Ищем всех (PreferredBoth)")
		cursor, err = r.collection.Find(ctx, bson.M{})
		if err != nil {
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}
	} else {
		// Преобразуем Gender пользователя в формат для preferred_gender
		// "Мужчина" -> "Мужчин", "Женщина" -> "Женщин"
		var userPreferredGender string
		if anketa.Gender.Value == domain.Man {
			userPreferredGender = domain.PreferredMan
		} else if anketa.Gender.Value == domain.Woman {
			userPreferredGender = domain.PreferredWoman
		} else {
			userPreferredGender = anketa.Gender.Value
		}
		
		// Преобразуем PreferredGender в формат для gender
		// "Мужчин" -> "Мужчина", "Женщин" -> "Женщина"
		var targetGender string
		if pref.Value == domain.PreferredMan {
			targetGender = domain.Man
		} else if pref.Value == domain.PreferredWoman {
			targetGender = domain.Woman
		} else {
			targetGender = pref.Value + "а"
		}
		
		log.Printf("Ищем: preferred_gender=%s, gender=%s", userPreferredGender, targetGender)
		
		cursor, err = r.collection.Find(ctx,
			bson.M{
				"preferred_gender": userPreferredGender,
				"gender": targetGender,
			})
		if err != nil {
			log.Printf("Ошибка поиска в БД: %v", err)
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}
	}

	var anketas []domain.Anketa
	count := 0
	for cursor.Next(ctx) {
		count++
		var anketaDTO anketaDTO
		err := cursor.Decode(&anketaDTO)
		if err != nil {
			log.Printf("Ошибка декодирования анкеты: %v", err)
			return []domain.Anketa{}, fmt.Errorf("Ошибка на стороне сервера, просим прощения, мы уже работаем над этим")
		}

		anketa, err := anketaDTOtoDomainAnketa(anketaDTO)
		if err != nil {
			log.Printf("Ошибка преобразования анкеты: %v", err)
			return []domain.Anketa{}, err
		}

		anketas = append(anketas, anketa)
	}
	
	log.Printf("Найдено анкет в БД: %d", count)
	
	// Фильтруем анкеты: исключаем свою анкету, тех, кого уже лайкнули, и тех, кто уже лайкнул нас
	var filteredAnketas []domain.Anketa
	for _, a := range anketas {
		// Исключаем свою анкету
		if a.ID == anketa.ID {
			continue
		}
		
		// Исключаем тех, кого уже лайкнули (кто в LikedBy текущего пользователя)
		alreadyLiked := false
		for _, likedId := range anketa.LikedBy {
			if likedId == a.ID {
				alreadyLiked = true
				break
			}
		}
		if alreadyLiked {
			continue
		}
		
		// Исключаем тех, кто уже лайкнул нас (текущий пользователь в их LikedBy)
		likedByThem := false
		for _, likedId := range a.LikedBy {
			if likedId == anketa.ID {
				likedByThem = true
				break
			}
		}
		if likedByThem {
			continue
		}
		
		filteredAnketas = append(filteredAnketas, a)
	}
	
	log.Printf("Анкеты до метчинга: %d штук (после фильтрации по LikedBy: %d)", len(anketas), len(filteredAnketas))
	for i, a := range filteredAnketas {
		log.Printf("  Анкета %d: ID=%s, Gender=%s, Age=%d, Tags=%d", i+1, a.ID.String(), a.Gender.Value, a.Age.Int(), len(a.Tags))
	}

	afterMatch := matchAnketas(anketa, filteredAnketas)

	log.Printf("Анкеты после метчинга: %d штук", len(afterMatch))
	for i, a := range afterMatch {
		log.Printf("  Анкета %d после метчинга: ID=%s, Gender=%s, Age=%d", i+1, a.ID.String(), a.Gender.Value, a.Age.Int())
	}

	return afterMatch, nil
}

func anketaDTOtoDomainAnketa(a anketaDTO) (domain.Anketa, error) {

	// Убираем @ если он есть в начале username
	cleanUsername := a.Username
	if len(cleanUsername) > 0 && cleanUsername[0] == '@' {
		cleanUsername = cleanUsername[1:]
	}

	tagsArray := make([]domain.Tag, 0, 10)
	photosArray := make([]domain.Photo, 0, 3)
	var likedBy []uuid.UUID
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
	username, err := valueObjects.NewUsername(cleanUsername)
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
	for _, id := range a.LikedBy {
		tag, err := uuid.Parse(id)
		if err != nil {
			log.Println("Ошибка с uuid", err)
			return domain.Anketa{}, err
		}
		likedBy = append(likedBy, tag)
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
		LikedBy:         likedBy,
	}, nil
}

func matchAnketas(userAnketa domain.Anketa, anketas []domain.Anketa) []domain.Anketa {
	log.Printf("=== НАЧАЛО МЕТЧИНГА ===")
	log.Printf("Анкета пользователя: ID=%s, Gender=%s, Age=%d, Tags=%d", 
		userAnketa.ID.String(), userAnketa.Gender.Value, userAnketa.Age.Int(), len(userAnketa.Tags))
	log.Printf("Всего анкет для проверки: %d", len(anketas))

	tagMap := make(map[domain.Tag]struct{})
	for _, tag := range userAnketa.Tags {
		tagMap[tag] = struct{}{}
	}
	log.Printf("Теги пользователя: %d", len(tagMap))

	type anketaCountPair struct {
		anketa *domain.Anketa
		count  int
	}

	var pairs []anketaCountPair

	// переписать на горутины
	for i, anketa := range anketas {
		var count int

		// Исключаем анкету самого пользователя
		if anketa.ID == userAnketa.ID {
			log.Printf("Анкета %d: пропущена (это анкета самого пользователя)", i+1)
			continue
		}

		// Проверка возраста
		ageDiff := userAnketa.Age.Int() - anketa.Age.Int()
		if ageDiff < 0 {
			ageDiff = -ageDiff
		}
		if ageDiff > AGE_DIFFERENCE {
			log.Printf("Анкета %d: пропущена (разница в возрасте %d > %d)", i+1, ageDiff, AGE_DIFFERENCE)
			continue
		}

		// Подсчет совпадающих тегов
		for _, tag := range anketa.Tags {
			_, ok := tagMap[tag]
			if ok == true {
				count++
			}
		}

		log.Printf("Анкета %d: прошла фильтры, совпадающих тегов: %d", i+1, count)
		pairs = append(pairs, anketaCountPair{&anketa, count})
	}

	log.Printf("Анкет прошло фильтры: %d", len(pairs))

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	var result []domain.Anketa
	for _, pair := range pairs {
		result = append(result, *pair.anketa)
	}

	log.Printf("=== КОНЕЦ МЕТЧИНГА, возвращаем %d анкет ===", len(result))
	return result
}
