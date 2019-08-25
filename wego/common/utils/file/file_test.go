package utils

import (
	"fmt"
	"testing"
)

func TestWrite(t *testing.T) {
	f, e := GetWriteFileHandle("TestFile")
	if e == nil {
		fmt.Println("Right Open")
		f.WriteLine("I am a he")
		f.WriteLine("I am a she")
		f.WriteLine("I am a it")
	} else {
		fmt.Println("Wrong Open")
	}
	f.Close()
}

func TestAppend(t *testing.T) {
	f, e := GetAppendFileHandle("TestFile")
	if e == nil {
		fmt.Println("Right Open")
		f.WriteLine("append:I am a he")
		f.WriteLine("append:I am a she")
		f.WriteLine("append:I am a it")
	} else {
		fmt.Println("Wrong Open")
	}
	f.Close()
}

func TestRead(t *testing.T) {
	f, e := GetReadFileHandle("TestFile")
	if e == nil {
		fmt.Println("Right Open")
		var s []byte
		for {
			s, e = f.ReadLineByte()
			if e == nil {
				fmt.Println(s)
			} else {
				break
			}
		}
	} else {
		fmt.Println("Wrong Open")
	}
	f.Close()
}

func TestWriteAndReadString(t *testing.T) {
	WriteStringToFile("TestStringFile", "This is your name:doglashi\nI have no name")
	fmt.Println(ReadStringFromFile("TestStringFile"))
}

func TestWriteAndReadByte(t *testing.T) {
	WriteByteToFile("TestByteFile", []byte("This is your name:doglashi\nI have no name"))
	fmt.Println(ReadStringFromFile("TestStringFile"))
	fmt.Println(ReadByteFromFile("TestStringFile"))
}
