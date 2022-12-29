package utils

type Thumbnail struct {
	Id       uint32 `json:"-" gorm:"primary_key"`
	ContId   string `json:"cont_id" validate:"required"`
	ContType string `json:"cont_type" validate:"required"`
	Field    string `json:"field" validate:"required"`
	Locale   string `json:"locale" validate:"required"`
	Size     string `json:"size"`
	Url      string `json:"url" validate:"required"`
}

const DEFAULT_THUMB_SIZE = "455x255_JPEG_90"
const THEUPDATE_THUMB_SIZE = "200x200_JPEG_80"
const MAGAZINEPOST_THUMB_SIZE = "690x320_JPEG_80"
const BOOK_SMALL_THUMB_SIZE = "230x280_JPEG_80"
const PODCAST_IMG_THUMB_SIZE = "455x455_JPEG_90"
const PODCAST_ALTER_IMG_THUMB_SIZE = "150x150_PNG_BestCompression"
const SKILL_THUMB_SIZE = "455x455_JPEG_80"
const BOOKSERIE_THUMB_SIZE = "455x455_JPEG_80"
