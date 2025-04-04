package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cast"
)

const (
	botToken        = ""
	chatID          = ""
	baseUrl         = "https://api.admin.u-code.io"
	logFunctionName = "ucode-template"
	IsHTTP          = true
	appId           = ""
)

/*
Answer below questions before starting the function.

When the function invoked?
 - table_slug -> AFTER | BEFORE | HTTP -> CREATE | UPDATE | MULTIPLE_UPDATE | DELETE | APPEND_MANY2MANY | DELETE_MANY2MANY
What does it do?
- Explain the purpose of the function.(O'zbekcha yozilsa ham bo'ladi.)
*/

// Request structures
type (
	NewRequestBody struct {
		RequestData HttpRequest `json:"request_data"`
		Auth        AuthData    `json:"auth"`
		Data        Data        `json:"data"`
	}

	HttpRequest struct {
		Method  string      `json:"method"`
		Path    string      `json:"path"`
		Headers http.Header `json:"headers"`
		Params  url.Values  `json:"params"`
		Body    []byte      `json:"body"`
	}

	AuthData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	Request struct {
		Data map[string]interface{} `json:"data"`
	}

	Data struct {
		AppId      string                 `json:"app_id"`
		Method     string                 `json:"method"`
		ObjectData map[string]interface{} `json:"object_data"`
		ObjectIds  []string               `json:"object_ids"`
		TableSlug  string                 `json:"table_slug"`
		UserId     string                 `json:"user_id"`
	}

	FunctionRequest struct {
		BaseUrl     string  `json:"base_url"`
		TableSlug   string  `json:"table_slug"`
		AppId       string  `json:"app_id"`
		Request     Request `json:"request"`
		DisableFaas bool    `json:"disable_faas"`
	}
)

// Response structures
type (
	Datas struct {
		Data struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		} `json:"data"`
	}

	ClientApiResponse struct {
		Data ClientApiData `json:"data"`
	}

	ClientApiData struct {
		Data ClientApiResp `json:"data"`
	}

	ClientApiResp struct {
		Response map[string]interface{} `json:"response"`
	}

	Response struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}

	GetListClientApiResponse struct {
		Data GetListClientApiData `json:"data"`
	}

	GetListClientApiData struct {
		Data GetListClientApiResp `json:"data"`
	}

	GetListClientApiResp struct {
		Response []map[string]interface{} `json:"response"`
	}

	ClientApiUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			TableSlug string                 `json:"table_slug"`
			Data      map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	ClientApiMultipleUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			Data struct {
				Objects []map[string]interface{} `json:"objects"`
			} `json:"data"`
		} `json:"data"`
	}

	ResponseStatus struct {
		Status string `json:"status"`
	}
)

type (
	Asserts struct {
		Request  NewRequestBody
		Response Response
	}

	FunctionAssert struct{}
)

func (f FunctionAssert) GetAsserts() []Asserts {
	return []Asserts{
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
				},
			},
			Response: Response{
				Status: "done",
			},
		},
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098"},
				},
			},
			Response: Response{Status: "error"},
		},
	}
}

func (f FunctionAssert) GetBenchmarkRequest() Asserts {
	return Asserts{
		Request: NewRequestBody{
			Data: Data{
				AppId:     appId,
				ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
			},
		},
		Response: Response{
			Status: "done",
		},
	}
}

func Send(text string) {
	client := &http.Client{}
	text = logFunctionName + " >>>>> " + time.Now().Format(time.RFC3339) + " >>>>> " + text
	var botUrl = fmt.Sprintf("https://api.telegram.org/bot"+botToken+"/sendMessage?chat_id="+chatID+"&text=%s", text)
	request, err := http.NewRequest("GET", botUrl, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	Send("data" + string(data))

	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}

func ConvertResponse(data []byte) (ResponseStatus, error) {
	response := ResponseStatus{}
	err := json.Unmarshal(data, &response)
	return response, err
}

