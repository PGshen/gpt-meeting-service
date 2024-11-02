package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("请指定要运行的函数")
		return
	}
	functionName := args[1]

	// 设置源 MongoDB 实例的连接信息
	sourceURI := "mongodb://source_mongo:27017"
	sourceDatabase := "meeting"
	sourceUsername := "root"
	sourcePasswd := "passwd"

	// 设置目标 MongoDB 实例的连接信息
	destinationURI := "mongodb://52.79.227.133:27017"
	destinationDatabase := "meeting"
	destinationUsername := "root"
	destinationPasswd := "passwd"

	switch functionName {
	case "exportData":
		exportData(sourceURI, sourceUsername, sourcePasswd, sourceDatabase)
	case "importData":
		importData(destinationURI, destinationUsername, destinationPasswd, destinationDatabase)
	default:
		fmt.Println("无效的函数名")
	}
	fmt.Println("数据处理完成！")
}

func exportData(sourceURI, username, passwd, sourceDatabase string) {
	// 导出数据到文件
	err := exportToFile(sourceURI, username, passwd, sourceDatabase, "role_template", "role_template.json")
	if err != nil {
		log.Fatal(err)
	}
	err = exportToFile(sourceURI, username, passwd, sourceDatabase, "meeting_template", "meeting_template.json")
	if err != nil {
		log.Fatal(err)
	}
}

func importData(destinationURI, username, passwd, destinationDatabase string) {
	// 从文件导入数据
	err := importFromFile(destinationURI, username, passwd, destinationDatabase, "role_template", "role_template.json")
	if err != nil {
		log.Fatal(err)
	}
	err = importFromFile(destinationURI, username, passwd, destinationDatabase, "meeting_template", "meeting_template.json")
	if err != nil {
		log.Fatal(err)
	}
}

// 导出数据到文件
func exportToFile(uri, username, passwd, database, collection, filename string) error {
	// 创建 MongoDB 客户端
	client, err := connectToMongoDB(uri, username, passwd)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// 获取数据库和集合
	db := client.Database(database)
	col := db.Collection(collection)

	// 查询所有文档
	cur, err := col.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	// 将文档序列化为 JSON，并保存到文件
	var documents []bson.M
	for cur.Next(context.Background()) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		documents = append(documents, doc)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(documents)
	if err != nil {
		return err
	}

	return nil
}

// 从文件导入数据
func importFromFile(uri, username, passwd, database, collection, filename string) error {
	// 创建 MongoDB 客户端
	client, err := connectToMongoDB(uri, username, passwd)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// 获取数据库和集合
	db := client.Database(database)
	col := db.Collection(collection)

	// 读取文件内容
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// 反序列化 JSON 数据并导入到 MongoDB
	var documents []bson.M
	err = json.Unmarshal(data, &documents)
	if err != nil {
		return err
	}

	// 清空目标集合
	_, err = col.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// 插入导入的文档
	for _, doc := range documents {
		xid, _ := primitive.ObjectIDFromHex(doc["_id"].(string))
		doc["_id"] = xid
		_, err := col.InsertOne(context.Background(), doc)
		if err != nil {
			return err
		}
	}

	return nil
}

// 创建 MongoDB 客户端
func connectToMongoDB(uri, username, passwd string) (*mongo.Client, error) {
	clientOptions := options.Client().SetAuth(
		options.Credential{
			Username: username,
			Password: passwd,
		}).ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
