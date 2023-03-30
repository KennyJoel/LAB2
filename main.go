package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define the Movie struct
type Movie struct {
	ID    string `bson:"_id,omitempty"`
	Title string `bson:"title,omitempty"`
	Year  int    `bson:"year,omitempty"`
	Genre string `bson:"genre,omitempty"`
}

// Define a function to get the MongoDB collection
func getCollection() (*mongo.Collection, error) {
	// Set up MongoDB client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Get the MongoDB collection
	collection := client.Database("joel").Collection("fullstack")

	return collection, nil
}

// Define a handler function for the GET method with filters
func getMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	queryParams := r.URL.Query()

	// Get the filter values from the query parameters
	title := queryParams.Get("title")
	genre := queryParams.Get("genre")
	yearStr := queryParams.Get("year")
	var year int
	var err error
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}
	}

	// Get the MongoDB collection
	collection, err := getCollection()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	// Build the filter query
	filter := bson.M{}
	if title != "" {
		filter["title"] = bson.M{"$regex": title, "$options": "i"}
	}
	if genre != "" {
		filter["genre"] = bson.M{"$regex": genre, "$options": "i"}
	}
	if yearStr != "" {
		filter["year"] = year
	}

	// Query the MongoDB collection
	var movies []Movie
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var movie Movie
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Return the movies as a JSON array
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// Define the main function
func main() {
	// Create a new router
	r := mux.NewRouter()

	// Handle GET requests with filters
	r.HandleFunc("/movies", getMoviesHandler).Methods("GET")

	// Start the server
	log.Fatal(http.ListenAndServe(":8000", r))
}
