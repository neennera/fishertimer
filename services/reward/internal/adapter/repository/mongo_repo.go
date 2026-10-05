package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
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
	ID          bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`
	DisplayName string        `bson:"display_name,omitempty"`
	CycleID     string        `bson:"cycle_id"`
	ItemID      bson.ObjectID `bson:"item_id"`
	AwardedAt   time.Time     `bson:"awarded_at"`
}

func (r *MongoRepository) ListItems(ctx context.Context) ([]domain.RewardItem, error) {
	cursor, err := r.items.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []struct {
		ID         bson.ObjectID `bson:"_id"`
		ItemName   string        `bson:"item_name"`
		Category   string        `bson:"category"`
		Rarity     string        `bson:"rarity"`
		BaseWeight float64       `bson:"base_weight"`
		ScoreValue int           `bson:"score_value"`
		AssetURL   string        `bson:"asset_url"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	items := make([]domain.RewardItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, domain.RewardItem{
			ID:         d.ID.Hex(),
			ItemName:   d.ItemName,
			Category:   d.Category,
			Rarity:     d.Rarity,
			BaseWeight: d.BaseWeight,
			ScoreValue: d.ScoreValue,
			AssetURL:   d.AssetURL,
		})
	}
	return items, nil
}

// awardedRewardID is the _id of draw n of a cycle: the first 12 bytes of
// sha256(cycle_id + ":" + n). UC-09 E-1: a retried cycle produces the same
// _ids, so Mongo's built-in unique _id index rejects the duplicates, with no
// extra index or field on user_rewards.
func awardedRewardID(cycleID string, n int) bson.ObjectID {
	sum := sha256.Sum256([]byte(cycleID + ":" + strconv.Itoa(n)))
	var id bson.ObjectID
	copy(id[:], sum[:12])
	return id
}

// AwardMany inserts in order, so a duplicate _id stops the insert at that
// row. Draw 0 goes first, so a losing concurrent request stores nothing.
func (r *MongoRepository) AwardMany(ctx context.Context, rewards []domain.UnlockedReward) error {
	docs := make([]userRewardDoc, len(rewards))
	for i, reward := range rewards {
		itemID, err := bson.ObjectIDFromHex(reward.ItemID)
		if err != nil {
			return fmt.Errorf("reward: item_id %q is not a reward_items _id: %w", reward.ItemID, err)
		}
		awardedAt := reward.AwardedAt
		if awardedAt.IsZero() {
			awardedAt = time.Now().UTC()
		}
		docs[i] = userRewardDoc{
			ID:          awardedRewardID(reward.CycleID, i),
			UserID:      reward.UserID,
			DisplayName: reward.DisplayName,
			CycleID:     reward.CycleID,
			ItemID:      itemID,
			AwardedAt:   awardedAt,
		}
	}

	res, err := r.unlock.InsertMany(ctx, docs)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrAlreadyAwarded
	}
	if err != nil {
		return err
	}
	for i, id := range res.InsertedIDs {
		if oid, ok := id.(bson.ObjectID); ok {
			rewards[i].UserRewardID = oid.Hex()
			rewards[i].ID = oid.Hex()
		}
		rewards[i].AwardedAt = docs[i].AwardedAt
	}
	return nil
}

func (r *MongoRepository) ListByCycle(ctx context.Context, cycleID string) ([]domain.UnlockedReward, error) {
	return r.listJoined(ctx, bson.M{"cycle_id": cycleID}, bson.D{{Key: "awarded_at", Value: 1}, {Key: "_id", Value: 1}})
}

func (r *MongoRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	match := bson.M{}
	if userID != "" {
		match["user_id"] = userID
	}
	return r.listJoined(ctx, match, bson.D{{Key: "awarded_at", Value: -1}})
}

// listJoined returns the matching user_rewards rows joined with their
// reward_items entry.
func (r *MongoRepository) listJoined(ctx context.Context, match bson.M, sort bson.D) ([]domain.UnlockedReward, error) {
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
		{{Key: "$sort", Value: sort}},
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
