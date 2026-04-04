package auth

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	BearerScheme            = "bearer"
)
