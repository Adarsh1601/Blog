package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// My local postgres credentials
const (
	DB_USER     = "postgres"
	DB_PASSWORD = "psql"
	DB_NAME     = "blog_post"
)

func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

// Structure for parameters
type Blog struct {
	PostId    int    `json:"post_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	AuthorId  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type JsonResponse struct {
	Type    string `json:"type"`
	Data    []Blog `json:"data,omitempty"`
	Message string `json:"message"`
}

func main() {
	router := mux.NewRouter()

	//To retrieve all blogs
	router.HandleFunc("/get-blogs/", GetBlog).Methods("GET")

	//To retrieve specific Blog By ID
	router.HandleFunc("/get-blogsbyid/{PostId}", GetBlogByID).Methods("GET")

	//To Create a new Blog
	router.HandleFunc("/add-blogs/", CreateBlog).Methods("POST")

	//To delete specific Blog post by id
	router.HandleFunc("/delete-blogsbyid/{PostId}", DeleteBlog).Methods("DELETE")

	//To update a specific blog by it's id
	router.HandleFunc("/update-blogsbyid/{PostId}", UpdateBlog).Methods("POST")

	//Starting the service at port 8000, we can modify it as per our need
	fmt.Println("Server Started at 8000")
	log.Fatal(http.ListenAndServe("localhost:8000", router))
}

func checkErr(err error) {
	if err != nil {
		//We are not using panic for this as it may cause code to breakdown in case of any unwanted case.
		fmt.Println("Eoor", err)
	}
}

func GetBlog(w http.ResponseWriter, r *http.Request) {
	db := setupDB()
	rows, err := db.Query("SELECT * FROM blogs")
	checkErr(err)
	var blogs []Blog

	for rows.Next() {
		var post_id int
		var title string
		var content string
		var author_id string
		var created_at string
		var updated_at string

		err = rows.Scan(&post_id, &title, &content, &author_id, &created_at, &updated_at)

		checkErr(err)

		blogs = append(blogs, Blog{PostId: post_id, Title: title, Content: content, AuthorId: author_id, CreatedAt: created_at, UpdatedAt: updated_at})
	}

	var response = JsonResponse{Type: "Success", Data: blogs, Message: "Successfully Fetched all Blogs"}

	json.NewEncoder(w).Encode(response)
}

func GetBlogByID(w http.ResponseWriter, r *http.Request) {
	db := setupDB()
	params := mux.Vars(r)
	PostId := params["PostId"]

	var response = JsonResponse{}

	if PostId == "" {
		response = JsonResponse{Type: "error", Message: "You are missing PostId parameter."}
		json.NewEncoder(w).Encode(response)

	} else {
		rows, err := db.Query("SELECT * FROM blogs Where post_id=$1", PostId)
		checkErr(err)

		var blogs []Blog

		for rows.Next() {
			var post_id int
			var title string
			var content string
			var author_id string
			var created_at string
			var updated_at string

			err = rows.Scan(&post_id, &title, &content, &author_id, &created_at, &updated_at)

			checkErr(err)

			blogs = append(blogs, Blog{PostId: post_id, Title: title, Content: content, AuthorId: author_id, CreatedAt: created_at, UpdatedAt: updated_at})
		}
		response = JsonResponse{Type: "Success", Data: blogs, Message: "Successfully Fetched blog For given post id"}

		json.NewEncoder(w).Encode(response)

	}
}

func CreateBlog(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var response = JsonResponse{}

	if payload.Title == "" || payload.Content == "" || payload.AuthorID == "" {
		response = JsonResponse{Type: "error", Message: "You are missing Title, Content or Author ID parameter."}
	} else {
		db := setupDB()
		var lastInsertID int
		err := db.QueryRow("INSERT INTO blogs(title, content,author_id) VALUES($1, $2, $3) returning post_id;", payload.Title, payload.Content, payload.AuthorID).Scan(&lastInsertID)
		checkErr(err)
		response = JsonResponse{Type: "success", Message: "The blog has been inserted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteBlog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	PostId := params["PostId"]

	var response = JsonResponse{}

	if PostId == "" {
		response = JsonResponse{Type: "error", Message: "You are missing PostId parameter."}
	} else {
		db := setupDB()
		_, err := db.Exec("DELETE FROM blogs where post_id = $1", PostId)
		checkErr(err)
		response = JsonResponse{Type: "success", Message: "The Blog has been deleted successfully!"}
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateBlog(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	PostId := params["PostId"]
	var response = JsonResponse{}

	if PostId == "" {
		response = JsonResponse{Type: "Failure", Message: "Please enter the post id!"}
	} else {
		var payload struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			AuthorID string `json:"author_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if payload.Title == "" || payload.Content == "" || payload.AuthorID == "" {
			response = JsonResponse{Type: "error", Message: "You are missing Title, Content or Author ID parameter."}
		} else {
			db := setupDB()
			var lastInsertID int
			err := db.QueryRow("UPDATE blogs SET title = $1,content =$2,author_id= $3 WHERE post_id = $4 ", payload.Title, payload.Content, payload.AuthorID, PostId).Scan(&lastInsertID)
			checkErr(err)
			response = JsonResponse{Type: "success", Message: "The blog has been updated successfully!"}
		}
	}
	json.NewEncoder(w).Encode(response)
}
