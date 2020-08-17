/*
jingtao add
read log from pod directly
*/

package logdb

import (
	"encoding/json"
	"fmt"
	ferror "github.com/fission/fission/pkg/error"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

func NewPodDB(serverUrl string) (PodDB, error) {
	return PodDB{endpoint: serverUrl}, nil
}

type PodDB struct {
	endpoint string
}
type Logs struct {
	Pod       string `json:"pod"`
	logs      string `json:"logs"`
	timestamp string `json:"timestamp"`
}

func (pod PodDB) GetLogs(filter LogFilter) ([]LogEntry, error) {

	queryURL := fmt.Sprintf("%v/v2/functions/logs/%v?namespace=%v&stime=%v", pod.endpoint, filter.Function, filter.FunctionNamespace, filter.Since.UnixNano()+1)
	//fmt.Printf("query %v\nnano: %v\n", filter.Since.String(), filter.Since.UnixNano())

	req, err := http.NewRequest(http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request for pod log")
	}

	httpClient := http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ferror.MakeErrorFromHTTP(resp)
	}

	var logs map[string]string
	err = json.NewDecoder(resp.Body).Decode(&logs)
	if err != nil {
		return nil, errors.Wrap(err, "error decode request body")
	}

	var LogEntrys []LogEntry

	if status, ok := logs["status"]; ok {
		if status != "success" {
			fmt.Printf("msg: %v\n", logs["msg"])
			return LogEntrys, nil
		}
	}

	trimLogs := strings.TrimSpace(logs["logs"])
	if len(trimLogs) == 0 {
		return LogEntrys, nil
	}

	logLines := strings.Split(trimLogs, "\n")

	for _, line := range logLines { // 提取日志中的时间信息
		if len(line) == 0 {
			continue
		}
		words := strings.Split(line, " ")
		strTime := words[0]
		message := strings.Join(words[0:], " ")
		t, _ := time.Parse(time.RFC3339, strTime)
		if t.UnixNano() <= filter.Since.UnixNano() {
			// TODO the ans has the old logs
			continue
		}
		LogEntrys = append(LogEntrys, LogEntry{
			Timestamp: t,
			Message:   message,
			Pod:       logs["Pod"],
		})
		//fmt.Printf("log %v\nunix: %v\n", t.String(), t.UnixNano())
	}
	resp.Body.Close()
	return LogEntrys, nil
}
