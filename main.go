package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	input := `{
  "@message": "Provider \"Environment\" is Started. \r\n\r\nDetails: \r\n\tProviderName=Environment\r\n\tNewProviderState=Started\r\n\r\n\tSequenceNumber=5\r\n\r\n\tHostName=ConsoleHost\r\n\tHostVersion=5.1.17763.2268\r\n\tHostId=902fd6a9-91b6-433e-9990-9395e9104b91\r\n\tHostApplication=powershell.exe -WindowStyle Hidden -nop -c \r\n\tEngineVersion=\r\n\tRunspaceId=\r\n\tPipelineId=\r\n\tCommandName=\r\n\tCommandType=\r\n\tScriptName=\r\n\tCommandPath=\r\n\tCommandLine=",
  "@facility": "user",
  "@level": "notice",
  "@tags": [
    "INFO"
  ],
  "@source": "18.117.240.147",
  "@sender": "18.117.240.147",
  "@timestamp": 1664898244000,
  "@fields": {
    "EventTime": "2022-10-04 15:44:03",
    "Hostname": "DEMO-NXLOG2",
    "Keywords": 36028797018963970,
    "EventType": "INFO",
    "SeverityValue": 2,
    "Severity": "INFO",
    "EventID": 600,
    "SourceName": "PowerShell",
    "Task": 6,
    "RecordNumber": 791023,
    "ProcessID": 0,
    "ThreadID": 0,
    "Channel": "Windows PowerShell",
    "Category": "Provider Lifecycle",
    "Opcode": "Info",
    "EventReceivedTime": "2022-10-04 15:44:04",
    "SourceModuleName": "powershell_in",
    "SourceModuleType": "im_msvistalog"
  },
  "@eventType": "nxlogAD",
  "@parser": "NxLogJsonProcessor",
  "@parserVersion": "20210608-1g",
  "@type": "event"
}
`

	var root map[string]interface{}
	err := json.Unmarshal([]byte(input), &root)
	if err != nil {
		fmt.Println("Failed to read json file")
		return
	}

	action := func(root string, val interface{}) {
		fmt.Println(root+":", val)
	}

	process("", "", root, action)
}

func process(root string, key string, val interface{}, action func(root string, val interface{})) {
	if root != "" && key != "" {
		root += "."
	}
	root += key

	switch val.(type) {
	case string, bool, float64:
		action(root, val)
	case []interface{}:
		iterateSlice(root, val.([]interface{}), action)
	default:
		iterateMap(root, val.(map[string]interface{}), action)
	}
}

func iterateMap(root string, group map[string]interface{}, action func(root string, val interface{})) {
	for key, val := range group {
		process(root, key, val, action)
	}
}

func iterateSlice(root string, slice []interface{}, action func(root string, val interface{})) {
	for _, val := range slice {
		process(root, "", val, action)
	}
}
