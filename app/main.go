package main

import (
	"app/handlers"
	"app/utils"
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/auth"
)

const (
	hostAndPort = "pkm-osgp-index-test.cluster-cfhqexhrq16e.ap-northeast-1.neptune.amazonaws.com:8182"
)

func init() {
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAVSZUTOAI4GJCXFU4")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
}

func main() {
	ctx := context.Background()

	fetchAuthToken := utils.FetchAWSSignedToken

	tokenManager := auth.BearerTokenManager(fetchAuthToken)

	dbUri := "bolt+ssc://" + hostAndPort + "/opencypher"

	driver, err := neo4j.NewDriverWithContext(dbUri, tokenManager)
	if err != nil {
		panic(err)
	}

	defer driver.Close(ctx)
	if err := driver.VerifyConnectivity(ctx); err != nil {
		fmt.Println("failed to verify connection", err)
	}

	router := gin.Default()
	router.POST("/addComment", func(c *gin.Context) {
		handlers.AddComment(c, driver, ctx)
	})
	router.POST("/toggleLike", func(c *gin.Context) {
		handlers.ToggleLike(c, driver, ctx)
	})
	router.POST("/addPost", func(c *gin.Context) {
		handlers.AddPost(c, driver, ctx)
	})
	router.POST("/queryComment", func(c *gin.Context) {
		handlers.QueryComment(c, driver, ctx)
	})
	router.Run(":8080")
}
