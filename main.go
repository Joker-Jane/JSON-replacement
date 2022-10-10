package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Rule struct {
	Order       int    `json:"order"`
	Type        string `json:"type"`
	FieldName   string `json:"field-name"`
	Original    string `json:"original"`
	Replacement string `json:"replacement"`
}

func main() {
	input1 := `
{
  "@fields": {
    "Actor": [
      {
        "ID": "howard@fluencysecurity.com",
        "Type": 5
      }
    ],
    "CreationTime": "2022-08-29T20:57:06",
    "ModifiedProperties": [
      {
        "Name": "AssignedLicense",
        "NewValue": "[]",
        "OldValue": "[\r\n  \"[SkuName=O365_BUSINESS_PREMIUM, AccountId=11111, SkuId=f245ecc8-75af-4f8e-b61f-27d8114de5f3, DisabledPlans=[]]\"\r\n]"
      },
      {
        "Name": "Included Updated Properties",
        "NewValue": "AssignedLicense, AssignedPlan",
        "OldValue": ""
      }
    ],
    "ObjectId": "emily@fluencysecurity.com",
    "Operation": "Update user.",
    "Target": [
      {
        "ID": "User",
        "Type": 2
      },
      {
        "ID": "emily@fluencysecurity.com",
        "Type": 5
      }
    ],
    "TargetContextId": "22222",
    "UserId": "howard@fluencysecurity.com",
    "UserKey": "123456789@fluencysecurity.com"
  },
  "@source": "Audit.AzureActiveDirectory",
  "@timestamp": 1661806626000,
  "@sender": "office365",
  "@message": "Operation: Update user. associated with: howard@fluencysecurity.com",
  "@parser": "Office365Adjustments",
  "@parserVersion": "20210804-1",
  "@type": "event"
}
`

	input2 :=
		`
{
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

	input3 := `
{
  "@fields": {
    "Actor": [
      {
        "ID": "howard@fluencysecurity.com",
        "Type": 5
      }
    ]
  }
}
`

	rule := `
[
	{
		"order": 1,
		"type": "global",
		"original": "fluencysecurity",
		"replacement": "alphacorp"
	},
	{
		"order": 2,
		"type": "per-field",
		"field-name": "@fields.UserId",
		"original": "howard",
		"replacement": "bob"
	}
]
`

	input1 = input1
	input2 = input2
	input3 = input3
	rule = rule

	var m interface{}
	_ = json.Unmarshal([]byte(input1), &m)

	var rules []Rule
	_ = json.Unmarshal([]byte(rule), &rules)

	for _, r := range rules {
		var isGlobal bool
		if r.Type == "per-field" {
			isGlobal = false
		} else if r.Type == "global" {
			isGlobal = true
		} else {
			panic("Invalid type '" + r.Type + "'")
		}
		processCollection("", m, r.Original, r.Replacement, r.FieldName, isGlobal)
	}

	result, _ := json.Marshal(m)
	fmt.Println(string(result))
}

func processCollection(k string, v interface{}, from string, to string, field string, isGlobal bool) {
	switch v.(type) {
	case map[string]interface{}:
		processMap(v.(map[string]interface{}), from, to, field, isGlobal)
	case []interface{}:
		processArray(v.([]interface{}), k, from, to, field, isGlobal)
	}
}

func processMap(m map[string]interface{}, from string, to string, field string, isGlobal bool) {
	path := strings.SplitN(field, ".", 2)

	for k, v := range m {
		if isGlobal || k == path[0] {
			switch v.(type) {
			case string:
				m[k] = strings.Replace(v.(string), from, to, -1)
				//fmt.Println(field, k, m[k])
			default:
				if isGlobal {
					processCollection(k, v, from, to, "", isGlobal)
				} else {
					processCollection(k, v, from, to, path[1], isGlobal)
				}
			}
		}
	}
}

func processArray(l []interface{}, k string, from string, to string, field string, isGlobal bool) {
	for i, v := range l {
		switch v.(type) {
		case string:
			l[i] = strings.Replace(v.(string), from, to, -1)
			//fmt.Println(field, k, l[i])
		default:
			processCollection(k, v, from, to, field, isGlobal)
		}
	}
}
