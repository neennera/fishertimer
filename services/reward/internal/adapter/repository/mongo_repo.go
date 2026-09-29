package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

// MongoRepository persists rewards in reward_db (reward_items, user_rewards)
// - see database/schemas/001_create_reward_collections.js.
type MongoRepository struct {
	items  *mongo.Collection
	unlock *mongo.Collection
}

func NewMongo(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		items:  db.Collection("reward_items"),
		unlock: db.Collection("user_rewards"),
	}
}

// userRewardDoc mirrors a user_rewards document.
type userRewardDoc struct {
	UserID      string        `bson:"user_id"`
	DisplayName string        `bson:"display_name,omitempty"`
	CycleID     string        `bson:"cycle_id"`
	ItemID      bson.ObjectID `bson:"item_id"`
	AwardedAt   time.Time     `bson:"awarded_at"`
}

func (r *MongoRepository) Award(ctx context.Context, reward *domain.UnlockedReward) error {
	var itemID bson.ObjectID
	var err error
	if reward.ItemID != "" {
		itemID, err = bson.ObjectIDFromHex(reward.ItemID)
		if err != nil {
			itemID = bson.NewObjectID()
		}
	} else {
		itemID = bson.NewObjectID()
	}

	awardedAt := reward.AwardedAt
	if awardedAt.IsZero() {
		awardedAt = time.Now().UTC()
	}

	res, err := r.unlock.InsertOne(ctx, userRewardDoc{
		UserID:      reward.UserID,
		DisplayName: reward.DisplayName,
		CycleID:     reward.CycleID,
		ItemID:      itemID,
		AwardedAt:   awardedAt,
	})
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		reward.UserRewardID = oid.Hex()
		reward.ID = oid.Hex()
	}
	return nil
}

func (r *MongoRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	match := bson.M{}
	if userID != "" {
		match["user_id"] = userID
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "reward_items",
			"localField":   "item_id",
			"foreignField": "_id",
			"as":           "item",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$item",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$sort", Value: bson.M{"awarded_at": -1}}},
	}

	cursor, err := r.unlock.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []struct {
		ID          bson.ObjectID `bson:"_id"`
		UserID      string        `bson:"user_id"`
		DisplayName string        `bson:"display_name"`
		CycleID     string        `bson:"cycle_id"`
		AwardedAt   time.Time     `bson:"awarded_at"`
		Item        *struct {
			ID         bson.ObjectID `bson:"_id"`
			ItemName   string        `bson:"item_name"`
			Category   string        `bson:"category"`
			Rarity     string        `bson:"rarity"`
			BaseWeight float64       `bson:"base_weight"`
			ScoreValue int           `bson:"score_value"`
			AssetURL   string        `bson:"asset_url"`
		} `bson:"item"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	rewards := make([]domain.UnlockedReward, 0, len(docs))
	for _, d := range docs {
		item := domain.UnlockedReward{
			UserRewardID: d.ID.Hex(),
			ID:           d.ID.Hex(),
			UserID:       d.UserID,
			DisplayName:  d.DisplayName,
			CycleID:      d.CycleID,
			AwardedAt:    d.AwardedAt,
		}
		if d.Item != nil {
			item.ItemID = d.Item.ID.Hex()
			item.ItemName = d.Item.ItemName
			item.Species = d.Item.ItemName
			item.Category = d.Item.Category
			item.Rarity = d.Item.Rarity
			item.BaseWeight = d.Item.BaseWeight
			item.ScoreValue = d.Item.ScoreValue
			item.AssetURL = d.Item.AssetURL
		}
		rewards = append(rewards, item)
	}
	return rewards, nil
}

// GetLastUpdate returns the awarded_at of the most recently inserted reward.
// The leaderboard service calls this to decide if its cache is stale.
func (r *MongoRepository) GetLastUpdate(ctx context.Context) (time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "awarded_at", Value: -1}})
	var doc userRewardDoc
	err := r.unlock.FindOne(ctx, bson.D{}, opts).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return doc.AwardedAt, nil
}
