package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mitchell/lambdarouter"
	"os"
)

type Item struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type ItemBatch struct {
	Batch []Item `json:"batch"`
}

type DeleteBatch struct {
	Batch []string `json:"batch"`
}

var r = lambdarouter.New("items")

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=5432 user=aherco dbname=%s password=!1qwerty",
			os.Getenv("HOST"),
			os.Getenv("DBNAME"),
		),
	)

	if err != nil {
		panic(err)
	}

	return db
}

func postItems(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var res events.APIGatewayProxyResponse
	var jsonb []byte
	var ib ItemBatch
	_ = json.Unmarshal([]byte(req.Body), &ib)

	for _, itm := range ib.Batch {
		db.Create(&itm)
	}

	jsonb, _ = json.Marshal(&ib)

	res.Body = string(jsonb)
	res.StatusCode = 201
	res.Headers = map[string]string{
		"Access-Control-Allow-Origin":      os.Getenv("ORIGIN"),
		"Access-Control-Allow-Credentials": "true",
	}

	return res, nil
}

func getItemsByChannelID(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var res events.APIGatewayProxyResponse
	var jsonb []byte
	var itms []Item
	cid := req.PathParameters["channel_id"]

	db.Where("channel_id = ?", cid).Order("created_at desc").Find(&itms)
	jsonb, _ = json.Marshal(&itms)

	res.Body = string(jsonb)
	res.StatusCode = 200
	res.Headers = map[string]string{
		"Access-Control-Allow-Origin":      os.Getenv("ORIGIN"),
		"Access-Control-Allow-Credentials": "true",
	}

	return res, nil
}

func deleteItemsByMessageID(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var res events.APIGatewayProxyResponse
	var dib DeleteBatch
	_ = json.Unmarshal([]byte(req.Body), &dib)

	for _, d := range dib.Batch {
		db.Delete(Item{}, "message_id = ?", d)
	}

	res.Body = "Delete successful"
	res.StatusCode = 204
	res.Headers = map[string]string{
		"Access-Control-Allow-Origin":      os.Getenv("ORIGIN"),
		"Access-Control-Allow-Credentials": "true",
	}

	return res, nil
}

func deleteItemByID(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var res events.APIGatewayProxyResponse
	iid := req.PathParameters["id"]
	mid := req.PathParameters["message_id"]

	db.Where("id = ? AND message_id = ?", iid, mid).Delete(Item{})

	res.Body = "Delete successful"
	res.StatusCode = 204
	res.Headers = map[string]string{
		"Access-Control-Allow-Origin":      os.Getenv("ORIGIN"),
		"Access-Control-Allow-Credentials": "true",
	}

	return res, nil
}

func init() {
	r.Post("", lambda.NewHandler(postItems))
	r.Get("channel/{channel_id}", lambda.NewHandler(getItemsByChannelID))
	r.Delete("", lambda.NewHandler(deleteItemsByMessageID))
	r.Delete("{id}/{message_id}", lambda.NewHandler(deleteItemByID))
}

func main() {
	lambda.StartHandler(r)
}
