package cache

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/server/model"
	c "github.com/patrickmn/go-cache"
)

type FlowStateCache struct {
	cache *c.Cache
}

func NewFlowStateCache() *FlowStateCache {
	return &FlowStateCache{
		cache: c.New(c.NoExpiration, 10*time.Minute),
	}
}

func (ch *FlowStateCache) SaveFlowState(flowId string, state model.FlowState) {
	ch.cache.Add(flowId, string(state), c.NoExpiration)
}

func (ch *FlowStateCache) GetFlowState(flowId string) (model.FlowState, bool) {
	stateStr, found := ch.cache.Get(flowId)
	if found {
		return model.FlowState(fmt.Sprintf("%v", stateStr)), true
	}
	return model.FlowState(""), false
}
