package models

import "time"

type Dump struct {
	FilePath  string    `json:"-" db:"dump_filepath"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	FileName  string    `json:"file_name"`
}
