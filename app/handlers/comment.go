package handlers

import (
	"app/models"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var commentStore = struct {
	sync.RWMutex
	data map[string]models.SessionData
}{data: make(map[string]models.SessionData)}

// TODO Define respond object schema
func AddComment(c *gin.Context, driver neo4j.DriverWithContext, ctx context.Context) {
	request := models.AddCommentRequest{}
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
	RETURN c, m.nickname AS nickname;`

	result, err := session.Run(ctx, query, params)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment "})
		return
	}
	comment := models.Comment{}
	for result.Next(ctx) {
		record := result.Record()
		comment_node, found := record.Get("c")
		if !found {
			fmt.Println("Failed to get comment node from record")
			continue
		}
		m_value, found := record.Get("nickname")
		if !found {
			fmt.Println("Failed to get nickname from record")
			continue
		}

		c_value, _ := comment_node.(neo4j.Node)
		nickname := m_value.(string)
		comment.CommentID = c_value.Props["commentId"].(string)
		comment.Content = c_value.Props["content"].(string)
		comment.MemberID = c_value.Props["memberID"].(string)
		comment.Nickname = nickname
		comment.PostID = c_value.Props["postID"].(string)
		comment.CreatedAt = c_value.Props["createdAt"].(time.Time).Format(time.RFC3339)
		comment.UpdatedAt = c_value.Props["updatedAt"].(time.Time).Format(time.RFC3339)
	}

	fmt.Printf("Comment retrieved: %+v\n", comment)
	c.IndentedJSON(http.StatusCreated, comment)
}

func QueryComment(c *gin.Context, driver neo4j.DriverWithContext, ctx context.Context) {
	request := bindAndSetDefaults(c)
	pageSize := 10
	nextPage := c.Query("next_page")
	prevPage := c.Query("prev_page")
	sessionID := c.Query("session_id")

	var comments []models.Comment
	var err error
	var sessionData models.SessionData
	var skip int
	var total int

	commentsChan := make(chan []models.Comment)
	countChan := make(chan int)
	errorChan := make(chan error)

	if sessionID == "" {
		go func() {
			session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
			defer session.Close(ctx)
			comments, err = fetchCommentsFromDB(session, ctx, request, 0, 100)
			if err != nil {
				errorChan <- err
				return
			}
			commentsChan <- comments
		}()
		go func() {
			session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
			defer session.Close(ctx)
			total, err := countCommentsFromDB(session, ctx, request)
			if err != nil {
				errorChan <- err
				return
			}
			countChan <- total
		}()
		select {
		case comments = <-commentsChan:
			sessionID, err = generateSessionID()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session ID"})
				return
			}

			sessionData = models.SessionData{
				Comments:         comments,
				LastFetchedIndex: 100,
			}
			select {
			case total = <-countChan:
				sessionData.Total = total
				commentStore.Lock()
				commentStore.data[sessionID] = sessionData
				commentStore.Unlock()
			case err = <-errorChan:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		case err = <-errorChan:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Existing session, fetch from memory
		commentStore.RLock()
		sessionData, found := commentStore.data[sessionID]
		commentStore.RUnlock()
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		if nextPage != "" {
			skip, _ := strconv.Atoi(nextPage)
			fmt.Println("Check if need Fetching pages")
			if skip >= len(sessionData.Comments) {
				session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
				defer session.Close(ctx)
				comments, err = fetchCommentsFromDB(session, ctx, request, sessionData.LastFetchedIndex, 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				fmt.Println("get comments:", comments)
				sessionData.Comments = append(sessionData.Comments, comments...)
				sessionData.LastFetchedIndex += 100
				commentStore.Lock()
				commentStore.data[sessionID] = sessionData
				commentStore.Unlock()
			}
		} else if prevPage != "" {
			skip, _ = strconv.Atoi(prevPage)
		}
		comments = sessionData.Comments
		total = sessionData.Total
	}

	if nextPage != "" {
		skip, _ = strconv.Atoi(nextPage)
	}

	start := skip
	end := skip + pageSize
	if start > len(comments) {
		start = len(comments)
	}
	if end > len(comments) {
		end = len(comments)
	}
	paginatedComments := comments[start:end]

	nextPageToken := ""
	if end < len(comments) {
		nextPageToken = strconv.Itoa(end)
	}
	prevPageToken := ""
	if start > 0 {
		prevPageToken = strconv.Itoa(start - pageSize)
		if start-pageSize < 0 {
			prevPageToken = "0"
		}
	}

	fmt.Println("Concat results")
	response := models.Comments{
		Data: paginatedComments,
		PageResult: models.PageResult{
			Total:         total,
			Current:       start / pageSize,
			PageSize:      pageSize,
			NextPageToken: nextPageToken,
			PrevPageToken: prevPageToken,
		},
	}

	c.IndentedJSON(http.StatusOK, gin.H{"session_id": sessionID, "comments": response})
}

func fetchCommentsFromDB(session neo4j.SessionWithContext, ctx context.Context, request models.QueryCommentRequest, skip, limit int) ([]models.Comment, error) {
	params := map[string]interface{}{}
	query := ""
	if request.MemberID == "" {
		params = map[string]interface{}{
			"postID": request.PostID,
			"skip":   skip,
			"limit":  limit,
		}
		query = `
		MATCH (p:post {postID:$postID})-[:COMMENTED_FROM]->(c:comment)
		OPTIONAL MATCH (c)<-[:LIKED]-(likedMember:member)
		RETURN c, p.posterAccount AS nickname, p.posterID AS queryerID, COLLECT(likedMember) AS likedMembers ,COUNT(likedMember) AS likeQty
		ORDER BY c.createdAt DESC
		SKIP $skip
		LIMIT $limit`
	} else {
		params = map[string]interface{}{
			"memberID": request.MemberID,
			"skip":     skip,
			"limit":    limit,
		}
		query = `
		MATCH (m:member {memberID:$memberID})-[:POSTED]->(:post)-[:COMMENTED_FROM]->(c:comment)
		OPTIONAL MATCH (c)<-[:LIKED]-(likedMember:member)
		RETURN c,m.nickname AS nickname, m.memberID AS queryerID, COLLECT(likedMember.memberID) AS likedMembers ,COUNT(likedMember) AS likeQty
		ORDER BY c.createdAt DESC
		SKIP $skip
		LIMIT $limit`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	comments := []models.Comment{}

	for result.Next(ctx) {
		record := result.Record()
		c_node, found := record.Get("c")
		if !found {
			fmt.Println("Failed to get comment node from record")
			continue
		}
		nickname_value, found := record.Get("nickname")
		if !found {
			fmt.Println("Failed to get nickname from record")
			continue
		}
		queryerID_value, found := record.Get("queryerID")
		if !found {
			fmt.Println("Failed to get queryerID from record")
			continue
		}
		likedMember_value, found := record.Get("likedMembers")
		if !found {
			fmt.Println("Failed to get likedMember from record")
			continue
		}
		likeQty_value, found := record.Get("likeQty")
		if !found {
			fmt.Println("Failed to get likeQty from record")
			continue
		}

		c_value, _ := c_node.(neo4j.Node)
		createdAt := c_value.Props["createdAt"].(time.Time).Format(time.RFC3339)
		updatedAt := c_value.Props["updatedAt"].(time.Time).Format(time.RFC3339)
		commentID := c_value.Props["commentId"].(string)
		memberID := c_value.Props["memberID"].(string)
		nickname := nickname_value.(string)
		likeQty := int32(likeQty_value.(int64))
		postID := c_value.Props["postID"].(string)
		content := c_value.Props["content"].(string)
		// TODO get from poster and member compare
		queryerID := queryerID_value.(string)
		likedMemberIDs, found := likedMember_value.([]string)
		if !found {
			likedMemberIDs = []string{}
		}
		fmt.Println("likedMember=", likedMemberIDs, "queryer or poster=", queryerID)
		memberLiked := contains(likedMemberIDs, queryerID)
		comment := models.Comment{
			CommentID:   commentID,
			Content:     content,
			MemberID:    memberID,
			PostID:      postID,
			Nickname:    nickname,
			LikeQty:     likeQty,
			MemberLiked: memberLiked,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func countCommentsFromDB(session neo4j.SessionWithContext, ctx context.Context, request models.QueryCommentRequest) (int, error) {
	params := map[string]interface{}{}
	query := ""
	if request.MemberID == "" {
		params = map[string]interface{}{"postID": request.PostID}
		query = `
		MATCH (p:post {postID:$postID})-[:COMMENTED_FROM]->(c:comment)
		RETURN COUNT(c) AS counts`
	} else {
		params = map[string]interface{}{"memberID": request.MemberID}
		query = `
		MATCH (:member {memberID:$memberID})-[:POSTED]->(:post)-[:COMMENTED_FROM]->(c:comment)
		RETURN COUNT(c) AS counts`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		total, found := result.Record().Get("counts")
		if !found {
			fmt.Println("Failed to get node from record")
			return 0, nil
		}
		return int(total.(int64)), nil
	}

	return 0, nil
}

func bindAndSetDefaults(c *gin.Context) models.QueryCommentRequest {
	var request models.QueryCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return request
	}
	return request
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func contains(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
