package forms

// CommentForm is the body for creating/updating a comment.
type CommentForm struct {
	Body string `json:"body" binding:"required,max=2000" example:"Grab the 2L one"`
}
