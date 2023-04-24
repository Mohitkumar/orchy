package analytics

type DataCollectorConfig struct {
	FileName      string
	CollectorType DataCollectorType
}

type DataCollectorType string

const LOG_FILE_DATA_COLLECTOR DataCollectorType = "LOG_FILE_DATA_COLLECTOR"
const ELASTIC_DATA_COLLECTOR DataCollectorType = "ELASTIC_DATA_COLLECTOR"

type WorkflowDataCollector interface {
	RecordActionSuccess(wfName string, flowId string, actionName string, actionId int, data map[string]any)
	RecordActionFailure(wfName string, flowId string, actionName string, actionId int, reson string)
}

var worklfowCollector WorkflowDataCollector

func InitDataCollector(config DataCollectorConfig) error {
	switch config.CollectorType {
	case LOG_FILE_DATA_COLLECTOR:
		c, err := NewLogFileDataCollector(config.FileName)
		if err != nil {
			return err
		}
		worklfowCollector = c
	}
	return nil
}

func RecordActionSuccess(wfName string, flowId string, actionName string, actionId int, data map[string]any) {
	worklfowCollector.RecordActionSuccess(wfName, flowId, actionName, actionId, data)
}
func RecordActionFailure(wfName string, flowId string, actionName string, actionId int, reson string) {
	worklfowCollector.RecordActionFailure(wfName, flowId, actionName, actionId, reson)
}
