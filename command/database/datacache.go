package database

import (
	"YaeDisk/logx"
	"YaeDisk/model"
	"time"
)

type File struct {
	Name   string
	IsDir  bool
	Parent *File
	Self   *model.FileRowStruct
	Child  []*File
}

var RootTree = File{Name: "Root", IsDir: true, Parent: nil, Self: &model.FileRowStruct{
	ID:          "0",
	IsDir:       true,
	OwnerFolder: "",
	OwnerID:     "",
	OwnerGroup:  "",
	Name:        "Root",
	Size:        0,
	ChangeTime:  time.Date(2022, 3, 12, 20, 47, 58, 0, time.Local),
	CreateTime:  time.Date(2022, 3, 12, 20, 47, 58, 0, time.Local),
}, Child: make([]*File, 0)}

var All = make([]*model.FileRowStruct, 0)

func map2Tree() {
	start2 := time.Now().Unix()
	logx.Info("数据库记录数", len(All))
	RootTree.Child = Convert(&RootTree, "0")
	end2 := time.Now().Unix()
	logx.Info("数据库读入花费的时间:", end2-start2)
	All = nil //数据不需要了
}

func Convert(items *File, parentID string) []*File {
	tree := make([]*File, 0)

	for _, item := range All {
		if item.OwnerFolder == parentID {
			u := File{
				Name:   item.Name,
				IsDir:  item.IsDir,
				Parent: items,
				Self:   item,
				Child:  make([]*File, 0),
			}
			if item.IsDir {
				u.Child = Convert(items, item.ID)
			}
			tree = append(tree, &u)
		}
	}

	return tree
}

// PathSearch 路径搜索
func PathSearch(path []string) (bool, *File, *model.ResultFileRowStruct) {
	if len(path) == 0 {
		return true, &RootTree, files2FolderAndFile(RootTree.Child)
	}
	lt := treeSearch(path, 0, &RootTree)
	if lt != nil {
		return true, lt, files2FolderAndFile(lt.Child)
	}
	return false, nil, nil
}

func treeSearch(path []string, pathDepth int, lastFolder *File) *File {
	if len(path) == pathDepth {
		return lastFolder
	}
	for _, v := range lastFolder.Child {
		if v.Name == path[pathDepth] && v.IsDir {
			return treeSearch(path, pathDepth+1, v)
		}
	}
	return nil
}

func files2FolderAndFile(input []*File) *model.ResultFileRowStruct {
	var files = make([]*model.FileRowStruct, 0)
	var folders = make([]*model.FileRowStruct, 0)
	for _, v := range input {
		if v.IsDir {
			folders = append(folders, v.Self)
		} else {
			files = append(files, v.Self)
		}
	}
	return &model.ResultFileRowStruct{File: files, Folder: folders}
}
