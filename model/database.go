package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"time"
)

var client *mongo.Client
var detailColl *mongo.Collection
var basicColl *mongo.Collection
var authorColl *mongo.Collection
var SonolusListColl *mongo.Collection

func InitDatabase() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		MongoURL,
	))
	if err != nil {
		log.Fatal(err)
	}
	detailColl = client.Database("ayachan").Collection("detail")
	basicColl = client.Database("ayachan").Collection("basic")
	authorColl = client.Database("ayachan").Collection("author")
	SonolusListColl = client.Database("ayachan").Collection("sonolus")
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
	cur, err := basicColl.Find(context.TODO(), filter)
	if err != nil {
		return chart, false
	}
	var res []Chart
	err = cur.All(context.TODO(), &res)
	if err != nil {
		return chart, false
	}
	if len(res) == 0 {
		filter = bson.M{"id": chartID, "diff": diff, "authorID": 0}
		cur, err := basicColl.Find(context.TODO(), filter)
		if err != nil {
			return chart, false
		}
		err = cur.All(context.TODO(), &res)
		if err != nil {
			return chart, false
		}
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
		filter = bson.M{"diff": diff, key: bson.M{"$gte": value}, "id": bson.M{"$lt": 500}}
	}
	rk, _ := detailColl.CountDocuments(context.TODO(), filter)
	return int(rk)
}

func QueryDiffDistribution(diff int) (distribution map[int]int, base int, ceil int) {
	var filter bson.M
	if diff >= 3 {
		filter = bson.M{"diff": bson.M{"$gte": 3}, "id": bson.M{"$lt": 500}}
	} else {
		filter = bson.M{"diff": diff, "id": bson.M{"$lt": 500}}
	}
	var res []Chart
	findOption := options.Find()
	findOption.SetProjection(bson.M{"level": 1})
	cur, _ := basicColl.Find(context.TODO(), filter, findOption)
	_ = cur.All(context.TODO(), &res)
	distribution = make(map[int]int)
	min := 50
	max := 2
	for _, item := range res {
		distribution[item.Level]++
		if item.Level > max {
			max = item.Level
		}
		if item.Level < min {
			min = item.Level
		}

	}
	return distribution, min, max
}

func GetSonolusList() (list []SonolusList, err error) {
	DeleteSonolusList()
	cur, _ := SonolusListColl.Find(context.TODO(), bson.M{})
	err = cur.All(context.TODO(), &list)
	return list, err
}

func CheckSonolusList(id int) (flag bool) {
	count, _ := SonolusListColl.CountDocuments(context.TODO(), bson.M{"id": id})
	return count != 0
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
	k = 1.2/(ceil-base) + 0.1
	b = float32(ceilLevel) + 0.2 - k*ceil
	return k, b
}

func UpdateBasic(chartID int, diff int, chart Chart) (err error) {
	tmp := chart.Notes
	UploadChart(chart)
	chart.Notes = nil
	filter := bson.M{"id": chartID, "diff": diff}
	count, err := basicColl.CountDocuments(context.TODO(), filter)
	if count == 0 {
		_, err := basicColl.InsertOne(context.TODO(), chart)
		return err
	} else {
		_, err = basicColl.UpdateOne(context.TODO(), filter, bson.M{"$set": chart})
	}
	chart.Notes = tmp
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

func UpdateSonolusList(listItem SonolusList) {
	DeleteSonolusList()
	listItem.Upload = time.Now().Unix()
	_, _ = SonolusListColl.InsertOne(context.TODO(), listItem)
}

func DeleteSonolusList() {
	filter := bson.M{"upload": bson.M{"$lte": time.Now().Add(-time.Hour * 6).Unix()}}
	_, _ = SonolusListColl.DeleteMany(context.TODO(), filter)
}
