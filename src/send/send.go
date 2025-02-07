package send

import (
	"fmt"
	"time"
	"encoding/json"
	s "strings"

	cf "hc/src/config"

	"github.com/google/uuid"
)

func SendAlimtalk(phn string){
	var startNow = time.Now()
	var serial_number = fmt.Sprintf("%04d%02d%02d-", startNow.Year(), startNow.Month(), startNow.Day())

	var alimtalk Alimtalk
	var attache AttachmentB
	// var attache2 AttachmentC
	// var link *Link
	// var supplement Supplement
	var button []Button
	// var quickreply []Quickreply
	// result := map[string]string{}

	alimtalk.Serial_number = serial_number + uuid.New().String()[:10] + "hc_" + phn

	alimtalk.Message_type = "AT"

	alimtalk.Sender_key = cf.Conf.PROFILE_KEY

	var cPhn string
	if s.HasPrefix(phn, "0"){
		cPhn = s.Replace(phn, "0", "82", 1)
	} else {
		cPhn = phn
	}
	alimtalk.Phone_number = cPhn

	alimtalk.Template_code = "web_monitor_01"

	alimtalk.Message = `웹 대기 발생
플랫폼 : 건강보험심사평가원
매장명 : 없음
발송시각 : 없음
발송종류 : 없음
발송번호 : 없음
전체/대기 : 에이전트 로그 확인 요함
확인시각 : 없음`
	var btn Button
	json.Unmarshal([]byte("{\"type\":\"WL\",\"name\":\"사이트 바로가기\",\"url_pc\":\"http://naver.com\",\"url_mobile\":\"http://naver.com\"}"), &btn)
	button = append(button, btn)

	alimtalk.Response_method = "realtime"

	attache.Buttons = button

	alimtalk.Attachment = attache

	sendKakaoAlimtalk(alimtalk)
}

func sendKakaoAlimtalk(alimtalk Alimtalk) {

	resp, err := cf.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetBody(alimtalk).
		Post(cf.Conf.API_SERVER + "/v3/" + cf.Conf.PARTNER_KEY + "/alimtalk/send")

	if err != nil {
		cf.Stdlog.Println("alimtalk server request error : ", err, " / serial_number : ", alimtalk.Serial_number)
	} else {
		cf.Stdlog.Println(string(resp.StatusCode()))
		cf.Stdlog.Println(string(resp.Body()))
	}
}