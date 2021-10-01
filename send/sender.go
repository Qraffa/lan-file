package send

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"lan-file/util"
)

var sendVar string

func GetSendVar() string {
	return sendVar
}

type fileInfo struct {
	filename string
	filesize int64
}

func SendFile(ip, targetDir, filename string) {
	// 初始化tcp连接池
	util.InitConnPool("tcp", ip)
	// 建立tcp连接
	files := make([]fileInfo, 0)
	// 遍历文件夹，获取全部文件
	err := filepath.WalkDir(filename, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fi, _ := os.Stat(path)
			files = append(files, fileInfo{
				filename: path,
				filesize: fi.Size(),
			})
		}
		return nil
	})
	if err != nil {
		fmt.Printf("walk dir err: %s\n", err)
		return
	}
	input := make(chan fileInfo, len(files))
	for _, f := range files {
		input <- f
	}
	// 限制 goroutine 数量
	var limit, total int
	limit = 16
	total = len(input)

	// 传输文件数量
	sendNumberOfFile(total)
	fmt.Printf("Total: %d file(s).\n", total)

	// 限制并发数 ==》 util
	// tcp conn 复用 ==> conn 连接池
	wg := sync.WaitGroup{}
	wg.Add(total)
	for i := 0; i < limit; i++ {
		go func() {
			for file := range input {
				sendFileHandle(ip, targetDir, file)
				wg.Done()
			}
		}()
	}
	wg.Wait()

	// sendFileHandle(targetDir, filename)
}

// 传输文件数量
func sendNumberOfFile(files int) {
	conn := util.GetConn()
	if conn == nil {
		fmt.Printf("get tcp conn failed.\n")
		return
	}
	defer util.ReleaseConn(conn)

	numberBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(numberBuf, uint64(files))
	if _, err := conn.Write(numberBuf); err != nil {
		fmt.Printf("write filename err: %s\n", err)
		return
	}
	respBuf := make([]byte, 1024)
	n, err := conn.Read(respBuf)
	if err != nil {
		fmt.Printf("receive filename resp err: %s\n", err)
		return
	}
	// 文件数量传输完成
	if string(respBuf[:n]) == "ok" {
		return
	}
}

func sendFileHandle(ip, targetDir string, fileinfo fileInfo) {
	// get tcp conn
	conn := util.GetConn()
	if conn == nil {
		fmt.Printf("get tcp conn failed.\n")
		return
	}
	// defer util.ReleaseConn(conn)

	// 拼接文件夹+文件名+文件大小
	nameBuf := make([]byte, 4096)
	targetFileName := fmt.Sprintf("%s%s%s@%d", targetDir, string(os.PathSeparator), fileinfo.filename, fileinfo.filesize)
	// 传输文件名
	if _, err := conn.Write([]byte(targetFileName)); err != nil {
		fmt.Printf("write filename err: %s\n", err)
		return
	}
	n, err := conn.Read(nameBuf)
	if err != nil {
		fmt.Printf("receive filename resp err: %s\n", err)
		return
	}
	// 文件名传输完成，传输文件内容
	if string(nameBuf[:n]) == "ok" {
		// 打开文件
		file, err := os.Open(fileinfo.filename)
		defer file.Close()
		if err != nil {
			fmt.Printf("open file err: %s\n", err)
			return
		}
		io.Copy(conn, file)
	}
}

// buf := make([]byte, 64*1024)
// // 读取文件写入tcp连接
// for {
// 	//io.Copy()
// 	n, err := file.Read(buf)
// 	if n == 0 {
// 		fmt.Println("read over!")
// 		return
// 	}
// 	if err != nil {
// 		fmt.Printf("read file err: %s\n", err)
// 		return
// 	}

// 	_, err = conn.Write(buf[:n])
// 	if err != nil {
// 		fmt.Printf("conn write err: %s\n", err)
// 		return
// 	}
// }
