package model

import "time"

type FileRowStruct struct {
	ID          string    `json:"id"`
	IsDir       bool      `json:"is_dir"`
	OwnerFolder string    `json:"owner_folder"`
	OwnerID     string    `json:"owner_id"`
	OwnerGroup  string    `json:"owner_group"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ChangeTime  time.Time `json:"change_time"`
	CreateTime  time.Time `json:"create_time"`
}

type ResultFileRowStruct struct {
	Folder []*FileRowStruct
	File   []*FileRowStruct
}
