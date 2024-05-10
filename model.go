package main

type AddCommentRequest struct {
	MemberID string `json:"memberID"`
	PostID   string `json:"postID"`
	Content  string `json:"content"`
}

type ToggleLikeRequest struct {
	MemberID  string `json:"memberID"`
	CommentID string `json:"commentID"`
}

type AddPostRequest struct {
	MemberID string `json:"memberID"`
	PostID   string `json:"postID"`
}

type QueryCommentRequest struct {
	MemberID string `json:"memberID"`
	PostID   string `json:"postID"`
}

type Comment struct {
    CommentID    string  `json:"commentID"`
    MemberID     string  `json:"memberID"`
    PostID       string  `json:"postID"`
    Content      string  `json:"content"`
    Nickname     string  `json:"nickname"`
    CreatedAt    string  `json:"createdAt"`
    UpdatedAt    string  `json:"updatedAt"`
    LikeQty      int32   `json:"likeQty"`
	MemberLiked  bool    `json:"memberLiked"`
}

type PageResult struct {
	Total    int  `json:"total"`
	Current  int  `json:"current"`
	PageSize int  `json:"pageSize"`
}

type Comments struct {
	Data       []Comment   `json:"data"`
	PageResult PageResult  `json:"pageResult"`
}