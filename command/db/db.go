package db

import (
	"YaeDisk/command/pack"
	"YaeDisk/logx"
	"YaeDisk/model"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"github.com/xujiajun/nutsdb"
	"log"
	"strconv"
	"sync"
)

var dbClient *nutsdb.DB

const FileBucket string = "FileBucket"
const FolderBucket string = "FolderBucket"
const UserBucket string = "UserBucket"
const ConfigBucket string = "ConfigBucket"

var fileIDCounter uint64
var fileIDCouterLock sync.Mutex
var folderIDCounter uint64
var folderIDCounterLock sync.Mutex
var userIDCounter uint64
var userIDCounterLock sync.Mutex

func getID(idType string) uint64 {
	switch idType {
	case FileBucket:
		fileIDCouterLock.Lock()
		defer fileIDCouterLock.Unlock()
		fileIDCounter = fileIDCounter + 1
		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				Key := []byte("FileIDCount")
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(fileIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
		return fileIDCounter

	case FolderBucket:
		folderIDCounterLock.Lock()
		defer folderIDCounterLock.Unlock()
		folderIDCounter = folderIDCounter + 1
		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				Key := []byte("FolderIDCount")
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(folderIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
		return folderIDCounter

	case UserBucket:
		userIDCounterLock.Lock()
		defer userIDCounterLock.Unlock()
		userIDCounter = userIDCounter + 1
		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				Key := []byte("UserIDCount")
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(userIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
		return userIDCounter
	}
	logx.Warn("ID 计数器输入未知类型", idType)
	return 0
}

var FileList = new(FileMap)

type FileMap struct {
	file map[uint64]model.FileStruct
	sync.RWMutex
}

func (m *FileMap) Get(fileID uint64) *model.FileStruct {
	m.RLock()
	defer m.RUnlock()
	if val, ok := m.file[fileID]; ok {
		return &val
	} else {
		return nil
	}
}

func (m *FileMap) Set(fileID uint64, file model.FileStruct) {
	m.Lock()
	m.file[fileID] = file
	defer m.Unlock()
}

func (m *FileMap) Init(input *nutsdb.Entries) {
	m.Lock()
	m.file = make(map[uint64]model.FileStruct, 0)
	if input != nil {
		for _, entry := range *input {
			parseUint, err := strconv.ParseUint(string(entry.Key), 0, 64)
			if err != nil {
				logx.Warn("转换出现错误", err)
			} else {
				p2 := model.FileStruct{}
				err = msgpack.Unmarshal(entry.Value, &p2) // 将二进制流转化回结构体
				if err != nil {
					logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
				} else {
					m.file[parseUint] = p2
				}
			}
		}
	}
	defer m.Unlock()
}

func (m *FileMap) FolderAllFile(folderID uint64) []model.FileStruct {
	m.RLock()
	defer m.RUnlock()
	rsl := make([]model.FileStruct, 0)
	for _, val := range m.file {
		if val.FolderID == folderID {
			rsl = append(rsl, val)
		}
	}
	return rsl
}

var FolderList = new(FolderMap)

type FolderMap struct {
	folder map[uint64]model.FolderStruct
	sync.RWMutex
}

func (m *FolderMap) Get(folderID uint64) *model.FolderStruct {
	m.RLock()
	defer m.RUnlock()
	if val, ok := m.folder[folderID]; ok {
		return &val
	} else {
		return nil
	}
}

func (m *FolderMap) Set(folderID uint64, folder model.FolderStruct) {
	m.Lock()
	m.folder[folderID] = folder
	defer m.Unlock()
}

func (m *FolderMap) SearchHave(OwnerFolderID uint64, Name string) bool {
	m.RLock()
	defer m.RUnlock()
	for _, folderStruct := range m.folder {
		if folderStruct.OwnerFolderID == OwnerFolderID && folderStruct.FolderName == Name {
			return true
		}
	}
	return false
}

func (m *FolderMap) PathSearch(path []string) (bool, *model.FolderStruct, []model.FileStruct, []model.FolderStruct) {
	m.RLock()
	defer m.RUnlock()
	rsl := make([]model.FolderStruct, 0)
	lastFolder := model.FolderStruct{
		FolderID:      0,
		FolderName:    "Root",
		CreateTime:    0,
		OwnerFolderID: 0,
		OwnerUserID:   0,
	}
	//logx.Debug(len(path))
	if len(path) == 0 {
		for _, i2 := range m.folder {
			if i2.OwnerFolderID == 0 {
				logx.Debug(i2.OwnerFolderID, i2.FolderName)
				rsl = append(rsl, i2)
			}
		}
		filersl := FileList.FolderAllFile(0)
		return true, &lastFolder, filersl, rsl
	} else {
		PLen := len(path)
		count := 0
		for PLen > 0 {
			for _, folderStruct := range m.folder {
				if folderStruct.OwnerFolderID == lastFolder.FolderID && folderStruct.FolderName == path[count] {
					lastFolder = folderStruct
					break
				} else {
					logx.Debug(folderStruct.OwnerFolderID == lastFolder.FolderID, folderStruct.FolderName, path[count])
				}
			}
			PLen--
		}
		if lastFolder.FolderID == 0 {
			return false, nil, nil, nil
		}
		for _, folderStruct := range m.folder {
			if folderStruct.OwnerFolderID == lastFolder.FolderID {
				rsl = append(rsl, folderStruct)
			}
		}
		filersl := FileList.FolderAllFile(lastFolder.FolderID)
		return true, &lastFolder, filersl, rsl
	}
}

func (m *FolderMap) Init(input *nutsdb.Entries) {
	m.Lock()
	m.folder = make(map[uint64]model.FolderStruct, 0)
	if input != nil {
		for _, entry := range *input {
			logx.Debug(entry.Key, string(entry.Key))
			parseUint, err := strconv.ParseUint(string(entry.Key), 10, 64)
			if err != nil {
				logx.Warn("转换出现错误", err)
			} else {
				var p2 = model.FolderStruct{}
				err := msgpack.Unmarshal(entry.Value, &p2) // 将二进制流转化回结构体
				if err != nil {
					logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
				} else {
					m.folder[parseUint] = p2
				}
			}
		}
	}
	defer m.Unlock()
}

func InitDB() {
	opt := nutsdb.DefaultOptions
	opt.Dir = "./nutsdb" //这边数据库会自动创建这个目录文件
	db, err := nutsdb.Open(opt)
	if err != nil {
		logx.Error(err)
	}
	dbClient = db
	basicConfig()
	loadFileAndFolder()
}

func Close() {
	dbClient.Close()
}

func InsFile() {

}

func UpFile() {

}

func InsFolder(folderID uint64, folderName string) (bool, uint64, string) {
	if FolderList.SearchHave(folderID, folderName) {
		return false, 0, "文件夹已存在"
	}
	FID := getID(FolderBucket)
	val := model.FolderStruct{
		FolderID:      FID,
		FolderName:    folderName,
		CreateTime:    0,
		OwnerFolderID: folderID,
		OwnerUserID:   0,
	}
	if err := dbClient.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(FolderBucket, []byte(strconv.FormatUint(FID, 10)), pack.EncodePack(val), 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}

	FolderList.Set(FID, val)
	return true, FID, ""
}

func UpFolder() {

}

func basicConfig() {
	Key := []byte("FileIDCount")
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(ConfigBucket, Key); err != nil {
				return err
			} else {
				var p2 uint64
				err = msgpack.Unmarshal(e.Value, &p2) // 将二进制流转化回结构体
				if err != nil {
					logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
				} else {
					fileIDCounter = p2
				}
			}
			return nil
		}); err != nil {

		logx.Warn("配置文件 FileIDCount 缺失", err)
		fileIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(fileIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
	}

	Key = []byte("FolderIDCount")
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(ConfigBucket, Key); err != nil {
				return err
			} else {
				var p2 uint64
				err = msgpack.Unmarshal(e.Value, &p2) // 将二进制流转化回结构体
				if err != nil {
					logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
				} else {
					folderIDCounter = p2
				}
			}
			return nil
		}); err != nil {
		logx.Warn("配置文件 FolderIDCount 缺失", err)
		folderIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(folderIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
	}

	Key = []byte("UserIDCount")
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(ConfigBucket, Key); err != nil {
				return err
			} else {
				var p2 uint64
				err = msgpack.Unmarshal(e.Value, &p2) // 将二进制流转化回结构体
				if err != nil {
					logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
				} else {
					userIDCounter = p2
				}
			}
			return nil
		}); err != nil {
		logx.Warn("配置文件 UserIDCount 缺失", err)
		userIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, pack.EncodePack(userIDCounter), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
	}

	Key = []byte("0")
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			if _, err := tx.Get(UserBucket, Key); err != nil {
				return err
			}
			return nil
		}); err != nil {
		logx.Warn("默认用户 0 缺失", err)
		val := model.UserStruct{
			UserID:   0,
			UserName: "root",
			Pass:     "123456",
		}

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(UserBucket, Key, pack.EncodePack(val), 0); err != nil {
					return err
				}
				return nil
			}); err != nil {
			log.Fatal(err)
		}
	}

	logx.Debug("文件ID计数器:", fileIDCounter)
	logx.Debug("文件夹ID计数器:", folderIDCounter)
}

func loadFileAndFolder() {
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(FileBucket)
			if err != nil {
				return err
			}

			FileList.Init(&entries)
			return nil
		}); err != nil {
		if err == nutsdb.ErrBucketEmpty {
			FileList.Init(nil)
		} else {
			logx.Error("载入文件列表出错", err)
		}
	}

	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(FolderBucket)
			if err != nil {
				return err
			}

			FolderList.Init(&entries)
			return nil
		}); err != nil {
		if err == nutsdb.ErrBucketEmpty {
			FolderList.Init(nil)
		} else {
			logx.Error("载入文件夹列表出错", err)
		}
	}
}
