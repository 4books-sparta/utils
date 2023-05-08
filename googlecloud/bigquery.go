package googlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
)

const BigQueryErrPrefix = "Google-Bigquery-"

type BigqueryClient struct {
	client  *bigquery.Client
	config  *BigQueryConfig
	Verbose bool
}

func NewBigqueryClient(cfg *BigQueryConfig) (*BigqueryClient, error) {
	ret := BigqueryClient{
		config: cfg,
	}
	bqClient, err := bigquery.NewClient(context.Background(), cfg.ProjectId)
	if err != nil {
		fmt.Println("Bigquery error", err)
		return nil, err
	}
	ret.client = bqClient

	return &ret, nil
}

type BigQuerySavePayload struct {
	Item map[string]bigquery.Value
}

func (s *BigQuerySavePayload) Save() (map[string]bigquery.Value, string, error) {
	return s.Item, bigquery.NoDedupeID, nil
}

type BigQueryConfig struct {
	Dataset        string
	ReplicaDataset string
	ProjectId      string
}

func GetBigQueryConfig(projectId, dataset, replicaDataset string) *BigQueryConfig {
	return &BigQueryConfig{
		ProjectId:      projectId,
		Dataset:        dataset,
		ReplicaDataset: replicaDataset,
	}
}

func (cl *BigqueryClient) Exec(ctx context.Context, sql string) (*bigquery.RowIterator, error) {
	if cl.client == nil {
		return nil, errors.New("bigquery-client-nil")
	}
	return cl.client.Query(sql).Read(ctx)
}

func (cl *BigqueryClient) SaveRecords(table string, payload []*BigQuerySavePayload) error {
	if cl.client == nil {
		return errors.New("bigquery-client-nil")
	}

	dataset := cl.config.Dataset

	ins := cl.client.Dataset(dataset).Table(table).Inserter()
	err := ins.Put(context.Background(), payload)
	if err != nil {
		if multiError, ok := err.(bigquery.PutMultiError); ok {
			for _, err1 := range multiError {
				for _, err2 := range err1.Errors {
					if cl.Verbose {
						fmt.Println(BigQueryErrPrefix, err2)
					}
				}
			}
		} else {
			fmt.Println(BigQueryErrPrefix, err)
		}
	}
	return err
}

func (cl *BigqueryClient) GetPayload(ev interface{}) (*BigQuerySavePayload, error) {
	if ev == nil {
		return nil, nil
	}
	payload := BigQuerySavePayload{}
	jsonString, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}

	item := make(map[string]bigquery.Value, 0)
	err = json.Unmarshal(jsonString, &item)
	if err != nil {
		return nil, err
	}
	payload.Item = item

	return &payload, nil
}

func (cl *BigqueryClient) GetCompleteTableName(tab string) string {
	return fmt.Sprintf("%s.%s.%s", cl.config.ProjectId, cl.config.Dataset, tab)
}
