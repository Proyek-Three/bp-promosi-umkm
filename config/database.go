package config

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ulbimongoconn *mongo.Database

func ConnectDB() {
	// URI MongoDB (ubah jika perlu)
	clientOptions := options.Client().ApplyURI("mongodb+srv://dzulkiflifaiz11:SAKTIMlucu12345@webservice.dqol9t4.mongodb.net/")

	// Koneksi ke MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Cek koneksi
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Cannot connect to MongoDB:", err)
	}

	// Pilih database
	Ulbimongoconn = client.Database("proyek3")

	fmt.Println("✅ Connected to MongoDB!")
}
