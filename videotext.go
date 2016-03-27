package main

import (
	"io"
	"fmt"
	"net/http"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"encoding/json"
	"time"
)

var collection *mgo.Collection

type Comment struct {
    Id     bson.ObjectId   `json:"id" bson:"_id,omitempty"`
    Type   string      `json:"type"` // code, footnote
    Start  float32 		`json:"start"`
    End    float32		`json:"end"`
    Text   string   `json:"text"`
    Author string   `json:"author"`
    Time   string	`json:"time"`
    MediaId string	`json:"mediaid" bson:"mediaid"`
}


func handleGetComments (mediaid string, w http.ResponseWriter, r *http.Request) {
		var results []Comment
		err := collection.Find(bson.M{"mediaid": mediaid}).All(&results)
		if err != nil {
			http.Error(w, "bad request", http.StatusInternalServerError)
			return
		}
		//fmt.Println(results[0])
		m, err := json.Marshal(results)
		if err != nil {
			http.Error(w, "bad request", http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(m))
}

func handlePostComment (mediaid string, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var c Comment
	err := decoder.Decode(&c)
	c.Time = time.Now().Format(time.UnixDate)
	c.MediaId = mediaid
	// Insert Datas
	err = collection.Insert(&c)
	if err != nil {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
}

func handleDeleteComment (mediaid string, w http.ResponseWriter, r *http.Request) {
	commentid := r.URL.Query().Get("commentid")
	if len(commentid) == 0 {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	collection.Remove(bson.M{"_id": bson.ObjectIdHex(commentid)})
}

func commentapi(w http.ResponseWriter, r *http.Request) {
	mediaid := r.URL.Query().Get("mediaid")
	if len(mediaid) == 0 {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case "GET":
		handleGetComments(mediaid, w, r)
	case "POST":
		handlePostComment(mediaid, w, r)
		break
	case "DELETE":
		handleDeleteComment(mediaid, w, r)
		break
	default:
		// Give an error message.
		http.Error(w, "bad request", http.StatusInternalServerError)
	}
}

func main() {
	session, err :=  mgo.Dial("127.0.0.1") 
	if err != nil {
		panic(err)
	}
	defer session.Close()
	collection = session.DB("videotext").C("comments")

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/comments/", commentapi)
	fmt.Println("listen on 3001")
	http.ListenAndServe(":3001", nil)
}
