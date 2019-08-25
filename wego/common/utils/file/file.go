package utils

import (
	"bufio"
	"io/ioutil"
	"os"
)

type FileControlType int

const (
	IOUTIL FileControlType = 1
	OS     FileControlType = 2
	BUFIO  FileControlType = 3
)

type File struct {
	Filename   string
	flag       int         // 操作权限
	permission os.FileMode // 文件权限,只有在创建文件才会生效
	file       *os.File
	reader     *bufio.Reader
	writer     *bufio.Writer
}

/*
flag
	1 O_RDONLY:  只读
	2 O_WRONLY:  只写
	3 O_RDWR: 读写
	4 O_APPEND: 追加
	5 O_CREATE: 不存在，则创建
	6 O_EXCL:如果文件存在，且标定了O_CREATE的话，则产生一个错误
	7 O_TRUNG:如果文件存在，且它成功地被打开为只写或读写方式，将其长度裁剪唯一。（覆盖）
	8 O_NOCTTY如果文件名代表一个终端设备，则不把该设备设为调用进程的控制设备
	9 O_NONBLOCK:如果文件名代表一个FIFO,或一个块设备，字符设备文件，则在以后的文件及I/O操作中置为非阻塞模式
	10 O_SYNC:当进行一系列写操作时，每次都要等待上次的I/O操作完成再进行
*/

func GetReadFileHandle(filename string) (f File, e error) {
	_, e = os.Open(filename)
	if e == nil {
		e = f.Init(filename, os.O_RDONLY, 0777)
	}
	return f, e
}

func GetWriteFileHandle(filename string) (f File, e error) {
	_, e = os.Open(filename)
	if e == nil {
		e = f.Init(filename, os.O_WRONLY, 0777)
	} else {
		e = f.Init(filename, os.O_CREATE|os.O_WRONLY, 0777)
	}
	return f, e
}

func GetAppendFileHandle(filename string) (f File, e error) {
	_, e = os.Open(filename)
	if e == nil {
		e = f.Init(filename, os.O_WRONLY|os.O_APPEND, 0777)
	} else {
		e = f.Init(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	}
	return f, e
}

func WriteStringToFile(filename string, all string) (e error) {
	data := []byte(all)
	e = WriteByteToFile(filename, data)
	return e
}
func WriteByteToFile(filename string, data []byte) (e error) {
	e = ioutil.WriteFile(filename, data, 0644)
	return e
}

func AppendStringToFile(filename string, all string) (e error) {
	data := []byte(all)
	e = AppendByteToFile(filename, data)
	return e
}
func AppendByteToFile(filename string, data []byte) (e error) {
	datar, er := ReadByteFromFile(filename)
	if er == nil {
		datar = append(data, datar...)
		e = WriteByteToFile(filename, datar)
	} else {
		e = er
	}
	return e
}

func ReadStringFromFile(filename string) (all string, e error) {
	var data []byte
	data, e = ReadByteFromFile(filename)
	if e == nil {
		all = string(data)
	}
	return all, e
}

func ReadByteFromFile(filename string) (data []byte, e error) {
	data, e = ioutil.ReadFile(filename)
	return data, e
}

func (F *File) Init(f string, fl int, p os.FileMode) (e error) {
	F.Filename = f
	F.flag = fl
	F.permission = p
	F.file, e = os.OpenFile(f, fl, p)
	switch fl & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		F.reader = bufio.NewReader(F.file)
	case os.O_WRONLY:
		F.writer = bufio.NewWriter(F.file)
	case os.O_RDWR:
		F.writer = bufio.NewWriter(F.file)
		F.reader = bufio.NewReader(F.file)
	}
	return e
}

// read all
func (F *File) ReadAll() (all string, e error) {
	var bt []byte
	bt, e = F.ReadAllByte()
	return string(bt), e
}

// read a line
func (F *File) ReadLine() (line string, e error) {
	bt, e := F.reader.ReadString('\n')
	if e == nil && len(bt) >= 1 {
		line = string(bt[:len(bt)-1])
	}
	return line, e
}

// write all
func (F *File) WriteAll(all string) (e error) {
	_, e = F.writer.WriteString(all)
	return e
}

// write a line
func (F *File) WriteLine(line string) (e error) {
	_, e = F.writer.WriteString(line + "\n")
	return e
}

// read all by byte
func (F *File) ReadAllByte() (all []byte, e error) {
	var b []byte
	for {
		if b, e = F.reader.ReadBytes('\n'); e != nil {
			break
		} else {
			all = append(all, b...)
		}
	}
	return all, nil

}

// read a line by byte
func (F *File) ReadLineByte() (line []byte, e error) {
	line, e = F.reader.ReadBytes('\n')
	if e == nil && len(line) >= 1 {
		line = line[:len(line)-1]
	}
	return line, e

}

// write all by byte
func (F *File) WriteAllByte(line []byte) (e error) {
	_, e = F.writer.Write(line)
	return e

}

// write a line by byte
func (F *File) WriteLineByte(line []byte) (e error) {
	line = append(line, '\n')
	_, e = F.writer.Write(line)
	return nil

}

func (F *File) Close() (e error) {
	if F.writer != nil {
		F.writer.Flush()
	}
	e = F.file.Close()
	return e
}
