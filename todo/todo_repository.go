package todo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type todoRepo struct {
	collectionName string
	collection     *mongo.Collection
}

func NewTodoRepo(collNames map[string]int, db *mongo.Database) (TodoRepository, error) {
	var collectionName = "todos"

	if _, exists := collNames[collectionName]; !exists {
		if err := db.CreateCollection(context.TODO(), collectionName); err != nil {
			return nil, err
		}
		index := mongo.IndexModel {
			Keys: bson.D{
				{Key: "title", Value: 1},
				{Key: "active_at", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		}
		_, err := db.Collection(collectionName).Indexes().CreateOne(context.TODO(), index)
		if err != nil {
			return nil, err
		}
	}

	return &todoRepo{
		collectionName: collectionName,
		collection:     db.Collection(collectionName),
	}, nil
}

func (repository *todoRepo) Create(todo *Todo) (*Todo, error) {
	todo.CreatedAt = time.Now().UTC()
	result, err := repository.collection.InsertOne(context.TODO(), todo)
	if err != nil {
		return nil, err
	}
	todo.ID = result.InsertedID.(primitive.ObjectID)
	return todo, nil
}

func (repository *todoRepo) FindByID(id primitive.ObjectID) (*Todo, error) {
	var todo Todo
	err := repository.collection.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}
	return &todo, nil
}

func (repository *todoRepo) FindAll(pointers TodoPointers) ([]*Todo, error) {
	query := bson.D{}
	if pointers.Title != nil {
		query = append(query, bson.E{Key: "title", Value: *pointers.Title})
	}
	if pointers.Status != nil {
		query = append(query, bson.E{Key: "status", Value: *pointers.Status})
	}
	if pointers.ActiveAt != nil {
		var comparisonOperator string
		switch *pointers.ActiveAt.ComparisonOperator {
		case ComparisonOperatorEQ:
			comparisonOperator = "$eq"
		case ComparisonOperatorGT:
			comparisonOperator = "$gt"
		case ComparisonOperatorGTE:
			comparisonOperator = "$gte"
		case ComparisonOperatorLT:
			comparisonOperator = "$lt"
		case ComparisonOperatorLTE:
			comparisonOperator = "$lte"
		default:
			return nil, ErrUnknownComparisonOperator
		}
		query = append(query, bson.E{Key: "active_at", Value: bson.M{comparisonOperator: *pointers.ActiveAt.ActiveAt}})
	}

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})
	cursor, err := repository.collection.Find(context.TODO(), query, opts)
	if err != nil {
		return nil, err
	}
	todos := make([]*Todo, 0, cursor.RemainingBatchLength())
	for cursor.Next(context.TODO()) {
		var todo Todo
		err := cursor.Decode(&todo)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}
	return todos, nil
}

func (repository *todoRepo) Update(upd TodoPointers) error {
	filter := bson.D{{"_id", *upd.ID}}
	values := bson.D{}
	if upd.Title != nil {
		values = append(values, bson.E{Key: "title", Value: *upd.Title})
	}
	if upd.ActiveAt != nil {
		values = append(values, bson.E{Key: "active_at", Value: *upd.ActiveAt.ActiveAt})
	}
	if upd.Status != nil {
		values = append(values, bson.E{Key: "status", Value: *upd.Status})
	}

	if len(values) == 0 {
		return ErrNothingToUpdate
	}

	updatedAt := time.Now().UTC()
	values = append(values, bson.E{Key: "updated_at", Value: updatedAt})
	update := bson.D{{"$set", values}}
	result := repository.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return ErrNothingToUpdate
		}
		return result.Err()
	}
	return nil
}

func (repository *todoRepo) Delete(id primitive.ObjectID) error {
	result := repository.collection.FindOneAndDelete(context.TODO(), bson.D{{"_id", id}})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return ErrTodoNotFound
		}
		return result.Err()
	}
	return nil
}