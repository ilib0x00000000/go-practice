package utils


import (
	"io/ioutil"
	"fmt"
	"os"
	"io"
	"log"
	"bufio"
)

/**
	golang 文件的读写
 */



// 直接读取整个文件，适用文件内容比较少的情况
func ReadFile(filename string) {
	context, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("<main.ReadFile error> ", err)
		return
	}

	// fmt.Println(Bytes2str(context))
	fmt.Println(string(context))
}


// 使用 os系统调用操作文件
// 适用读写和其他操作的场景
func OperateFile(filename string) {
	fd, err := os.Open(filename)

	if err != nil {
		fmt.Println("<main.OperateFile error>: ", err)
		return
	}
	defer fd.Close()

	chunks := make([]byte, 1024, 1024)

	for {
		n, err := fd.Read(chunks)

		if err != nil {
			if err != io.EOF {
				// 读取文件失败
				log.Println("读取文件失败")
			} else {
				// read finish
				break
			}
		}

		fmt.Printf("读取到 %d 字节数据： %s\n", n, string(chunks[:n]))
	}
}


// 使用 reader/writer
// 适用场景： 与其他IO读写统一接口
func ReadWriter(filename string) {
	buff := make([]byte, 1024, 1024)
	fd, err := os.Open(filename)
	if err != nil {
		fmt.Println("<main.ReadWriter error>: ", err)
		return
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	for {
		n, err := reader.Read(buff)

		if err != nil {
			if err != io.EOF {
				log.Fatal("读取文件失败<main.ReadWriter> ", err)
			} else {
				break        // 读完
			}
		}

		fmt.Printf("读取到 %d 字节数据： %s\n", n, string(buff[:n]))
	}
}

