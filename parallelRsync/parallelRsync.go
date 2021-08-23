package parallelRsync

import (
	"context"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var maxGoroutineNums = runtime.NumCPU()

const maxChanSize = 10240

type pRsyncAction struct {
	src         string      // source path
	dest        string      // destination path
	copyDir     string      // when src=/home/test, copyDir=test
	filesChan   chan string // walkDir() get and store files' path in channel
	ctx         context.Context
	cancel      context.CancelFunc
	wgFind      sync.WaitGroup // use for dir
	wg          sync.WaitGroup // use for file
	findFileCnt uint64         // used to show copying progress
	completeCnt uint64         // used to show copying progress
}

func (pRsync *pRsyncAction) replaceWildcard(old string) string {
	old = strings.ReplaceAll(old, "$", "\\$")
	old = strings.ReplaceAll(old, "\"", "\\\"")
	return old
}

func (pRsync *pRsyncAction) generateRsyncCmd(filePath string) (string, error) {
	destinationPath := pRsync.matchDestination(filePath)

	tmpSrc := pRsync.replaceWildcard(filePath)
	tmpDest := pRsync.replaceWildcard(destinationPath)
	return fmt.Sprintf("rsync -aA --stats \"%s\" \"%s\"", tmpSrc, tmpDest), nil
}

func (pRsync *pRsyncAction) pRsyncExe() {
	defer pRsync.wg.Done()
LOOP:
	for {
		select {
		case <-pRsync.ctx.Done():
			break LOOP
		default:
			filePath, ok := <-pRsync.filesChan
			if !ok {
				break LOOP
			}
			cmd, err := pRsync.generateRsyncCmd(filePath)
			if err != nil {
				log.Println(err)
				break LOOP
			}
			_ = runner.Execute(cmd)
			//if err != nil {
			//	log.Println(err)
			//	//break LOOP
			//}
			atomic.AddUint64(&pRsync.completeCnt, 1)
		}
	}
}

/*
 * example: src=/vol8/home dest=/vol8 copyDir=home srcDirPath=/vol7/home/test/test1
 * split(srcDirPath, src) and get sub directory "/test/test1"
 * return  dest/copyDir/${sub directory} equal "/vol8/home/test/test1"
 */
func (pRsync *pRsyncAction) matchDestination(filePath string) string {
	subDir := strings.Split(filePath, pRsync.src)
	destDirPath := path.Join(pRsync.dest, pRsync.copyDir, subDir[1])
	return destDirPath
}

func (pRsync *pRsyncAction) progressShow() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
LOOP:
	for {
		select {
		case <-pRsync.ctx.Done():
			break LOOP
		case <-ticker.C:
			fmt.Printf("\r%d/%d", pRsync.completeCnt, pRsync.findFileCnt)
		}
	}
}

// PRsyncStart  external func only can use this entrance
func PRsyncStart(source string, destination string) error {
	pRsync := pRsyncAction{src: source, dest: destination, copyDir: filepath.Base(source),
		filesChan: make(chan string, maxChanSize), completeCnt: 0, findFileCnt: 1}
	pRsync.ctx, pRsync.cancel = context.WithCancel(context.Background())
	defer func() {
		pRsync.cancel()
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("source = %s, destination = %s, copyDir = %s\n", pRsync.src, pRsync.dest, pRsync.copyDir)
	pRsync.wgFind.Add(1)
	go pRsync.find(source)
	go pRsync.progressShow()
	go func() {
		pRsync.wgFind.Wait()
		close(pRsync.filesChan)
	}()
	for i := 0; i < maxGoroutineNums; i++ {
		pRsync.wg.Add(1)
		go pRsync.pRsyncExe()
	}
	pRsync.wg.Wait()
	time.Sleep(time.Second)
	return nil
}
