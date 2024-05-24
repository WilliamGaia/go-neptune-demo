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
	fmt.Println("The uniqueID of postID is:", uniqueID)

	params := map[string]interface{}{
		"posterID":         request.MemberID,
		"postID":           uniqueID,
		"currentTimestamp": currentTimestamp,
		"content":          request.Content,
		"hashTags":         "s2otaiwan,s20,大佳河濱公園,比基尼,bikini,taipei,taiwan,",
		"price":            "{}",
		"purchasable":      true,
		"reviewStatus":     1,
		"status":           1,
		"subscribeLevel":   1,
	}
	query := `
	MATCH (m:member {memberID:$posterID})
	CREATE (p:post {
		` + "`~id`" + `: $postID,
		postID: $postID,
		status: $status,
		reviewStatus: $reviewStatus,
		subscribeLevel: $subscribeLevel,
		posterID: $posterID,
		posterAccount: m.account,
		content: $content,
		hashTags: $hashTags,
		price: $price,
		purchasable: $purchasable,
		updatedAt: datetime($currentTimestamp),
		createdAt: datetime($currentTimestamp)
	})
	MERGE (m)-[:POSTED]->(p)
	RETURN p.postID AS postID;`

	result, err := session.Run(ctx, query, params)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add post "})
		return
	}
	postID := ""
	for result.Next(ctx) {
		record := result.Record()
		p_value, found := record.Get("postID")
		if !found {
			fmt.Println("Failed to get postID from record")
			continue
		}
		postID = p_value.(string)
	}

	fmt.Printf("Created post: %+v\n", postID)
	c.IndentedJSON(http.StatusCreated, postID)
}
