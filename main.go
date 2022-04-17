package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type data struct {
	ID      primitive.ObjectID `bson:"_id" json:"_id"`
	Message string             `bson:"message" json:"message"`
	Check   string             `bson:"check" json:"check"`
}

var upgrader = websocket.Upgrader{}

func main() {

	//MongoDB Connection

	username := "devyadav"
	password := "devyadav"
	cluster := "cluster0.n8tks.mongodb.net"
	authSource := "gosocket"
	authMechanism := "SCRAM-SHA-1"

	uri := "mongodb+srv://" + url.QueryEscape(username) + ":" +
		url.QueryEscape(password) + "@" + cluster +
		"/?authSource=" + authSource +
		"&authMechanism=" + authMechanism

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	//MongoDB Get Data

	collection := client.Database("gosocket").Collection("gosocket")
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	var results []data
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	mongoMessage := results[0].Message + " in " + results[0].Check

	//Sockets

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed:", err)
				break
			}

			// Print the message to the console
			// fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			for range time.Tick(time.Second * 60) {
				msg = []byte(mongoMessage)

				// Write message back to browser
				err = conn.WriteMessage(msgType, msg)
				if err != nil {
					log.Println("write failed:", err)
					break
				}
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}
