package util

import "google.golang.org/protobuf/types/known/structpb"

func ConvertToProto(data map[string]any) map[string]*structpb.Value {
	out := make(map[string]*structpb.Value)
	for k, v := range data {
		if val, err := structpb.NewValue(v); err == nil {
			out[k] = val
		}
	}
	return out
}

func ConvertMapToStructPb(data map[string]any) *structpb.Value {
	if val, err := structpb.NewValue(data); err == nil {
		return val
	}
	return nil
}

func ConvertFromProto(data map[string]*structpb.Value) map[string]any {
	out := make(map[string]any)
	for k, v := range data {
		val := v.AsInterface()
		out[k] = val
	}
	return out
}
