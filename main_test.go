package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"lan-file/send"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	_ "unsafe"

	"github.com/smartystreets/goconvey/convey"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var dirname = "a"
var target = "hole"

func TestDir(t *testing.T) {
	var files []string
	var dirs []string
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(dirs)
	fmt.Println(files)
}

func TestMkdir(t *testing.T) {
	err := os.MkdirAll("hole/a/b/c", os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func TestSp(t *testing.T) {
	filename := "hole/a/b/c/0.data"
	pos := strings.LastIndex(filename, "/")
	fmt.Println(filename[:pos])
}

var input chan string

type tf struct {
	dir      string
	filename string
}

func TestCopyDir(t *testing.T) {
	var files []tf
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, tf{
				dir:      fmt.Sprintf("%s/%s", target, path[:len(path)-len(d.Name())-1]),
				filename: d.Name(),
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		createDir(file.dir)
		createFile(fmt.Sprintf("%s/%s", file.dir, file.filename))
	}

	// input = make(chan string, len(files))
	// for _, file := range files {
	// 	input <- file
	// }
	// close(input)
	// concurr(3, len(files))
}

func concurr(limit, total int) {
	wg := sync.WaitGroup{}
	wg.Add(total)
	for i := 0; i < limit; i++ {
		go func() {
			for file := range input {
				createFile(file)
				wg.Done()
			}
		}()
	}
	wg.Wait()
}

func createDir(dirname string) {
	err := os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func createFile(filename string) {
	fd, err := os.Create(filename)
	defer fd.Close()
	if err != nil {
		panic(err)
	}
}

func TestFile(t *testing.T) {
	fi, _ := os.Stat("tmp-1000")
	fmt.Printf("fi.Size(): %v\n", fi.Size())

}

func TestBar(t *testing.T) {
	var wg sync.WaitGroup
	// passed &wg will be accounted at p.Wait() call
	p := mpb.New(mpb.WithWaitGroup(&wg))
	total, numBars := 100, 10
	totals := []int{1024 * 1024 * 100, 1024 * 1024 * 500, 1024 * 1024 * 1000, 1024 * 1024 * 100, 1024 * 1024 * 500, 1024 * 1024 * 1000, 1024 * 1024 * 100, 1024 * 1024 * 500, 1024 * 1024 * 1000, 1024 * 1024 * 100, 1024 * 1024 * 500, 1024 * 1024 * 1000}
	_ = total
	wg.Add(numBars)

	for i := 0; i < numBars; i++ {
		go func(i int) {
			name := fmt.Sprintf("Bar#%d:", i)
			bar := p.Add(int64(totals[i]),
				// mpb.PrependDecorators(
				// 	// simple name decorator
				// 	decor.Name(name),
				// 	// decor.DSyncWidth bit enables column width synchronization
				// 	decor.Percentage(decor.WCSyncSpace),
				// ),
				// mpb.AppendDecorators(
				// 	// replace ETA decorator with "done" message, OnComplete event
				// 	decor.OnComplete(
				// 		// ETA decorator with ewma age of 60
				// 		decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WCSyncWidth), "done",
				// 	),
				// ),

				mpb.NewBarFiller(mpb.BarStyle().Rbound("|")),
				mpb.PrependDecorators(
					decor.Name(name),
					decor.CountersKibiByte("% .2f / % .2f"),
				),
				mpb.AppendDecorators(
					decor.EwmaETA(decor.ET_STYLE_GO, 90),
					decor.Name(" ] "),
					decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
				),
			)
			// simulating some work
			defer wg.Done()
			reader := io.LimitReader(rand.Reader, int64(totals[i]))
			// create proxy reader
			proxyReader := bar.ProxyReader(reader)
			defer proxyReader.Close()

			// copy from proxyReader, ignoring errors
			io.Copy(ioutil.Discard, proxyReader)
		}(i)
	}
	// Waiting for passed &wg and for all bars to complete and flush
	p.Wait()

}

func TestFF(t *testing.T) {
	nameBuf := []byte("hole/a/b/c/0.data@102400")
	filenameAndSize := string(nameBuf)
	posSize := strings.LastIndex(filenameAndSize, "@")
	fileSize, _ := strconv.Atoi(filenameAndSize[posSize+1:])
	filename := filenameAndSize[:posSize]
	// filename := fmt.Sprintf("%s-file", time.Now().Format("2006-01-02-15:04:05"))
	// 创建文件夹
	pos := strings.LastIndex(filename, "/")
	dirname := filename[:pos]

	fmt.Println(fileSize, filename, dirname)
}

func TestCch(t *testing.T) {
	ch := make(chan struct{})
	close(ch)
	<-ch
	fmt.Println("go here")
}

func TestCScon(t *testing.T) {
	ch := make(chan error, 4)
	ch <- errors.New("h")
	fmt.Println(len(ch))
}

func Sum(a, b int) int {
	return a + b
}

func TestSum(t *testing.T) {
	convey.Convey("test mock", t, func() {
		convey.Convey("sum", func() {
			Mock(Sum).Return(3).Build()
			ret := Sum(1, 1)
			convey.So(ret, convey.ShouldEqual, 2)
			fmt.Printf("ret: %v\n", ret)
		})
	})
}

//go:linkname sendPrivate send.sendVar
var sendPrivate string

func TestPrivate(t *testing.T) {
	fmt.Println(send.GetSendVar())
	sendPrivate = "asd"
	fmt.Println(send.GetSendVar())
}
