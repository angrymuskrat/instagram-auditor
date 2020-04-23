package data

type Location struct {
	Slug string
	Id   string
	Name string
}

type Post struct {
	Shortcode       string
	CommentsCount   int
	LikesCount      int
	Timestamp       int
	IsVideo         bool

	ImageUrl        string
	Image           []byte
	Caption         string
}
