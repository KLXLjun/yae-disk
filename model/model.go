package model

import "net"

type FileStruct struct {
	FileID     uint64   `json:"file_id,string"`
	FolderID   uint64   `json:"folder_id,string"`
	FileName   string   `json:"file_name"`
	FileSize   uint64   `json:"file_size,string"`
	ChangeTime string   `json:"change_time"`
	CreateTime string   `json:"create_time"`
	OwnerID    uint64   `json:"file_owner,string"`
	Groups     []uint64 `json:"groups"`
	Permission map[uint64]PermissionStruct
}

type FolderStruct struct {
	FolderID      uint64 `json:"folder_id,string"`
	FolderName    string `json:"folder_name"`
	ChangeTime    string `json:"change_time"`
	CreateTime    string `json:"create_time"`
	OwnerFolderID uint64 `json:"owner_folder_id,string"`
	OwnerUserID   uint64 `json:"owner_user_id,string"`
	Permission    map[uint64]PermissionStruct
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

// PermissionStruct 权限
type PermissionStruct struct {
	Read    string
	Writer  string
	Visible string
}

// UserStruct 用户信息
type UserStruct struct {
	UserID   uint64   `json:"user_id,string"`
	UserName string   `json:"user_name"`
	Pass     string   `json:"pass"`
	Groups   []uint64 `json:"groups"`
}

// UserGroupStruct 用户组信息
type UserGroupStruct struct {
	GroupID     uint64   `json:"group_id"`
	GroupName   string   `json:"group_name"`
	OwnerUserID uint64   `json:"owner_user_id"`
	GroupUsers  []uint64 `json:"group_users"`
}

type UserAuthToken struct {
	UserID    uint64
	IP        net.IP
	UserToken string
}
