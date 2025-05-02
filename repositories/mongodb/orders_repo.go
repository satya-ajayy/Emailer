package mongodb

import (
	// Go Internal Packages
	"context"

	// Local Packages
	models "emailer/models"

	// External Packages
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrdersRepository struct {
	client     *mongo.Client
	collection string
}

func NewOrdersRepository(client *mongo.Client) *OrdersRepository {
	return &OrdersRepository{client: client, collection: "orders"}
}

// GetOrder returns the order with the given ID from the database
func (r *OrdersRepository) GetOrder(ctx context.Context, orderID string) (models.Order, error) {
	collection := r.client.Database("mybase").Collection(r.collection)
	filter := bson.M{"_id": orderID}
	order := models.Order{}
	err := collection.FindOne(ctx, filter).Decode(&order)
	return order, err
}
