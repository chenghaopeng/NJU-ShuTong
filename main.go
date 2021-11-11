package main

import (
	"encoding/json"
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

	"github.com/robfig/cron/v3"
)

var EAI_SESS = os.Getenv("EAI_SESS")
var ROOM = strings.ReplaceAll(os.Getenv("ROOM"), " ", "")
var MOD_AUTH_CAS = os.Getenv("MOD_AUTH_CAS")

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

func doGet(url string, headers map[string]string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bodyByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyByte)
	return body, nil
}

func getElectricBalance(areaId string, buildId string, roomId string) (float64, error) {
	url := fmt.Sprintf("https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=%s&build_id=%s&room_id=%s", areaId, buildId, roomId)
	body, err := doGet(url, map[string]string{
		"Cookie": "eai-sess=" + EAI_SESS,
	})
	if err != nil {
		return 0, err
	}
	reg := regexp.MustCompile(`dianyue\w*:\w*"([\d\.]+)"`)
	result := reg.FindStringSubmatch(body)
	if len(result) != 2 {
		return 0, errors.New("EAI_SESS 失效了")
	}
	return strconv.ParseFloat(result[1], 64)
}

func getLatestHealthReport() (checked bool, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = errors.New(fmt.Sprint(err2))
		}
	}()
	body, err := doGet("http://ehallapp.nju.edu.cn/xgfw/sys/yqfxmrjkdkappnju/apply/getApplyInfoList.do", map[string]string{
		"Cookie": "MOD_AUTH_CAS=" + MOD_AUTH_CAS,
	})
	if err != nil {
		return false, errors.New("MOD_AUTH_CAS 失效了")
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return false, err
	}
	var report map[string]interface{} = result["data"].([]interface{})[0].(map[string]interface{})
	return report["TBZT"] == "1", nil
}

func taskElectricBalance() {
	wg.Add(1)
	defer wg.Done()
	ids := strings.SplitN(ROOM, ",", 3)
	if len(EAI_SESS) == 0 || len(ids) != 3 {
		sendMessage("寝室未配置或配置错误，或未配置 EAI_SESS，不开启寝室电量监测")
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

func taskHealthReport() {
	wg.Add(1)
	defer wg.Done()
	if len(MOD_AUTH_CAS) == 0 {
		sendMessage("未配置 MOD_AUTH_CAS，不开启每日健康打卡监测")
		return
	}
	sendMessage("书童开始为你监测每日健康打卡啦～")
	sentErr := false
	c := cron.New()
	c.AddFunc("0 10,22 * * *", func() {
		checked, err := getLatestHealthReport()
		log.Println("taskHealthReport", checked, err)
		if err != nil {
			if !sentErr {
				sentErr = true
				sendMessage("获取每日健康打卡出错了：" + err.Error())
			}
		} else {
			sentErr = false
			if !checked {
				sendMessage("记得今天的每日健康打卡喔～")
			}
		}
	})
	c.Start()
	defer c.Stop()
	select {}
}

func main() {
	go taskElectricBalance()
	go taskHealthReport()
	time.Sleep(time.Second)
	wg.Wait()
}
