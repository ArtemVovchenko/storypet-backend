package models

type Dump struct {
	FilePath  string `json:"-" db:"dump_filepath"`
	CreatedAt string `json:"created_at" db:"created_at"`
	FileName  string `json:"file_name"`
}
