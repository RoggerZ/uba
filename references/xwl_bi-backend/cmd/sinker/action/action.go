package action

import "sync"

var (
	MetaEventMap        sync.Map
	AttributeMap        sync.Map
	MetaAttrRelationSet sync.Map
)

const (
	PresetAttribute  = 1
	CustomAttribute  = 2
	IsUserAttribute  = 1
	IsEventAttribute = 2
)
