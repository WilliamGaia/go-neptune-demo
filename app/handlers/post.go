package handlers

import (
	"app/models"
	"context"
	"fmt"
	mrand "math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func AddPost(c *gin.Context, driver neo4j.DriverWithContext, ctx context.Context) {
	request := models.AddPostRequest{}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	currentTimestamp := time.Now().UTC().Format(time.RFC3339)
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	uniqueID := fmt.Sprintf("%d%d", time.Now().UnixNano(), r.Intn(1000))
	fmt.Println("The uniqueID of commentID is:", uniqueID)

	params := map[string]interface{}{
		"commentId":        uniqueID,
		"memberID":         request.MemberID,
		"postID":           request.PostID,
		"status":           1,
		"serialNumber":     329,
		"currentTimestamp": currentTimestamp,
		"content":          request.Content,
	}
	query := `
	MATCH (p:post {postID: $postID})
	MATCH (m:member {memberID:$memberID})
	MATCH (poster:member {memberID:p.posterID})
	CREATE (c:comment {
		` + "`~id`" + `: $commentId,
		memberID: $memberID,
		postID: $postID,
		commentId: $commentId,
		status: $status,
		serialNumber: $serialNumber,
		posterID: p.posterID,
		updatedAt: datetime($currentTimestamp),
		content: $content,
		posterAccount: p.posterAccount,
		createdAt: datetime($currentTimestamp)
	})
	MERGE (c)-[:COMMENTED_BY]->(m)
	MERGE (p)-[:COMMENTED_FROM]->(c)
	MERGE (poster)-[:RECEIVED_FROM {posterID:p.posterID}]->(c)
	RETURN c, m;`

	result, err := session.Run(ctx, query, params)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query comments "})
		return
	}
	for result.Next(ctx) {
		record := result.Record()
		comment_node, found := record.Get("c")
		if !found {
			fmt.Println("Failed to get node from record")
			continue
		}
		member_node, found := record.Get("m")
		if !found {
			fmt.Println("Failed to get node from record")
			continue
		}

		c_value, _ := comment_node.(neo4j.Node)
		m_value, _ := member_node.(neo4j.Node)
		comment.CommentID = c_value.Props["commentId"].(string)
		comment.Content = c_value.Props["content"].(string)
		comment.MemberID = c_value.Props["memberID"].(string)
		comment.Nickname = m_value.Props["nickname"].(string)
		comment.PostID = c_value.Props["postID"].(string)
		comment.CreatedAt = c_value.Props["createdAt"].(time.Time).Format(time.RFC3339)
		comment.UpdatedAt = c_value.Props["updatedAt"].(time.Time).Format(time.RFC3339)
	}

	fmt.Printf("Comment retrieved: %+v\n", comment)
	c.IndentedJSON(http.StatusCreated, comment)
}
