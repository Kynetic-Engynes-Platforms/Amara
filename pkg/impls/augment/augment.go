package augment

import (
	adminservices "github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/admin_services"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/collection"
	nodemanager "github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/node_manager"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types"
)

func AugmentedClient(client *types.Client, cfg types.Config) (*types.Client, error) {
	nm, err := nodemanager.NewNodeManager(cfg.Nodes, cfg.HealthWaitTime)
	if err != nil {
		return nil, err
	}

	client.Nodes = nm

	client.Collections = &collection.CollectionsService{Client: client}
	client.Aliases = &adminservices.AliasesService{Client: client}
	client.Keys = &adminservices.KeysService{Client: client}
	client.Analytics = &adminservices.AnalyticsService{Client: client}
	client.Operations = &adminservices.OperationsService{Client: client}

	return client, nil
}
