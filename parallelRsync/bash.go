package parallelRsync

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

type Run interface {
	Execute(cmd string) error
}

type run struct {
}

// TODO: need some code sharing here!!!
func (*run) Execute(cmdStr string) error {
	//fmt.Printf("Local command : %s\n", cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)

	timer := time.AfterFunc(time.Minute*5, func() {
		log.Println("Time up, waited more than 5 mins to complete.")
		if err := cmd.Process.Kill(); err != nil {
			log.Panicf("error trying to kill process: %s", err.Error())
		}
	})

	output, err := cmd.CombinedOutput()
	timer.Stop()

	if err == nil {
		//fmt.Println("Completed local command run: Local command : ", cmd)
		//log.Println(string(output))
		return nil
	} else {
		fmt.Printf("Error in local command run: '%s' error: %s\n", cmdStr, err.Error())
		log.Println(string(output))
		return err
		//return fmt.Errorf("%s : %s\n", err, err.Error())
	}
}

var runner Run = &run{}
