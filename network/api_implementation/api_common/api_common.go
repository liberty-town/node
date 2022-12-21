package api_common

import (
	"liberty-town/node/pandora-pay/helpers/generics"
	"time"
)

type APICommon struct {
	temporaryList         *generics.Value[*APINetworkNodesReply]
	temporaryListCreation *generics.Value[time.Time]
}

func NewAPICommon() (api *APICommon, err error) {

	api = &APICommon{
		&generics.Value[*APINetworkNodesReply]{},
		&generics.Value[time.Time]{},
	}

	api.temporaryListCreation.Store(time.Now())

	return
}
