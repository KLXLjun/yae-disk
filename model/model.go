package model

type FileStruct struct {
	FileID     uint64 `json:"file_id,string"`
	FolderID   uint64 `json:"folder_id,string"`
	FileName   string `json:"file_name"`
	FileSize   uint64 `json:"file_size,string"`
	ChangeTime string `json:"change_time"`
	CreateTime string `json:"create_time"`
	OwnerID    uint64 `json:"file_owner,string"`
}

type FolderStruct struct {
	FolderID      uint64 `json:"folder_id,string"`
	FolderName    string `json:"folder_name"`
	ChangeTime    string `json:"change_time"`
	CreateTime    string `json:"create_time"`
	OwnerFolderID uint64 `json:"owner_folder_id,string"`
	OwnerUserID   uint64 `json:"owner_user_id,string"`
}

type ResultFolderStruct struct {
	FolderID      uint64         `json:"folder_id,string"`
	FolderName    string         `json:"folder_name"`
	ChangeTime    string         `json:"change_time"`
	CreateTime    string         `json:"create_time"`
	OwnerFolderID uint64         `json:"owner_folder_id,string"`
	OwnerUserID   uint64         `json:"owner_user_id,string"`
	File          []FileStruct   `json:"file"`
	Folder        []FolderStruct `json:"folder"`
}

type UserStruct struct {
	UserID   uint64 `json:"user_id,string"`
	UserName string `json:"user_name"`
	Pass     string `json:"pass"`
}

type UserAuthToken struct {
	UserID    uint64
	UserToken string
}
