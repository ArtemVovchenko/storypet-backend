package models

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/filesutil"
	"time"
)

type Dump struct {
	FilePath  string    `json:"-" db:"dump_filepath"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	FileName  string    `json:"file_name"`
}

func (d *Dump) AfterCreate() {
	d.FileName = filesutil.ExtractFileName(d.FilePath)
}
