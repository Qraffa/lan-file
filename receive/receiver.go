package receive

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	progress *mpb.Progress
	wg       sync.WaitGroup
)

func ReceiveFile(ip string) {
	// 监听tcp连接
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ip))
	if err != nil {
		fmt.Printf("net listen err: %s\n", err)
		return
	}
	defer listener.Close()

	progress = mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.WithWaitGroup(&wg),
	)

	// 阻塞等待第一次 tcp 连接，因为第一次连接传输的是文件数量
	firstConn, err := listener.Accept()
	if err != nil {
		fmt.Print("accept tcp err: %s\n", err)
		return
	}
	files := receiveNumberOfFile(firstConn)
	fmt.Printf("Total: %d file(s).\n", files)
	if files <= 0 {
		return
	}

	wg.Add(int(files))

	finish := 0
	for {
		// 建立tcp新连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept tcp err: %s\n", err)
			return
		}
		go receiveFileHandle(conn)
		finish++
		// 全部文件传输完
		if finish >= int(files) {
			break
		}
	}

	// 统计文件数，接受完成后exit
	progress.Wait()

}

func receiveNumberOfFile(conn net.Conn) uint64 {
	numberBuf := make([]byte, 8)
	n, err := conn.Read(numberBuf)
	if err != nil {
		fmt.Printf("read number of file err: %s\n", err)
		return 0
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		fmt.Printf("write filename resp err: %s\n", err)
		return 0
	}
	return binary.BigEndian.Uint64(numberBuf[:n])
}

func receiveFileHandle(conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	nameBuf := make([]byte, 4096)
	n, err := conn.Read(nameBuf)
	if err != nil {
		fmt.Printf("read filename err: %s\n", err)
		return
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		fmt.Printf("write filename resp err: %s\n", err)
		return
	}

	filenameAndSize := string(nameBuf[:n])
	posSize := strings.LastIndex(filenameAndSize, "@")
	fileSize, _ := strconv.Atoi(filenameAndSize[posSize+1:])
	filename := filenameAndSize[:posSize]
	// filename := fmt.Sprintf("%s-file", time.Now().Format("2006-01-02-15:04:05"))
	// 创建文件夹
	pos := strings.LastIndex(filename, string(os.PathSeparator))
	dirname := filename[:pos]
	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		fmt.Printf("mkdir dir err: %s\n", err)
		return
	}

	// 创建文件写入内容
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		fmt.Printf("create file err: %s\n", err)
		return
	}

	var total int64 = int64(fileSize)

	bar := progress.Add(total,
		mpb.NewBarFiller(mpb.BarStyle().Rbound("|")),
		mpb.PrependDecorators(
			decor.Name(filename, decor.WC{W: len(filename) + 1, C: decor.DidentRight}),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	// create proxy reader
	proxyReader := bar.ProxyReader(conn)
	defer proxyReader.Close()

	// copy from proxyReader, ignoring errors
	io.Copy(file, proxyReader)
	// io.Copy(file, conn)

}

// buf := make([]byte, 64*1024)
// for {
// 	n, err := conn.Read(buf)
// 	if n == 0 {
// 		fmt.Println("read over!")
// 		break
// 	}
// 	if err != nil {
// 		fmt.Printf("read conn err: %s\n", err)
// 		break
// 	}
// 	atomic.AddUint64(counter.Total, uint64(n))
// 	file.Write(buf[:n])
// }
