package database

import (
	"YaeDisk/logx"
	"YaeDisk/model"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type File struct {
	Parent *File
	Self   *model.FileRowStruct
	Child  cmap.ConcurrentMap[string, *File]
}

var RootTree = File{Parent: nil, Self: &model.FileRowStruct{
	ID:          "0",
	IsDir:       true,
	OwnerFolder: "",
	OwnerID:     "",
	OwnerGroup:  "",
	Name:        "Root",
	Size:        0,
	ChangeTime:  time.Date(2022, 3, 12, 20, 47, 58, 0, time.Local),
	CreateTime:  time.Date(2022, 3, 12, 20, 47, 58, 0, time.Local),
}, Child: cmap.New[*File]()}

var All = make([]*model.FileRowStruct, 0)

func map2Tree() {
	start2 := time.Now().Unix()
	logx.Info("数据库记录数", len(All))
	RootTree.Child = Convert(&RootTree, "0")
	end2 := time.Now().Unix()
	logx.Info("数据库读入花费的时间:", end2-start2)
	All = nil //数据不需要了
}

func Convert(items *File, parentID string) cmap.ConcurrentMap[string, *File] {
	tree := cmap.New[*File]()

	for _, item := range All {
		if item.OwnerFolder == parentID {
			u := File{
				Parent: items,
				Self:   item,
				Child:  cmap.New[*File](),
			}
			if item.IsDir {
				u.Child = Convert(items, item.ID)
			}
			tree.Set(item.Name, &u)
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
		if lt.Self.IsDir {
			return true, lt, files2FolderAndFile(lt.Child)
		}
		return true, lt, nil
	}
	return false, nil, nil
}

func treeSearch(path []string, pathDepth int, lastFolder *File) *File {
	if len(path) == pathDepth {
		return lastFolder
	}
	t, have := lastFolder.Child.Get(path[pathDepth])
	if !have {
		return nil
	}
	if t.Self.IsDir {
		return treeSearch(path, pathDepth+1, t)
	} else {
		return t
	}
}

func files2FolderAndFile(input cmap.ConcurrentMap[string, *File]) *model.ResultFileRowStruct {
	var files = make([]*model.FileRowStruct, 0)
	var folders = make([]*model.FileRowStruct, 0)
	for _, v := range input.Items() {
		if v.Self.IsDir {
			folders = append(folders, v.Self)
		} else {
			files = append(files, v.Self)
		}
	}
	return &model.ResultFileRowStruct{File: files, Folder: folders}
}
