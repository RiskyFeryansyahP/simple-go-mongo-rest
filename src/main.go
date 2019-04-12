package main

/**
* Created By Risky Feryansyah Pribadi
* Description : Tutorial how to make simple rest api with MongoDB
*/

import (
	"context"
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type E struct {
	Key string
	Value interface{}
}

type Person struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname string `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var database *mongo.Database

func createPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(request.Body).Decode(&person)
	collection := database.Collection("people")
	result, err := collection.InsertOne(context.Background(), person)
	if err != nil {
		log.Fatal("Error Inserted Data to Collection :", err)
	}

	json.NewEncoder(response).Encode(result)

}

func getPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var people []Person
	collection := database.Collection("people")
	result, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message" : "` + err.Error() + `" }`))
		return
	}
	defer result.Close(context.Background())
	for result.Next(context.Background()) {
		var person Person
		result.Decode(&person)
		people = append(people, person)
	}
	if err := result.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"Message" : "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(people)
}

func getPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	var person Person
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := database.Collection("people")
	err := collection.FindOne(context.Background(), Person{ID : id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(&person)
}

func updatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	var person Person
	id, _ := primitive.ObjectIDFromHex(params["id"])
	json.NewDecoder(request.Body).Decode(&person)
	collection := database.Collection("people")
	result, err := collection.UpdateOne(context.Background(), Person{ID : id}, bson.D{
		primitive.E{
			Key : "$set",
			Value : person,
		},
	})

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(response).Encode(result)
	
}

func deletePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := database.Collection("people")
	result, err := collection.DeleteOne(context.Background(), Person{ID : id})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message" : "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func main()  {
	fmt.Println("Application Running on : localhost:8000")

	// config database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:admin123@ds229722.mlab.com:29722/go-mongo"))
	if err != nil {
		log.Fatal("Can't Connect to Database :", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success Connected to Database")
	database = client.Database("go-mongo")
	
	// setting router to get endpoint with mux
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/person", createPersonEndpoint).Methods("POST")
	router.HandleFunc("/people", getPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", getPersonEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", updatePersonEndpoint).Methods("PUT")
	router.HandleFunc("/person/{id}", deletePersonEndpoint).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}