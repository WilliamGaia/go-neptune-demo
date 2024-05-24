package handlers

import (
	"app/models"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ToggleLike(c *gin.Context, driver neo4j.DriverWithContext, ctx context.Context) {
	request := models.ToggleLikeRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	params := map[string]interface{}{
		"memberID":  request.MemberID,
		"commentID": request.CommentID,
	}
	checkQuery := `
	MATCH (m:member {memberID: $memberID}), (c:comment {commentId: $commentID})
	OPTIONAL MATCH (m)-[l:LIKED]->(c)
	RETURN COUNT(l) AS count
	`

	result, err := session.Run(ctx, checkQuery, params)
	if err != nil {
		fmt.Println(err.Error())
		relation := "Failed to check like from %s to %s"
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf(relation, request.MemberID, request.CommentID)})
		return
	}
	var count int64
	for result.Next(ctx) {
		result_value, found := result.Record().Get("count")
		if !found {
			fmt.Println("Failed to check liked exist from record")
			continue
		}
		count = result_value.(int64)
	}
	var actionQuery string
	if count > 0 {
		actionQuery = `
		MATCH (m:member {memberID: $memberID})-[l:LIKED]->(c:comment {commentId: $commentID})
		DELETE l
		RETURN false AS result
		`
	} else {
		actionQuery = `
		MATCH (m:member {memberID: $memberID}), (c:comment {commentId: $commentID})
		MERGE (m)-[:LIKED]->(c)
		RETURN true AS result
		`
	}

	result, err = session.Run(ctx, actionQuery, params)
	if err != nil {
		fmt.Println("Error running action query:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
		return
	}

	var toggleLike bool
	for result.Next(ctx) {
		record := result.Record()
		l_value, found := record.Get("result")
		if !found {
			fmt.Println("Failed to get postID from record")
			continue
		}
		toggleLike = l_value.(bool)
	}
	relation := "toggle like from %s to %s as %s"
	c.IndentedJSON(http.StatusCreated, fmt.Sprintf(relation, request.MemberID, request.CommentID, toggleLike))
}
