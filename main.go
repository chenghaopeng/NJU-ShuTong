package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var EAI_SESS = os.Getenv("EAI_SESS")
var ROOM = strings.ReplaceAll(os.Getenv("ROOM"), " ", "")

var wg sync.WaitGroup

func execScript(name string, env []string) {
	cmd := exec.Cmd{
		Path: "/bin/sh",
		Args: []string{"/bin/sh", "/scripts/" + name + ".sh"},
		Env:  env,
	}
	output, err := cmd.Output()
	log.Println("execScript", name, env, string(output), err)
}

func sendMessage(content string) {
	execScript("message", []string{"CONTENT=" + content})
	log.Println("sendMessage", content)
}

func getElectricBalance(areaId string, buildId string, roomId string) (float64, error) {
	url := fmt.Sprintf("https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=%s&build_id=%s&room_id=%s", areaId, buildId, roomId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Add("Cookie", "eai-sess="+EAI_SESS)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	bodyByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	body := string(bodyByte)
	reg := regexp.MustCompile(`dianyue\w*:\w*"([\d\.]+)"`)
	result := reg.FindStringSubmatch(body)
	if len(result) != 2 {
		return 0, errors.New("登录失效或未找到电量余额")
	}
	return strconv.ParseFloat(result[1], 64)
}

func taskElectricBalance() {
	wg.Add(1)
	defer wg.Done()
	ids := strings.SplitN(ROOM, ",", 3)
	if len(ids) != 3 {
		sendMessage("未设定寝室或者设定有误，不开启寝室电量检测")
		return
	}
	sendMessage("书童开始为你监测寝室电量啦～")
	sentErr, sentEle := false, false
	for {
		electric, err := getElectricBalance(ids[0], ids[1], ids[2])
		log.Println("taskElectricBalance", electric, err)
		if err != nil {
			if !sentErr {
				sentErr = true
				sendMessage("获取寝室电量出错了：" + err.Error())
			}
		} else {
			sentErr = false
			if electric <= 10 {
				if !sentEle {
					sentEle = true
					sendMessage("寝室电量不足 10 度啦～")
				}
			} else {
				sentEle = false
			}
		}
		time.Sleep(time.Minute)
	}
}

func main() {
	go taskElectricBalance()
	time.Sleep(time.Second)
	wg.Wait()
}
