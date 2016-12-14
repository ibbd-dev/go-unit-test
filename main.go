package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ibbd-dev/go-tools/timer"
)

const (
	mainPort = "8188"
)

var runningProjectName string
var startTime time.Time

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("./index.tmpl")

	r.GET("/index", showIndex)
	r.GET("/action/:projectName/:action", processProject)

	s := &http.Server{
		Addr:    ":" + mainPort,
		Handler: r,
	}

	s.ListenAndServe()

	// 定期清理过期的进程
	startTime = time.Now()
	timer.AddFunc(func() {
		now := time.Now()
		for _, v := range projects {
			if v.Name == runningProjectName && startTime.Add(closeDuration).Before(now) {
				pid, err := getPid()
				if err != nil {
					fmt.Printf("duration getId error\n")
					return
				}
				if err = stopProcess(pid); err != nil {
					fmt.Printf("duration stop error: %s\n", pid)
					return
				}

				fmt.Printf("duration stop success: %s\n", pid)
			}
		}
	}, time.Minute)
}

func showIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"host":     processHost,
		"port":     processPort,
		"projects": projects,
	})
}

func processProject(c *gin.Context) {
	defer c.Request.Body.Close()

	prjName := c.Param("projectName")
	id, prj, err := getProject(prjName)
	if err != nil {
		c.String(http.StatusBadRequest, "BAD Project name: %s", prjName)
	}

	action := c.Param("action")
	switch action {
	case actionShow:
		showProcess(projects[id], c)

		// 输出
		c.String(http.StatusOK, "http://%s:%d", processHost, processPort)
		return

	case actionStop:
		pid, err := getPid()
		if err != nil {
			c.String(http.StatusBadRequest, "%s getId error: %s\n", action, err.Error())
			return
		}

		if err = stopProcess(pid); err != nil {
			c.String(http.StatusBadRequest, "%s pid: %s error: %s\n", action, pid, err.Error())
			return
		}

	case actionStart:
		// 创建新的单元测试
		if err = startProcess(prj); err != nil {
			c.String(http.StatusInternalServerError, "showUnitTest startProcess: %s", err.Error())
			return
		}

	case actionRestart:
		pid, err := getPid()
		if err != nil {
			c.String(http.StatusBadRequest, "%s getId error: %s\n", action, err.Error())
			return
		}

		if err = stopProcess(pid); err != nil {
			c.String(http.StatusBadRequest, "%s pid: %s error: %s\n", action, pid, err.Error())
			return
		}

		if err = startProcess(prj); err != nil {
			c.String(http.StatusInternalServerError, "showUnitTest startProcess: %s", err.Error())
			return
		}

		// 输出
		c.String(http.StatusOK, "http://%s:%d", processHost, processPort)
		return

	default:
		c.String(http.StatusBadRequest, "BAD action name: %s", action)
	}

	// 输出
	c.String(http.StatusOK, "action: %s success", action)
}

func getProject(prjName string) (key int, prj Project, err error) {
	for key, prj = range projects {
		if prj.Name == prjName {
			return key, prj, nil
		}
	}

	return key, prj, errors.New("project is not existed for name: " + prjName)
}

func showProcess(prj Project, c *gin.Context) {
	pid, err := getPid()
	if err != nil {
		c.String(http.StatusInternalServerError, "showUnitTest getPid: %s", err.Error())
		return
	}
	if len(pid) > 2 {
		if runningProjectName != prj.Name {
			// 进程已经启动，但是运行的不是当前的项目
			stopProcess(pid)
		}
	} else {
		// 创建新的单元测试
		if err = startProcess(prj); err != nil {
			c.String(http.StatusInternalServerError, "showUnitTest startProcess: %s", err.Error())
		}
	}
}

func startProcess(prj Project) error {
	cmdStr := fmt.Sprintf("cd %s; $GOPATH/bin/goconvey -host %s -port %d", prj.Path, processHost, processPort)
	fmt.Printf("cmd: %s\n", cmdStr)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	// 这里需要异步
	go cmd.Run()
	startTime = time.Now()
	runningProjectName = prj.Name
	return nil
}

func stopProcess(pid string) error {
	cmd := exec.Command("/bin/bash", "-c", "kill -9 "+pid)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf("kill success pid: %s\n", pid)
	runningProjectName = ""

	return nil
}

func getPid() (pid string, err error) {
	// 判断是否有正在运行的单元测试，如果有则停止
	cmd := exec.Command("/bin/bash", "-c", "ps -A |grep 'goconvey'| awk '{print $1}'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return pid, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return pid, err
	}

	if err := cmd.Start(); err != nil {
		return pid, err
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return pid, err
	}

	if len(bytesErr) > 0 {
		return pid, errors.New("stderr is not nil: " + string(bytesErr))
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return pid, err
	}

	pid = string(bytes)
	if len(pid) < 3 {
		fmt.Println("not running")
	}

	return pid, nil
}
