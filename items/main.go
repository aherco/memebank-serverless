package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Item struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

type ItemBatch struct {
	Batch []Item `json:"batch"`
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(
		"postgres",
		"host=memebankdb.cmjoxuixlefr.us-east-1.rds.amazonaws.com port=5432 user=aherco dbname=membankdb password=!1qwerty",
	)

	if err != nil {
		panic(err)
	}

	return db
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var res events.APIGatewayProxyResponse
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	res.Headers = map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
	}

	switch request.HTTPMethod {

	case "POST":
		var jsonb []byte
		var ib ItemBatch
		_ = json.Unmarshal([]byte(request.Body), &ib)

		for _, itm := range ib.Batch {
			db.Create(&itm)
		}

		jsonb, _ = json.Marshal(&ib)

		res.Body = string(jsonb)
		res.StatusCode = 201
		return res, nil

	case "GET":
		var jsonb []byte
		var itms []Item
		gid := request.PathParameters["guild_id"]

		db.Where("guild_id = ?", gid).Find(&itms)
		jsonb, _ = json.Marshal(&itms)

		res.Body = string(jsonb)
		res.StatusCode = 200
		return res, nil
	}

	res.Body = "FAILED"
	res.StatusCode = 420
	return res, nil
}

func main() {
	lambda.Start(Handler)
}
