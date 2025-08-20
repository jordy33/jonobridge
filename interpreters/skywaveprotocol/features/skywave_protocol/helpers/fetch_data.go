package helpers

func FetchInEventCodes(hexString string, mapValues map[string]any) any {
	var code string

	if len(hexString) >= 2 {
		code = hexString[0:2]
	} else if len(hexString) == 1 {
		code = hexString
	} else {

		return map[string]any{
			"code": -1,
			"name": "Code undefined: " + hexString,
		}
	}

	if event, exists := mapValues[code]; exists {
		return event
	} else {
		return map[string]any{
			"code": -1,
			"name": "Code undefined: " + hexString,
		}
	}
}

func FetchInAlarmTypes(hexString string, mapValues map[string]any) any {
	if event, exists := mapValues[hexString]; exists {
		return event
	} else {
		return nil
	}
}
