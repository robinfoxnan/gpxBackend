package db

import (
	"fmt"
	"github.com/gocql/gocql"
	"zhituBackend/common"
)

var session *gocql.Session

// 初始化
func InitCassandra() {
	cluster := gocql.NewCluster("10.128.6.104:9042")
	cluster.Keyspace = "robindb"
	cluster.Consistency = gocql.Consistency(1)
	cluster.NumConns = 3
	//cluster.Authenticator = gocql.PasswordAuthenticator{Username: "test", Password: "testpwd"}
	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		//log.Panic("start", err)
		common.Logger.Panic("can't connect to server cassandra:")
		return
	}
}

// 创建表
func CreateTable() {
	query := fmt.Sprintf(`create table u_inbox(p int, id bigint, tm bigint, c text, PRIMARY KEY(p, id, tm ) );`)
	session.Query(query).Exec()
}

// 插入数据
func InsertInbox() {
	query := fmt.Sprintf(`insert into u_inbox(p, id, tm, c)  VALUES(1, 13610501000, 6, 'test it');`)
	err := session.Query(query).Exec()
	if err != nil {
		fmt.Println(err)
	}
}

//queryExec("INSERT INTO USERS VALUES(?,?,?,?)", userId, emailId, mobileNo, gender)

func InsertDataCommon(query string, args ...interface{}) (err error) {
	err = session.Query(query, args).Exec()
	return err
}

// 查询数据
func FindInboxData() {
	query := fmt.Sprintf("SELECT * from u_inbox;")
	iter := session.Query(query).Iter()
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()
	var p int
	var id int64
	var tm int64
	var msg string
	for iter.Scan(&p, &id, &tm, &msg) {
		fmt.Println(id, msg)
	}
}
