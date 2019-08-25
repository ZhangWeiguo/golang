package client

import (
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DefaultConnNum = 300
)

// Warn 返回的结果Row务必要进行关闭row.Close()
type MySQLClient struct {
	Host        string
	Port        string
	DbName      string
	UserName    string
	Password    string
	MaxOpenConn int // 最大连接数
	MaxIdleConn int // 最大空闲数
	Qps         int32
	qpsCnt      int32
	Db          *sql.DB
}

func (m *MySQLClient) connect() error {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", m.UserName, m.Password, m.Host, m.Port, m.DbName)
	db, err := sql.Open("mysql", connStr)
	if err == nil {
		m.Db = db
		if m.MaxIdleConn >= 0 && m.MaxOpenConn >= 0 {
			if m.MaxIdleConn >= m.MaxOpenConn {
				m.Db.SetMaxOpenConns(m.MaxOpenConn)
				m.Db.SetMaxIdleConns(m.MaxIdleConn)
			} else {
				m.Db.SetMaxOpenConns(m.MaxOpenConn)
			}
		} else {
			m.Db.SetMaxOpenConns(DefaultConnNum)
		}
		_, err = m.Db.Exec(fmt.Sprintf("USE %s", m.DbName))
		if err != nil {
			return err
		}
		err = m.Db.Ping()
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func (m *MySQLClient) Init() error {
	err := m.connect()
	go m.countQps()
	go m.keepAlive()
	return err
}

func (m *MySQLClient) QueryGetOne(query string, args ...interface{}) (result *sql.Row, err error) {
	if m.Db != nil {
		atomic.AddInt32(&(m.qpsCnt), 1)
		result = m.Db.QueryRow(query, args...)
		return result, nil
	}
	return nil, errors.New("Db Is Nil ")
}

func (m *MySQLClient) QueryGetAll(query string, args ...interface{}) (result *sql.Rows, err error) {
	if m.Db != nil {
		atomic.AddInt32(&(m.qpsCnt), 1)
		result, err = m.Db.Query(query, args...)
		if err == nil {
			if err = result.Err(); err == nil {
				return result, nil
			}
		}
		return nil, err
	}
	return nil, errors.New("Db Is Nil ")
}

func (m *MySQLClient) ExecGetEffect(exec string, args ...interface{}) (result sql.Result, err error) {
	if m.Db != nil {
		atomic.AddInt32(&(m.qpsCnt), 1)
		result, err = m.Db.Exec(exec, args...)
		return result, err
	} else {
		return result, errors.New("Db Is Nil ")
	}
}

func (m *MySQLClient) Template(exec string) (*sql.Stmt, error) {
	return m.Db.Prepare(exec)
}

func (m *MySQLClient) QPS() int32 {
	return m.Qps
}

func (m *MySQLClient) Close() error {
	if m.Db != nil {
		err := m.Db.Close()
		if err != nil {
			return err
		}
		m.Db = nil
	}
	return nil
}

func (m *MySQLClient) keepAlive() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C
		err := m.Db.Ping()
		if err != nil {
			err = m.connect()
			if err != nil {
				break
			}
		}
	}
}

func (m *MySQLClient) countQps() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		m.Qps = m.qpsCnt
		m.qpsCnt = 0
	}
}
