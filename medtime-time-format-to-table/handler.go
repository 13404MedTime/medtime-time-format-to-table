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
    IsHTTP          = true // if this is true banchmark test works.
    appId           = ""
)

/*
Answer below questions before starting the function.

When the function invoked?
 - table_slug -> AFTER | BEFORE | HTTP -> CREATE | UPDATE | MULTIPLE_UPDATE | DELETE | APPEND_MANY2MANY | DELETE_MANY2MANY
What does it do?
*/

// Request structures
type (
    // Handle request body
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

    // Function request body >>>>> GET_LIST, GET_LIST_SLIM, CREATE, UPDATE
    Request struct {
        Data map[string]interface{} `json:"data"`
    }

    // most common request structure -> UPDATE, MULTIPLE_UPDATE, CREATE, DELETE
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
    // Create function response body >>>>> CREATE
    Datas struct {
        Data struct {
            Data struct {
                Data map[string]interface{} `json:"data"`
            } `json:"data"`
        } `json:"data"`
    }

    // ClientApiResponse This is get single api response >>>>> GET_SINGLE_BY_ID, GET_SLIM_BY_ID
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

    // GetListClientApiResponse This is get list api response >>>>> GET_LIST, GET_LIST_SLIM
    GetListClientApiResponse struct {
        Data GetListClientApiData `json:"data"`
    }

    GetListClientApiData struct {
        Data GetListClientApiResp `json:"data"`
    }

    GetListClientApiResp struct {
        Response []map[string]interface{} `json:"response"`
    }

    // ClientApiUpdateResponse This is single update api response >>>>> UPDATE
    ClientApiUpdateResponse struct {
        Status      string `json:"status"`
        Description string `json:"description"`
        Data        struct {
            TableSlug string                 `json:"table_slug"`
            Data      map[string]interface{} `json:"data"`
        } `json:"data"`
    }

    // ClientApiMultipleUpdateResponse This is multiple update api response >>>>> MULTIPLE_UPDATE
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

// Testing types
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

// Handle a serverless request
func Handle(req []byte) string {
    var (
        response Response
        request  NewRequestBody
    )

    defer func() {
        responseByte, _ := json.Marshal(response)
        Send(string(responseByte))
    }()

    err := json.Unmarshal(req, &request)
    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while unmarshalling request", "error": err.Error()}
        response.Status = "error"
        responseByte, _ := json.Marshal(response)
        return string(responseByte)
    }

    inputTimeStr := cast.ToString(request.Data.ObjectData["date"])

    inputTime, err := time.Parse(time.RFC3339, inputTimeStr)
    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while parsing time", "error": err.Error()}
        response.Status = "error"
        responseByte, _ := json.Marshal(response)
        return string(responseByte)
    }

    outputFormat := "02.01.2006 15:04"
    outputTimeStr := inputTime.Format(outputFormat)

    updateObjectRequest := Request{
        Data: map[string]interface{}{
            "guid": cast.ToString(request.Data.ObjectIds[0]),
            "date": outputTimeStr,
        },
    }

    res, response, err := UpdateObject(FunctionRequest{
        BaseUrl:     baseUrl,
        TableSlug:   cast.ToString(request.Data.TableSlug),
        AppId:       request.Data.AppId,
        Request:     updateObjectRequest,
        DisableFaas: true,
    })

    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while unmarshalling request", "error": err.Error()}
        response.Status = "error"
        responseByte, _ := json.Marshal(response)
        return string(responseByte)
    }

    response.Data = res.Data.Data
    response.Data = map[string]interface{}{}
    response.Status = "done" //if all will be ok else "error"
    responseByte, _ := json.Marshal(response)

    return string(responseByte)
}


func UpdateObject(in FunctionRequest) (ClientApiUpdateResponse, Response, error) {
    response := Response{
        Status: "done",
    }

    var updateObject ClientApiUpdateResponse
    updateObjectResponseInByte, err := DoRequest(fmt.Sprintf("%s/v1/object/%s?from-ofs=%t", in.BaseUrl, in.TableSlug, in.DisableFaas), "PUT", in.Request, in.AppId)
    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while updating object", "error": err.Error()}
        response.Status = "error"
        return ClientApiUpdateResponse{}, response, errors.New("error")
    }

    err = json.Unmarshal(updateObjectResponseInByte, &updateObject)
    if err != nil {
        response.Data = map[string]interface{}{"message": "Error while unmarshalling update object", "error": err.Error()}
        response.Status = "error"
        return ClientApiUpdateResponse{}, response, errors.New("error")
    }

    return updateObject, response, nil
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

func ConvertResponse(data []byte) (ResponseStatus, error) {
    response := ResponseStatus{}

    err := json.Unmarshal(data, &response)

    return response, err
}

