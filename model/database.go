package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var detailColl *mongo.Collection
var basicColl *mongo.Collection

func InitDatabase() (err error) {
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	client, err = mongo.NewClient(clientOptions)
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	detailColl = client.Database("ayachan").Collection("detail")
	basicColl = client.Database("ayachan").Collection("basic")
	return nil
}

func QueryChartBasic(chartID int, diff int) (chart Chart, empty bool) {
	filter := bson.M{"id": chartID, "diff": diff}
	cur, _ := basicColl.Find(context.TODO(), filter)
	var res []Chart
	cur.All(context.TODO(), &res)
	if len(res) == 0 {
		return chart, true
	}
	return res[0], false
}

func QueryChartDetail(chartID int, diff int) (detail Detail, err error) {
	filter := make(bson.M)
	filter["id"] = chartID
	if chartID <= 500 || chartID == 1000 || chartID == 1001 {
		filter["diff"] = diff
	}
	err = detailColl.FindOne(context.TODO(), filter).Decode(&detail)
	if err != nil {
		return detail, err
	}
	return detail, err
}

func UpdateBasic(chartID int, diff int, chart Chart) (err error) {
	filter := bson.M{"id": chartID, "diff": diff}
	count, err := basicColl.CountDocuments(context.TODO(), filter)
	if count == 0 {
		_, err := basicColl.InsertOne(context.TODO(), chart)
		return err
	}
	_, err = basicColl.UpdateOne(context.TODO(), filter, bson.M{"$set": chart})
	return err
}

func UpdateDetail(chartID int, diff int, detail Detail) (err error) {
	filter := bson.M{"id": chartID, "diff": diff}
	count, err := detailColl.CountDocuments(context.TODO(), filter)
	if count == 0 {
		_, err := detailColl.InsertOne(context.TODO(), detail)
		return err
	}
	_, err = detailColl.UpdateOne(context.TODO(), filter, bson.M{"$set": detail})
	return err
}
