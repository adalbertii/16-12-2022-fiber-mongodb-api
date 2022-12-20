package main

import (
	"fmt"

	"context"
	"log"

	"encoding/json"

	"github.com/gofiber/fiber"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const dbName = "personsdb"
const collectionName = "person"
const port = 8000
const mogodbUri = "mongodb://localhost:27011"
const ststusError = 500
const ststusErrorMiss = 404

type Person struct {
	_id       string `json:”id,omitempty”`
	FirstName string `json:”firstname,omitempty”`
	LastName  string `json:”lastname,omitempty”`
	Email     string `json:”email,omitempty”`
	Age       int    `json:”age,omitempty”`
}

//-------------------------------------------------------------------------------------
func main() {

	app := fiber.New()
	app.Get("/person/:id?", getPerson)
	app.Post("/person", createPerson)
	app.Put("/person/:id", updatePerson)
	app.Delete("/person/:id", deletePerson)

	fmt.Println("Microservice listening on port: ", port)
	app.Listen(port)
}

//-------------------------------------------------------------------------------------
func getPerson(c *fiber.Ctx) {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	var filter bson.M = bson.M{}

	if c.Params("id") != "" {
		id := c.Params("id")
		objID, _ := primitive.ObjectIDFromHex(id)
		filter = bson.M{"_id": objID}
	}

	var results []bson.M
	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	cur.All(context.Background(), &results)

	if results == nil {
		c.SendStatus(ststusErrorMiss)
		return
	}

	json, _ := json.Marshal(results)
	c.Send(json)
}

//-------------------------------------------------------------------------------------
func createPerson(c *fiber.Ctx) {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	var person Person
	json.Unmarshal([]byte(c.Body()), &person)

	res, err := collection.InsertOne(context.Background(), person)
	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	response, _ := json.Marshal(res)
	c.Send(response)

}

//-------------------------------------------------------------------------------------
func updatePerson(c *fiber.Ctx) {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}
	var person Person
	json.Unmarshal([]byte(c.Body()), &person)

	update := bson.M{
		"$set": person,
	}

	objID, _ := primitive.ObjectIDFromHex(c.Params("id"))
	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)

	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	response, _ := json.Marshal(res)

	c.Send(response)

}

//-------------------------------------------------------------------------------------
func deletePerson(c *fiber.Ctx) {
	collection, err := getMongoDbCollection(dbName, collectionName)

	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	objID, _ := primitive.ObjectIDFromHex(c.Params("id"))
	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if err != nil {
		c.Status(ststusError).Send(err)
		return
	}

	jsonResponse, _ := json.Marshal(res)
	c.Send(jsonResponse)
}

//-------------------------------------------------------------------------------------
func getMongoDbConnection() (*mongo.Client, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mogodbUri))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

//-------------------------------------------------------------------------------------
func getMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {
	client, err := getMongoDbConnection()

	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}
