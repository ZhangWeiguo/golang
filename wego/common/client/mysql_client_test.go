package client

import (
	"fmt"
	"testing"
)

func TestMySQLClient(t *testing.T) {
	mysqlClient := MySQLClient{
		Host:        "127.0.0.1",
		Port:        "3306",
		DbName:      "test",
		UserName:    "pig",
		Password:    "123456",
		MaxOpenConn: 20,
		MaxIdleConn: 20,
	}
	err := mysqlClient.Init()
	if err != nil {
		panic(err)
	}
	fmt.Println(mysqlClient.ExecGetEffect("insert into ugc_video_in(id,like_num) values('23',56)"))
	rows, _ := mysqlClient.QueryGetAll("select id,like_num from ugc_video_in where like_num>10")
	var id, like string
	for rows.Next() {
		_ = rows.Scan(&id, &like)
		fmt.Println(fmt.Sprintf("id:%15s like:%5s", id, like))
	}
}
