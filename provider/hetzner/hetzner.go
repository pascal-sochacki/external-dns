package hetzner

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

type HetznerProvider struct {
	provider.BaseProvider
	client hetznerAPI
}

type HetznerConfig struct {
	// Enabled dry-run will print any modifying actions rather than execute them.
	dryRun bool
	Token  string
}

func (h HetznerProvider) getAllZones(ctx context.Context) ([]Zone, error) {
	zones, err := h.client.GetAllZones()
	if err != nil {
		return nil, err
	}
	return zones.Zones, err
}

func (h HetznerProvider) getAllRecordsForZone(ctx context.Context, zone Zone) ([]Record, error) {
	result := make([]Record, 0)
	records, err := h.client.GetAllRecords(GetAllRecordsParams{
		zone: zone.Id,
	})
	if err != nil {
		return nil, err
	}

	for _, record := range records.Records {
		result = append(result, record)
	}

	return result, nil
}

func (h HetznerProvider) forEachRecordDo(ctx context.Context, action func(name string, record Record)) error {
	zones, err := h.getAllZones(ctx)
	if err != nil {
		return err
	}
	for _, zone := range zones {
		records, err := h.getAllRecordsForZone(ctx, zone)
		if err != nil {
			return err
		}

		for _, record := range records {
			if provider.SupportedRecordType(record.Type) {

				name := record.Name + "." + zone.Name

				if record.Name == "@" {
					name = zone.Name
				}

				action(name, record)
			}
		}

	}
	return nil
}

func (h HetznerProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	result := make([]*endpoint.Endpoint, 0)
	err := h.forEachRecordDo(ctx, func(name string, record Record) {
		result = append(result, endpoint.NewEndpointWithTTL(
			name,
			record.Type,
			endpoint.TTL(record.Ttl),
			record.Value),
		)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type Key struct {
	Name, Type string
}

func (h HetznerProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {

	dic := map[Key]Record{}
	err := h.forEachRecordDo(ctx, func(name string, record Record) {
		dic[Key{Type: record.Type, Name: name}] = record
	})
	if err != nil {
		return err
	}

	endpointsToDelete := append(changes.Delete, changes.UpdateOld...)
	if len(endpointsToDelete) == 0 {
		log.Debug("Nothing to endpoints")
	} else {
		err := h.deleteDelete(endpointsToDelete, dic)
		if err != nil {
			return err
		}
	}

	endpointsToCreate := append(changes.Delete, changes.UpdateNew...)
	if len(endpointsToCreate) == 0 {
		log.Debug("Nothing to create")
	} else {
		err := h.createEndpoints(endpointsToCreate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h HetznerProvider) deleteDelete(endpoint []*endpoint.Endpoint, dict map[Key]Record) error {
	for _, toDelete := range endpoint {

		record, ok := dict[Key{Type: toDelete.RecordType, Name: toDelete.DNSName}]
		if !ok {
			log.Error("could not find record to delete")
			return errors.New("could not find record to delete")
		}

		err := h.client.DeleteRecord(record.Id)
		if err != nil {
			log.Error("could not delete record")
			return err
		}

	}
	return nil
}

func (h HetznerProvider) createEndpoints(endpoint []*endpoint.Endpoint) error {
	for _, toCreate := range endpoint {
		err := h.client.CreateRecord(CreateRecordParams{
			Type: toCreate.RecordType,
			//TODO: find zone id
			ZoneId: "",
			Name:   toCreate.DNSName,
			//TODO: is this correct,
			Value: toCreate.Targets[0],
		})
		if err != nil {
			log.Error("could not delete record")
			return err
		}

	}
	return nil
}

func NewHetznerProvider(cfg HetznerConfig) (*HetznerProvider, error) {
	return &HetznerProvider{
		client: newHetznerClient(""),
	}, nil
}
