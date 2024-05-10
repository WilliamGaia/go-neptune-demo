package main

import (
	"net/http"
	"fmt"
	"time"
	"github.com/gin-gonic/gin"
)

func init() {
	
}

func main() {
	router := gin.Default()
	router.POST("/addComment", addComment)
	router.POST("/toggleLike", toggleLike)
	router.POST("/addPost", addPost)
	router.POST("/queryComment", queryComment)
	router.Run(":8080")
}
// TODO Define respond object schema 
func addComment(c *gin.Context) {
	request := AddCommentRequest{}
	if err:= c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
	fmt.Println("Request Body:  ",request.Content)
	c.IndentedJSON(http.StatusCreated, "Comment added.")
}

func toggleLike(c *gin.Context) {
	request := ToggleLikeRequest{}
	if err:= c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
	fmt.Println("Request Body:  ",request.CommentID)
	c.IndentedJSON(http.StatusCreated, "Toggled success.")
}

func addPost(c *gin.Context) {
	request := AddPostRequest{}
	if err:= c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
	fmt.Println("Request Body:  ",request.PostID)
	c.IndentedJSON(http.StatusCreated, "Post added.")
}

func queryComment(c *gin.Context) {
	request := QueryCommentRequest{}
	if err:= c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
	fmt.Println("Request Body:  ",request.MemberID)
	comments := Comments{
		Data: []Comment{},
		PageResult: PageResult{},
	}
	newComment := Comment{
		CommentID:   "c123",
		MemberID:    "m456",
		PostID:      "p789",
		Content:     "This is a sample comment.",
		Nickname:    "SampleUser",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
		LikeQty:     10,
		MemberLiked: true,
	}
	comments.Data = append(comments.Data, newComment)
	comments.PageResult.Total = len(comments.Data)
	comments.PageResult.Current = 1
	comments.PageResult.PageSize = 3

	fmt.Printf("New comment added: %+v\n", comments)
	c.IndentedJSON(http.StatusOK, comments)
}
