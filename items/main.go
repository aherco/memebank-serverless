package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aherco/lambdarouter"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"
)

type Item struct {
	gorm.Model
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
}

type ItemBatch struct {
	Batch []Item `json:"batch"`
}

type DeleteBatch struct {
	Batch []string `json:"batch"`
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=5432 user=aherco dbname=postgres password=!1qwerty",
			os.Getenv("HOST"),
		),
	)

	if err != nil {
		panic(err)
	}

	return db
}

func postItems(ctx *lambdarouter.APIGContext) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var jsonb []byte
	var ib ItemBatch
	_ = json.Unmarshal(ctx.Body, &ib)

	for _, itm := range ib.Batch {
		db.Create(&itm)
	}

	jsonb, _ = json.Marshal(&ib)

	ctx.Body = jsonb
	ctx.Status = 201
}

func getItemsByChannelID(ctx *lambdarouter.APIGContext) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var jsonb []byte
	var itms []Item
	cid := ctx.Path["channel_id"]

	db.Where("channel_id = ?", cid).Order("created_at desc").Find(&itms)
	jsonb, _ = json.Marshal(&itms)

	ctx.Body = jsonb
	ctx.Status = 200
}

func deleteItemsByMessageID(ctx *lambdarouter.APIGContext) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})

	var dib DeleteBatch
	_ = json.Unmarshal([]byte(ctx.Body), &dib)

	for _, d := range dib.Batch {
		db.Delete(Item{}, "message_id = ?", d)
	}

	ctx.Body = []byte("Delete successful")
	ctx.Status = 204
}

func deleteItemByID(ctx *lambdarouter.APIGContext) {
	db := ConnectDB()
	db.AutoMigrate(&Item{})
	iid := ctx.Path["id"]

	db.Delete(Item{}, "id = ?", iid)

	ctx.Body = []byte("Delete successful")
	ctx.Status = 204
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var (
		cfg lambdarouter.APIGRouterConfig
		r   *lambdarouter.APIGRouter
	)

	cfg.Context = context.Background()
	cfg.Request = &request
	cfg.Prefix = "/items"
	cfg.Headers = map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
	}

	r = lambdarouter.NewAPIGRouter(&cfg)
	r.Post("/", postItems)
	r.Get("/channel/{channel_id}", getItemsByChannelID)
	r.Delete("/", deleteItemsByMessageID)
	r.Delete("/{id}", deleteItemByID)

	return r.Respond(), nil
}

func main() {
	lambda.Start(Handler)
}
