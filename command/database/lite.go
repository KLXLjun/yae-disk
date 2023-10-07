package database

import (
	"YaeDisk/logx"
	"YaeDisk/model"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var dbClient *sql.DB

func Start() bool {
	db, err := sql.Open("sqlite3", "save.db")

	if err != nil {
		logx.Error("载入数据库错误", err)
		return false
	}

	dbClient = db
	return true
}

func Close() {
	if dbClient != nil {
		dbClient.Close()
	}
}

func Insert(input *model.FileRowStruct) error {
	stmt, PrepareErr := dbClient.Prepare("INSERT INTO FileTable(ID, IsDir, OwnerFolder, OwnerID, OwnerGroup, Name, Size, ChangeTime, CreateTime) values(?,?,?,?,?,?,?,?,?)")
	if PrepareErr != nil {
		logx.Error("初始化插入数据错误", PrepareErr)
		return PrepareErr
	}

	logx.Trace(input.OwnerFolder)
	_, ExecErr := stmt.Exec(input.ID, input.IsDir, input.OwnerFolder, input.OwnerID, input.OwnerGroup, input.Name, input.Size, input.ChangeTime, input.CreateTime)
	if ExecErr != nil {
		logx.Error("插入数据错误", ExecErr)
		return PrepareErr
	}
	return nil
}

func LoadAll() {
	rows, err := dbClient.Query("SELECT * FROM FileTable")

	if err != nil {
		logx.Error(err)
	}

	defer rows.Close()

	for rows.Next() {

		var Id string
		var IsDir bool
		var OwnerFolder string
		var OwnerID string
		var OwnerGroup string
		var Name string
		var Size int64
		var ChangeTime time.Time
		var CreateTime time.Time

		err = rows.Scan(&Id, &IsDir, &OwnerFolder, &OwnerID, &OwnerGroup, &Name, &Size, &ChangeTime, &CreateTime)

		if err != nil {
			logx.Error(err)
		}

		All = append(All, &model.FileRowStruct{
			ID:          Id,
			IsDir:       IsDir,
			OwnerFolder: OwnerFolder,
			OwnerID:     OwnerID,
			OwnerGroup:  OwnerGroup,
			Name:        Name,
			Size:        Size,
			ChangeTime:  ChangeTime,
			CreateTime:  CreateTime,
		})
	}

	map2Tree()
}
