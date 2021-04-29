package entity

import "time"

type Comment struct {
	ID       int
	CityID   int
	PosterID int
	Text     string
	Created  time.Time
	Modified time.Time
}

type GetCityCommentsInput struct {
	City             City
	NumberOfComments int
}

type GetCityCommentsOutput struct {
	City     City
	Comments []CommentWithPosterName
}

type CommentWithPosterName struct {
	Comment    Comment
	PosterName string
}
