package parallelRsync

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
)

var sema = make(chan struct{}, maxGoroutineNums)

func (pRsync *pRsyncAction) pMkdir(sourcePath string, destinationPath string) error {
	tmpSrc := pRsync.replaceWildcard(sourcePath)
	tmpDest := pRsync.replaceWildcard(destinationPath)
	cmd := fmt.Sprintf("rsync -lptgoDd \"%s\" \"%s\"", tmpSrc, tmpDest)
	if err := runner.Execute(cmd); err != nil {
		return err
	}
	return nil
}

// 枚举一个目录下的所有入口
func (pRsync *pRsyncAction) dirent(dir string) []os.DirEntry {
	file, err := os.ReadDir(dir)
	if err != nil {
		log.Panic(err)
	}
	return file
}

func (pRsync *pRsyncAction) find(parDir string) {
	defer pRsync.wgFind.Done()
	sema <- struct{}{}
	defer func() { <-sema }()
	if err := pRsync.pMkdir(parDir, pRsync.matchDestination(parDir)); err != nil {
		return
		//log.Println(err)
		//pRsync.cancel()
	}
	atomic.AddUint64(&pRsync.completeCnt, 1)
	files := pRsync.dirent(parDir)
	for _, file := range files {
		atomic.AddUint64(&pRsync.findFileCnt, 1)
		filePath := filepath.Join(parDir, file.Name())
		if file.IsDir() {
			pRsync.wgFind.Add(1)
			go pRsync.find(filePath)
		} else {
			pRsync.filesChan <- filePath
		}
	}
}
