package util

import (
	"strings"

	"github.com/mohitkumar/orchy/model"
	"github.com/oliveagle/jsonpath"
)

func ResolveInputParams(flowContext *model.FlowContext, inputParams map[string]any) map[string]any {
	flowData := flowContext.Data
	data := make(map[string]any)
	resolveParams(flowData, inputParams, data)
	return data
}

func resolveParams(flowData map[string]any, params map[string]any, output map[string]any) {
	for k, v := range params {
		switch v.(type) {
		case map[string]any:
			out := make(map[string]any)
			output[k] = out
			resolveParams(flowData, v.(map[string]any), out)
		case string:
			if strings.HasPrefix(v.(string), "$") {
				value, _ := jsonpath.JsonPathLookup(flowData, v.(string))
				output[k] = value
			} else {
				output[k] = v
			}
		case []any:
			l := v.([]any)
			output[k] = resolveList(flowData, l)
		default:
			output[k] = v
		}
	}
}

func resolveList(flowData map[string]any, list []any) []any {
	var output []any
	for _, v := range list {
		switch v.(type) {
		case map[string]any:
			out := make(map[string]any)
			output = append(output, out)
			resolveParams(flowData, v.(map[string]any), out)
		case string:
			if strings.HasPrefix(v.(string), "$") {
				value, _ := jsonpath.JsonPathLookup(flowData, v.(string))
				output = append(output, value)
			} else {
				output = append(output, v)
			}
		case []any:
			l := v.([]any)
			outList := resolveList(flowData, l)
			output = append(output, outList...)
		default:
			output = append(output, v)
		}
	}
	return output
}
