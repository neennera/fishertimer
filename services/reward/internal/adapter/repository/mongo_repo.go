package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

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
	UserID    string        `bson:"user_id"`
	CycleID   string        `bson:"cycle_id"`
	ItemID    bson.ObjectID `bson:"item_id"`
	AwardedAt time.Time     `bson:"awarded_at"`
}


func (r *MongoRepository) Award(ctx context.Context, reward *domain.UnlockedReward) error {
	itemID, err := bson.ObjectIDFromHex(reward.ItemID)
	if err != nil {
		return domain.ErrInvalid
	}

	awardedAt := reward.AwardedAt
	if awardedAt.IsZero() {
		awardedAt = time.Now().UTC()
	}

	_, err = r.unlock.InsertOne(ctx, userRewardDoc{
		UserID:    reward.UserID,
		CycleID:   reward.CycleID,
		ItemID:    itemID,
		AwardedAt: awardedAt,
	})
	return err
}

func (r *MongoRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "reward_items",
			"localField":   "item_id",
			"foreignField": "_id",
			"as":           "item",
		}}},
		{{Key: "$unwind", Value: "$item"}},
		{{Key: "$sort", Value: bson.M{"awarded_at": 1}}},
	}

	cursor, err := r.unlock.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []struct {
		ID        bson.ObjectID `bson:"_id"`
		UserID    string        `bson:"user_id"`
		CycleID   string        `bson:"cycle_id"`
		AwardedAt time.Time     `bson:"awarded_at"`
		Item      struct {
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
		rewards = append(rewards, domain.UnlockedReward{
			UserRewardID: d.ID.Hex(),
			ItemID:       d.Item.ID.Hex(),
			UserID:       d.UserID,
			CycleID:      d.CycleID,
			ItemName:     d.Item.ItemName,
			Category:     d.Item.Category,
			Rarity:       d.Item.Rarity,
			BaseWeight:   d.Item.BaseWeight,
			ScoreValue:   d.Item.ScoreValue,
			AssetURL:     d.Item.AssetURL,
			AwardedAt:    d.AwardedAt,
		})
	}
	return rewards, nil
}
