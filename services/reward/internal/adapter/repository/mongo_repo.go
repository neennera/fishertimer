package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

const (
	dbName         = "reward_db"
	collectionName = "fish_rewards"
)

// fishRewardDoc is the MongoDB document shape for the fish_rewards collection.
// We use bson tags so field names match the JS schema exactly.
type fishRewardDoc struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`
	DisplayName string        `bson:"display_name"`
	Species     string        `bson:"species"`
	Rarity      string        `bson:"rarity"`
	AwardedAt   time.Time     `bson:"awarded_at"`
}

func (d fishRewardDoc) toDomain() domain.FishReward {
	return domain.FishReward{
		ID:          d.ID.Hex(),
		UserID:      d.UserID,
		DisplayName: d.DisplayName,
		Species:     d.Species,
		Rarity:      d.Rarity,
		AwardedAt:   d.AwardedAt,
	}
}

// MongoRepository implements domain.Repository backed by MongoDB.
type MongoRepository struct {
	col *mongo.Collection
}

// NewMongo creates a MongoRepository connected to the given URI.
// It also ensures the required indexes exist.
func NewMongo(ctx context.Context, mongoURI string) (*MongoRepository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Verify connectivity.
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	col := client.Database(dbName).Collection(collectionName)

	// Ensure indexes (idempotent — safe to run on every startup).
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "awarded_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "awarded_at", Value: -1}}},
	}
	if _, err := col.Indexes().CreateMany(ctx, indexes); err != nil {
		return nil, err
	}

	return &MongoRepository{col: col}, nil
}

// Award inserts a new fish reward document into MongoDB.
func (r *MongoRepository) Award(ctx context.Context, reward *domain.FishReward) error {
	if reward.AwardedAt.IsZero() {
		reward.AwardedAt = time.Now().UTC()
	}
	doc := fishRewardDoc{
		UserID:      reward.UserID,
		DisplayName: reward.DisplayName,
		Species:     reward.Species,
		Rarity:      reward.Rarity,
		AwardedAt:   reward.AwardedAt,
	}
	res, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	// Write back the generated _id so caller has it.
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		reward.ID = oid.Hex()
	}
	return nil
}

// ListByUser returns rewards for a specific user, or all users when userID == "".
// Results are ordered by awarded_at descending (newest first).
func (r *MongoRepository) ListByUser(ctx context.Context, userID string) ([]domain.FishReward, error) {
	filter := bson.D{}
	if userID != "" {
		filter = bson.D{{Key: "user_id", Value: userID}}
	}
	opts := options.Find().SetSort(bson.D{{Key: "awarded_at", Value: -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var docs []fishRewardDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}

	rewards := make([]domain.FishReward, len(docs))
	for i, d := range docs {
		rewards[i] = d.toDomain()
	}
	return rewards, nil
}

// GetLastUpdate returns the awarded_at of the most recently inserted reward.
// The leaderboard service calls this endpoint to decide if its cache is stale.
func (r *MongoRepository) GetLastUpdate(ctx context.Context) (time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "awarded_at", Value: -1}})
	var doc fishRewardDoc
	err := r.col.FindOne(ctx, bson.D{}, opts).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return doc.AwardedAt, nil
}
