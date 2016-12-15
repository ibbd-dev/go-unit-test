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

// 正在运行的项目名称
var runningProjectName string

// 项目的启动时间
var startTime time.Time

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("./index.tmpl")

	r.GET("/index", showIndex)
	r.GET("/action/:projectName/:action", processProject)

	s := &http.Server{
		Addr:    ":" + fmt.Sprint(mainPort),
		Handler: r,
	}

	fmt.Printf("Start from: http://%s:%d/index", host, mainPort)
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

// showIndex 显示首页
func showIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"host":     host,
		"port":     fmt.Sprint(mainPort),
		"time":     time.Now().Unix(),
		"projects": projects,
	})
}

// startSuccess 启动成功之后，跳转到目标地址
func startSuccess(c *gin.Context) {
	url := fmt.Sprintf("http://%s:%d\n", host, processPort)
	//c.Redirect(http.StatusFound, url)
	c.Redirect(http.StatusMovedPermanently, url)
}

// processProject 处理单元测试项目
func processProject(c *gin.Context) {
	defer c.Request.Body.Close()

	prjName := c.Param("projectName")
	_, prj, err := getProject(prjName)
	if err != nil {
		c.String(http.StatusBadRequest, "BAD Project name: %s", prjName)
		return
	}

	action := c.Param("action")
	switch action {
	case actionShow:
		pid, err := getPid()
		if err != nil {
			c.String(http.StatusInternalServerError, "show action: %s", err.Error())
			return
		}
		if len(pid) > 2 {
			if runningProjectName != prj.Name {
				// 进程已经启动，但是运行的不是当前的项目
				stopProcess(pid)
			}
		}

		// 创建新的单元测试
		if err = startProcess(prj); err != nil {
			c.String(http.StatusInternalServerError, "show action: %s", err.Error())
			return
		}

		// 输出
		startSuccess(c)
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
		// 只是启动单元测试项目
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
		startSuccess(c)
		return

	default:
		c.String(http.StatusBadRequest, "BAD action name: %s", action)
	}
}

func getProject(prjName string) (key int, prj Project, err error) {
	for key, prj = range projects {
		if prj.Name == prjName {
			return key, prj, nil
		}
	}

	return key, prj, errors.New("project is not existed for name: " + prjName)
}

func startProcess(prj Project) error {
	cmdStr := fmt.Sprintf("cd %s; $GOPATH/bin/goconvey -host %s -port %d", prj.Path, host, processPort)
	fmt.Printf("cmd: %s\n", cmdStr)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	// 这里需要异步
	//go cmd.Run()
	err := cmd.Run()
	if err != nil {
		return err
	}

	for i := 0; i < 120; i++ {
		time.Sleep(time.Second)
		outBts, err := cmd.Output()
		if err != nil {
			return err
		}
		out := string(outBts)
		fmt.Printf("%s", out)
	}

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
