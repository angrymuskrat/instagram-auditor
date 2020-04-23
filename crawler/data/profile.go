package data

type Profile struct {
	Id        string
	Username  string
	FullName  string
	Biography string

	FollowedBy int
	Follow     int

	IsBusinessAccount    bool
	IsJoinedRecently     bool
	BusinessCategoryName string
	CategoryId           string
	IsPrivate            bool
	IsVerified           bool
	ConnectedFbPage      bool

	ProfilePicUrl string
	ProfilePic    []byte

	PostsCount int
	Posts      []Post
}
