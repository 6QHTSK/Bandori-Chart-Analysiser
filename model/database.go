package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var detailColl *mongo.Collection
var basicColl *mongo.Collection
var authorColl *mongo.Collection

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
	authorColl = client.Database("ayachan").Collection("author")
	filter := bson.M{"authorID": 0}
	cur, _ := authorColl.Find(context.TODO(), filter)
	var res []Author
	cur.All(context.TODO(), &res)
	if len(res) == 0 {
		var authorInfo Author
		authorInfo.AuthorID = 0
		authorInfo.UserName = "Bushiroad"
		_, err = authorColl.InsertOne(context.TODO(), authorInfo)
	}
	return nil
}

func QueryChartBasic(chartID int, diff int) (chart Chart, empty bool) {
	filter := bson.M{"id": chartID, "authorID": bson.M{"$ne": 0}}
	cur, _ := basicColl.Find(context.TODO(), filter)
	var res []Chart
	cur.All(context.TODO(), &res)
	if len(res) == 0 {
		filter = bson.M{"id": chartID, "diff": diff, "authorID": 0}
		cur, _ := basicColl.Find(context.TODO(), filter)
		cur.All(context.TODO(), &res)
		if len(res) == 0 {
			return chart, true
		}
	}
	return res[0], false
}

func QueryChartDetail(chartID int, diff int, authorID int) (detail Detail, err error) {
	filter := make(bson.M)
	filter["id"] = chartID
	if authorID == 0 {
		filter["diff"] = diff
	}
	err = detailColl.FindOne(context.TODO(), filter).Decode(&detail)
	if err != nil {
		return detail, err
	}
	return detail, err
}

func QueryAuthor(UserName string) (AuthorInfo Author, empty bool) {
	filter := bson.M{"username": UserName}
	cur, _ := authorColl.Find(context.TODO(), filter)
	var res []Author
	cur.All(context.TODO(), &res)
	if len(res) == 0 {
		return AuthorInfo, true
	}
	return res[0], false
}

func QueryAuthorID(AuthorID int) (AuthorInfo Author, empty bool) {
	filter := bson.M{"authorID": AuthorID}
	cur, _ := authorColl.Find(context.TODO(), filter)
	var res []Author
	cur.All(context.TODO(), &res)
	if len(res) == 0 {
		return AuthorInfo, true
	}
	return res[0], false
}

func QueryRank(key string, value float32, diff int) (rank int) {
	var filter bson.M
	if diff >= 3 {
		filter = bson.M{"diff": bson.M{"$gte": 3}, key: bson.M{"$gte": value}, "id": bson.M{"$lt": 500}}
	} else {
		filter = bson.M{"diff": diff, "authorID": 0, key: bson.M{"$gte": value}, "id": bson.M{"$lt": 500}}
	}
	rk, _ := detailColl.CountDocuments(context.TODO(), filter)
	return int(rk)
}

func QueryDiffDistribution(diff int) (distribution map[int]int, base int, ceil int) {
	var baseLevel = [5]int{5, 10, 15, 20, 20}
	currentLevel := baseLevel[diff]
	distribution = make(map[int]int)
	for {
		var filter bson.M
		if diff >= 3 {
			filter = bson.M{"level": bson.M{"$gte": currentLevel}, "diff": bson.M{"$gte": 3}, "id": bson.M{"$lt": 500}}
		} else {
			filter = bson.M{"level": bson.M{"$gte": currentLevel}, "diff": diff, "id": bson.M{"$lt": 500}}
		}
		count, _ := basicColl.CountDocuments(context.TODO(), filter)
		if count == 0 {
			break
		}
		distribution[currentLevel] = int(count)
		currentLevel++
	}
	return distribution, baseLevel[diff], currentLevel - 1
}

func CalcDiffLiner(key string, diff int, baseRank int, ceilLevel int) (k float32, b float32) {
	var filter bson.M
	if diff >= 3 {
		filter = bson.M{"diff": bson.M{"$gte": 3}, "id": bson.M{"$lt": 500}}
	} else {
		filter = bson.M{"diff": diff, "id": bson.M{"$lt": 500}}
	}
	filterOption := options.Find()
	filterOption.SetSort(bson.M{key: -1})
	filterOption.SetProjection(bson.M{key: 1})
	filterOption.SetLimit(int64(baseRank))
	cur, _ := detailColl.Find(context.TODO(), filter, filterOption)
	var res []Detail
	_ = cur.All(context.TODO(), &res)
	tempKey := []byte(key)
	tempKey[0] -= 32
	key = string(tempKey)
	ceilReflect := reflect.ValueOf(res[0])
	ceilReflectdata := ceilReflect.FieldByName(key)
	ceil := float32(ceilReflectdata.Float())
	base := float32(reflect.ValueOf(res[len(res)-1]).FieldByName(key).Float())
	k = 1.2 / (ceil - base)
	b = float32(ceilLevel) + 0.2 - k*ceil
	return k, b
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

func UpdateAuthor(username string, nickname string) (err error) {
	filter := bson.M{"username": username}
	count, err := authorColl.CountDocuments(context.TODO(), filter)
	if count == 0 {
		total, err := authorColl.CountDocuments(context.TODO(), bson.M{})
		if err != nil {
			return err
		}
		var AuthorInfo Author
		AuthorInfo.AuthorID = int(total)
		AuthorInfo.UserName = username
		AuthorInfo.NickName = nickname
		_, err = authorColl.InsertOne(context.TODO(), AuthorInfo)
		return err
	}
	_, err = authorColl.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"nickname": nickname}})
	return err
}
