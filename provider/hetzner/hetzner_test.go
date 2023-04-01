package hetzner

import (
	"context"
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

type HetznerAPIStub struct {
	Zones   []Zone
	Records []Record
}

func (h HetznerAPIStub) GetAllZones() (*AllZones, error) {
	return &AllZones{Zones: h.Zones}, nil
}

func (h HetznerAPIStub) GetAllRecords(params GetAllRecordsParams) (*AllRecords, error) {
	return &AllRecords{Records: h.Records}, nil
}

func (h HetznerAPIStub) DeleteRecord(id string) error {
	//TODO implement me
	panic("implement me")
}

func TestHetznerProvider_ApplyChanges(t *testing.T) {
	type fields struct {
		BaseProvider provider.BaseProvider
		client       hetznerAPI
	}
	type args struct {
		ctx     context.Context
		changes *plan.Changes
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			fields: fields{
				client: HetznerAPIStub{
					Records: []Record{
						{
							Name: "test",
							Type: "A",
							Id:   "id-to-delete",
						},
					},
					Zones: []Zone{
						{
							Name: "example.de",
						},
					},
				},
			},
			args: args{
				changes: &plan.Changes{
					Delete: []*endpoint.Endpoint{
						{
							RecordType: "A",
							DNSName:    "test.example.de",
						},
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HetznerProvider{
				BaseProvider: tt.fields.BaseProvider,
				client:       tt.fields.client,
			}
			if err := h.ApplyChanges(tt.args.ctx, tt.args.changes); (err != nil) != tt.wantErr {
				t.Errorf("ApplyChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
