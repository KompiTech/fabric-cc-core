package mocktimelog

import (
	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
)

// these are required for tests to pass

// AttachToMockIncident attaches new TimeLog to existing Incident
var AttachToMockIncident = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	uuid := postPatch.Mapa[konst.AssetIdKey].(string)
	incident, err := ctx.GetRegistry().GetAsset("mockincident", postPatch.Mapa["incident"].(string), false, true) //asset.Get(ctx, "mockincident", postPatch.Mapa["incident"].(string), false)
	if err != nil {
		return rmap.Rmap{}, err
	}
	// get existing timelogs from Incident
	timelogsI, timelogsExists := incident.Mapa["timelogs"]
	timelogs := []string{}
	if timelogsExists {
		for _, timelogI := range timelogsI.([]interface{}) {
			timelogs = append(timelogs, timelogI.(string))
		}
	}
	// append new timelog to end of array
	timelogs = append(timelogs, uuid)
	// update incident with new array
	patchMap := map[string]interface{}{
		"timelogs": timelogs,
	}

	err = incident.ApplyMergePatch(rmap.NewFromMap(patchMap))
	if err != nil {
		return rmap.Rmap{}, err
	}
	if err := ctx.GetRegistry().PutAsset(incident, false); err != nil {
		return rmap.Rmap{}, err
	}
	return postPatch, nil
}
