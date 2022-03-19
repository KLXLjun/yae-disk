package db

import (
	"YaeDisk/command/pack"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"github.com/xujiajun/nutsdb"
	"log"
	"sync"
	"time"
)

var dbClient *nutsdb.DB

const FileBucket string = "FileBucket"
const FolderBucket string = "FolderBucket"
const UserBucket string = "UserBucket"
const ConfigBucket string = "ConfigBucket"
const FilePathBucket string = "FilePathBucket"

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
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(fileIDCounter), 0); err != nil {
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
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(folderIDCounter), 0); err != nil {
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
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(userIDCounter), 0); err != nil {
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
			parseUint := utils.BytesToInt64(entry.Key)
			p2 := model.FileStruct{}
			err := msgpack.Unmarshal(entry.Value, &p2) // 将二进制流转化回结构体
			if err != nil {
				logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
			} else {
				logx.Debug("file", parseUint, p2)
				m.file[parseUint] = p2
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

func (m *FolderMap) PathSearch(path []string) (bool, model.FolderStruct, []model.FileStruct, []model.FolderStruct) {
	m.RLock()
	defer m.RUnlock()
	rsl := make([]model.FolderStruct, 0)
	lastFolder := model.FolderStruct{
		FolderID:      0,
		FolderName:    "Root",
		CreateTime:    "2022-03-12 20:47:58",
		ChangeTime:    "2022-03-12 20:47:58",
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
		return true, lastFolder, filersl, rsl
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
			return false, lastFolder, nil, nil
		}
		for _, folderStruct := range m.folder {
			if folderStruct.OwnerFolderID == lastFolder.FolderID {
				rsl = append(rsl, folderStruct)
			}
		}
		fileRsl := FileList.FolderAllFile(lastFolder.FolderID)
		return true, lastFolder, fileRsl, rsl
	}
}

func (m *FolderMap) Init(input *nutsdb.Entries) {
	m.Lock()
	m.folder = make(map[uint64]model.FolderStruct, 0)
	if input != nil {
		for _, entry := range *input {
			logx.Debug(entry.Key, utils.BytesToInt64(entry.Key))
			parseUint := utils.BytesToInt64(entry.Key)
			var p2 = model.FolderStruct{}
			err := msgpack.Unmarshal(entry.Value, &p2) // 将二进制流转化回结构体
			if err != nil {
				logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
			} else {
				m.folder[parseUint] = p2
			}
		}
	}
	defer m.Unlock()
}

var FilePathList = new(FilePathMap)

type FilePathMap struct {
	folder map[uint64]string
	sync.RWMutex
}

func (m *FilePathMap) Get(folderID uint64) string {
	m.RLock()
	defer m.RUnlock()
	if val, ok := m.folder[folderID]; ok {
		return val
	} else {
		return ""
	}
}

func (m *FilePathMap) Set(folderID uint64, filepath string) {
	m.Lock()
	m.folder[folderID] = filepath
	defer m.Unlock()
}

func (m *FilePathMap) Init(input *nutsdb.Entries) {
	m.Lock()
	m.folder = make(map[uint64]string, 0)
	if input != nil {
		for _, entry := range *input {
			logx.Debug(entry.Key, utils.BytesToInt64(entry.Key))
			parseUint := utils.BytesToInt64(entry.Key)
			var p2 = ""
			err := msgpack.Unmarshal(entry.Value, &p2) // 将二进制流转化回结构体
			if err != nil {
				logx.Warn(fmt.Sprintf("msgpack unmarshal failed,err:%v", err))
			} else {
				m.folder[parseUint] = p2
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

func InsFile(folderID uint64, file model.FileStruct) bool {
	updateSet := model.FileStruct{}
	var FID uint64 = 0
	if have, rsl := HaveFile(folderID, file.FileName); have {
		logx.Debug("检测到相同文件", rsl)
		sets := rsl
		sets.ChangeTime = file.ChangeTime
		sets.FileSize = file.FileSize
		FID = sets.FileID
		updateSet = sets
	} else {
		FID = getID(FileBucket)
		sets := file
		sets.FileID = FID
		updateSet = sets
	}
	logx.Debug("save", FID, updateSet)
	if err := dbClient.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(FileBucket, utils.UInt64ToBytes(FID), pack.EncodePack(updateSet), 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		logx.Warn("创建文件/更新文件出现错误", err)
		return false
	}
	FileList.Set(FID, updateSet)
	return true
}

func HaveFile(folderID uint64, fileName string) (bool, model.FileStruct) {
	list := FileList.FolderAllFile(folderID)
	for _, fileStruct := range list {
		if fileStruct.FileName == fileName {
			return true, fileStruct
		}
	}
	return false, model.FileStruct{}
}

func HaveFolder(folderID uint64) bool {
	if folderID == 0 {
		return true
	}
	return FolderList.Get(folderID) == nil
}

func InsFolder(folderID uint64, folderName string) (bool, uint64, string) {
	if FolderList.SearchHave(folderID, folderName) {
		return false, 0, "文件夹已存在"
	}
	FID := getID(FolderBucket)
	tm := time.Now()
	nowtime := tm.Format("2006-01-02 15:04:05")
	val := model.FolderStruct{
		FolderID:      FID,
		FolderName:    folderName,
		CreateTime:    nowtime,
		ChangeTime:    nowtime,
		OwnerFolderID: folderID,
		OwnerUserID:   0,
	}
	if err := dbClient.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(FolderBucket, utils.UInt64ToBytes(FID), pack.EncodePack(val), 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		logx.Warn("创建文件夹发生错误", err)
	}

	FolderList.Set(FID, val)
	return true, FID, ""
}

func basicConfig() {
	Key := []byte("FileIDCount")
	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(ConfigBucket, Key); err != nil {
				return err
			} else {
				fileIDCounter = utils.BytesToInt64(e.Value)
			}
			return nil
		}); err != nil {

		logx.Warn("配置文件 FileIDCount 缺失", err)
		fileIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(fileIDCounter), 0); err != nil {
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
				folderIDCounter = utils.BytesToInt64(e.Value)
			}
			return nil
		}); err != nil {
		logx.Warn("配置文件 FolderIDCount 缺失", err)
		folderIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(folderIDCounter), 0); err != nil {
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
				userIDCounter = utils.BytesToInt64(e.Value)
			}
			return nil
		}); err != nil {
		logx.Warn("配置文件 UserIDCount 缺失", err)
		userIDCounter = 0

		if err := dbClient.Update(
			func(tx *nutsdb.Tx) error {
				if err := tx.Put(ConfigBucket, Key, utils.UInt64ToBytes(userIDCounter), 0); err != nil {
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

	if err := dbClient.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(FilePathBucket)
			if err != nil {
				return err
			}

			FilePathList.Init(&entries)
			return nil
		}); err != nil {
		if err == nutsdb.ErrBucketEmpty {
			FilePathList.Init(nil)
		} else {
			logx.Error("载入文件列表出错", err)
		}
	}
}
