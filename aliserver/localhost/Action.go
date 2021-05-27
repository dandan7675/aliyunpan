package localhost

import (
	"aliserver/aliyun"
	"aliserver/download"
	"aliserver/utils"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

//ActionPing 返回ping
func ActionPing(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("x-md-by", "xiaobaiyang")
	w.WriteHeader(200)
	w.Write([]byte("success"))
}

//ActionURL 拉取url内容返回
func ActionURL(w http.ResponseWriter, req *http.Request) {

	jsonstr := ""
	defer func() {
		if err := recover(); err != nil {
			log.Println("ActionURLError ", " input=", jsonstr, " error=", err)
		}
	}()

	var jsonbs []byte
	if req.Body != nil {
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		decodeBytes, _ := base64.StdEncoding.DecodeString(string(bodyBytes))
		jsonstr = string(decodeBytes)

		url := gjson.Get(jsonstr, "url").String()
		method := gjson.Get(jsonstr, "method").String()
		header := gjson.Get(jsonstr, "header").String()
		postdata := gjson.Get(jsonstr, "postdata").String()
		println("postdata=", postdata)
		if strings.HasPrefix(strings.ToLower(url), "http") {
			jsonbs = []byte(HookURL(url, method, header, postdata, req.UserAgent()))
		} else {
			ishook, hookresult := HookAction(url, postdata)
			if ishook {

				if len(hookresult) > 40 {
					println(hookresult[0:40])
				} else {
					println(hookresult)
				}
				jsonbs = []byte(`{"code":200, "header": "", "body": ` + hookresult + `}`)
				//jsonbs = []byte(utils.ToSuccessJSON3("code", 200, "header", "", "body", hookresult))
			} else {
				jsonbs = []byte(`{"code":200, "header": "` + utils.ToJSONString(jsonstr) + `", "body": {"code":503,"message":"不支持的操作"}}`)
				//jsonbs = []byte(utils.ToSuccessJSON3("code", 200, "header", jsonstr, "body", utils.ToErrorMessageJSON("不支持的操作")))
			}
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-md-by", "xiaobaiyang")
	w.WriteHeader(200)
	if len(jsonbs) > 0 {
		w.Write(jsonbs)
	}
}

//HookURL 联网加载指定的url数据
func HookURL(url string, method string, header string, postdata string, useragent string) string {

	if useragent != "" && !strings.Contains(header, "User-Agent") {
		header = "User-Agent: " + useragent + "\n" + header
	}
	if strings.ToLower(method) == "get" {
		respcode, resphead, respbody := utils.GetHTTPString(url, header)
		if respcode != 200 {
			resphead = url + header
		}
		return utils.ToSuccessJSON3("code", respcode, "header", resphead, "body", respbody)
	}
	respcode, resphead, respbody := utils.PostHTTPString(url, header, postdata)
	if respcode != 200 {
		resphead = url + header + postdata
	}
	return utils.ToSuccessJSON3("code", respcode, "header", resphead, "body", respbody)
}

//HookAction 拦截自定义命令
func HookAction(url string, postdata string) (bool, string) {

	switch url {
	case "Ping":
		return true, utils.ToSuccessJSON("Ping", "success")
	case "ApiUserInfo":
		return true, aliyun.ApiUserInfo()
	case "ApiUserLogoff":
		return true, aliyun.ApiUserLogoff()
	case "ApiQrcodeGenerate":
		return true, aliyun.ApiQrcodeGenerate()
	case "ApiQrcodeQuery":
		return true, aliyun.ApiQrcodeQuery()
	case "ApiTokenRefresh":
		refreshToken := gjson.Get(postdata, "refreshToken").String()
		return true, aliyun.ApiTokenRefresh(refreshToken)
	case "ApiFileList":
		parentid := gjson.Get(postdata, "parentid").String()
		marker := gjson.Get(postdata, "marker").String()
		return true, HookActionFileList(parentid, marker)
	case "ApiCreatForder":
		parentid := gjson.Get(postdata, "parentid").String()
		name := gjson.Get(postdata, "name").String()
		return true, aliyun.ApiCreatForder(parentid, name)
	case "ApiRename":
		file_id := gjson.Get(postdata, "file_id").String()
		name := gjson.Get(postdata, "name").String()
		return true, aliyun.ApiRename(file_id, name)
	case "ApiMoveBatch": //移动文件功能还没做
		movetoid := gjson.Get(postdata, "movetoid").String()
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		return true, aliyun.ApiMoveBatch(movetoid, list)
	case "ApiTrashBatch":
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		return true, aliyun.ApiTrashBatch(list)
	case "ApiTrashDeleteBatch":
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		return true, aliyun.ApiTrashDeleteBatch(list)
	case "ApiTrashRestoreBatch":
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		return true, aliyun.ApiTrashRestoreBatch(list)
	case "ApiTrashFileList":
		marker := gjson.Get(postdata, "marker").String()
		return true, aliyun.ApiTrashFileList(marker)
	case "ApiFavorFileList":
		marker := gjson.Get(postdata, "marker").String()
		return true, aliyun.ApiFavorFileList(marker)
	case "ApiFavor":
		file_id := gjson.Get(postdata, "file_id").String()
		isfavor := gjson.Get(postdata, "isfavor").Bool()
		return true, aliyun.ApiFavor(file_id, isfavor)
	case "ApiFavorBatch":
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		isfavor := gjson.Get(postdata, "isfavor").Bool()
		return true, aliyun.ApiFavorBatch(list, isfavor)
	case "GoPlay":
		file_id := gjson.Get(postdata, "file_id").String()
		return true, aliyun.ApiPlay(file_id)
	case "GoDownFile":
		filelist := gjson.Get(postdata, "filelist").Array()
		list := make([]string, len(filelist))
		for i := 0; i < len(filelist); i++ {
			list[i] = filelist[i].String()
		}
		savepath := gjson.Get(postdata, "savepath").String() //DownSavePath+RootPath 下载保存位置+该文件相对root的路径
		parentid := gjson.Get(postdata, "parentid").String() //D:\Down\     +新建文件夹\002\    + filelist[name]
		return true, download.DownFileMutil(parentid, savepath, list)
	case "GoDowningList":
		return true, download.DowningList()
	case "GoDownedList":
		return true, download.DownedList()

	case "GoDowningStart":
		downid := gjson.Get(postdata, "downid").String()
		if downid == "all" {
			return true, download.DowningStartAll()
		}
		return true, download.DowningStart(downid)
	case "GoDowningStop":
		downid := gjson.Get(postdata, "downid").String()
		if downid == "all" {
			return true, download.DowningStopAll()
		}
		return true, download.DowningStop(downid)
	case "GoDowningDelete":
		downid := gjson.Get(postdata, "downid").String()
		if downid == "all" {
			return true, download.DowningDeleteAll()
		}
		return true, download.DowningDelete(downid)
	case "GoDownedDelete":
		downid := gjson.Get(postdata, "downid").String()
		if downid == "all" {
			return true, download.DownedDeleteAll()
		}
		return true, download.DownedDelete(downid)
	case "GoDowningForder":
		downid := gjson.Get(postdata, "downid").String()
		return true, download.DowningForder(downid)
	case "GoDownedForder":
		downid := gjson.Get(postdata, "downid").String()
		return true, download.DownedForder(downid)

	case "GoSetting":
		key := gjson.Get(postdata, "key").String()
		val := gjson.Get(postdata, "val").String()
		return true, GoSetting(key, val)
	}
	return false, ""
}

//HookActionFileList 特殊处理 1.识别回收站、收藏夹、保险箱、最近访问 2.出错自动重试3次
func HookActionFileList(parentid string, marker string) string {
	jsonstr := ""
	for i := 0; i < 3; i++ {
		if parentid == "trash" {
			jsonstr = aliyun.ApiTrashFileList(marker)
			if len(jsonstr) > 60 {
				break
			}
		} else if parentid == "favorite" {
			jsonstr = aliyun.ApiFavorFileList(marker)
			if len(jsonstr) > 60 {
				break
			}
		} else if parentid == "safebox" {
			jsonstr = `{"code":0,"message":"success","next_marker":"","items":[]}`
			break
		} else if parentid == "calendar" {
			jsonstr = `{"code":0,"message":"success","next_marker":"","items":[]}`
			break
		} else {
			jsonstr = aliyun.ApiFileList(parentid, marker)
			if len(jsonstr) > 60 {
				break
			}
		}
	}
	return jsonstr
}
