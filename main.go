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
	"time"
)

var EAI_SESS = os.Getenv("EAI_SESS")
var ROOM = os.Getenv("ROOM")

func execScript(name string, env []string) {
	cmd := exec.Cmd{
		Path: "/bin/sh",
		Args: []string{"/bin/sh", "/scripts/" + name + ".sh"},
		Env:  env,
	}
	output, err := cmd.Output()
	log.Println(string(output), err)
}

func getElectricBalance() (float64, error) {
	ids := strings.SplitN(ROOM, ",", 3)
	url := fmt.Sprintf("https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=%s&build_id=%s&room_id=%s", ids[0], ids[1], ids[2])
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

func main() {
	send := func(content string) {
		execScript("electric", []string{"CONTENT=" + content})
	}
	sentErr, sentEle := false, false
	for {
		electric, err := getElectricBalance()
		log.Println(electric, err)
		if err != nil {
			if !sentErr {
				sentErr = true
				send("获取电量出错了：" + err.Error())
			}
		} else {
			sentErr = false
			if electric <= 10 {
				if !sentEle {
					sentEle = true
					send("电量不足啦～")
				}
			} else {
				sentEle = false
			}
		}
		time.Sleep(time.Minute)
	}
}
