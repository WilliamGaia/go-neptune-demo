package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func init() {

}

func main() {
	dbUri := "neo4j://localhost" // scheme://host(:port) (default port is 7687)
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		panic(err)
	}
	// Starting with 5.0, you can control the execution of most driver APIs
	// To keep things simple, we create here a never-cancelling context
	// Read https://pkg.go.dev/context to learn more about contexts
	ctx := context.Background()
	// Handle driver lifetime based on your application lifetime requirements.
	// driver's lifetime is usually bound by the application lifetime, which usually implies one driver instance per
	// application

	defer driver.Close(ctx)
	router := gin.Default()
	router.POST("/addComment", addComment)
	router.POST("/toggleLike", toggleLike)
	router.POST("/addPost", addPost)
	router.POST("/queryComment", func(c *gin.Context) {
		queryComment(c, driver, ctx)
	})
	router.Run(":8080")
}

// TODO Define respond object schema
func addComment(c *gin.Context) {
	request := AddCommentRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	fmt.Println("Request Body:  ", request.Content)
	c.IndentedJSON(http.StatusCreated, "Comment added.")
}

func toggleLike(c *gin.Context) {
	request := ToggleLikeRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	fmt.Println("Request Body:  ", request.CommentID)
	c.IndentedJSON(http.StatusCreated, "Toggled success.")
}

func addPost(c *gin.Context) {
	request := AddPostRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	fmt.Println("Request Body:  ", request.PostID)
	c.IndentedJSON(http.StatusCreated, "Post added.")
}

func queryComment(c *gin.Context, driver neo4j.DriverWithContext, ctx context.Context) {
	request := QueryCommentRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	fmt.Println("Request Body:  ", request.MemberID)

	// Create Neo4j session
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Execute Neo4j query
	query := `
		MATCH (c:Comment)-[:POSTED_BY]->(m:Member {id: $memberID})
		RETURN c.id as comment_id, c.memberID as member_id, c.postID as post_id, c.content as content,
		       c.nickname as nickname, c.createdAt as created_at, c.updatedAt as updated_at, c.likeQty as like_qty,
		       c.memberLiked as member_liked
	`
	params := map[string]interface{}{
		"memberID": request.MemberID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query comments"})
		return
	}

	comments := Comments{
		Data:       []Comment{},
		PageResult: PageResult{},
	}

	for result.Next(ctx) {
		record := result.Record()
		comment := Comment{
			CommentID:   record.Values[0].(string),
			MemberID:    record.Values[1].(string),
			PostID:      record.Values[2].(string),
			Content:     record.Values[3].(string),
			Nickname:    record.Values[4].(string),
			CreatedAt:   record.Values[5].(string),
			UpdatedAt:   record.Values[6].(string),
			LikeQty:     record.Values[7].(int32),
			MemberLiked: record.Values[8].(bool),
		}
		comments.Data = append(comments.Data, comment)
	}

	comments.PageResult.Total = len(comments.Data)
	comments.PageResult.Current = 1
	comments.PageResult.PageSize = 3

	fmt.Printf("Comments retrieved: %+v\n", comments)
	c.IndentedJSON(http.StatusOK, comments)
}
