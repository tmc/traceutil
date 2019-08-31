// Command
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	cloudtrace "google.golang.org/api/cloudtrace/v1"
)

func main() {
	flagDataset := flag.String("dataset", "cloud_trace_spans", "Name of dataset")
	flagTableName := flag.String("table", "cloud_trace_spans", "Name of table")
	flag.Parse()

	ctx := context.Background()

	traces, err := fetchTraces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := insertSpans(ctx, *flagDataset, *flagTableName, traces); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func fetchTraces(ctx context.Context) ([]*cloudtrace.Trace, error) {
	c, err := google.DefaultClient(ctx, cloudtrace.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	cloudtraceService, err := cloudtrace.New(c)
	if err != nil {
		log.Fatal(err)
	}

	traces := []*cloudtrace.Trace{}
	fmt.Println("fetching spans")
	req := cloudtraceService.Projects.Traces.List(os.Getenv("PROJECT_ID")).View("COMPLETE")
	if err := req.Pages(ctx, func(page *cloudtrace.ListTracesResponse) error {
		for _, trace := range page.Traces {
			// TODO: Change code below to process each `trace` resource:
			traces = append(traces, trace)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return traces, nil
}

func insertSpans(ctx context.Context, datasetName string, tableName string, traces []*cloudtrace.Trace) error {
	client, err := bigquery.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		return xerrors.Errorf("issue creating client: %w", err)
	}

	table := client.Dataset(datasetName).Table(tableName)
	/*
		if err := table.Create(ctx, &bigquery.TableMetadata{}); err != nil {
			return xerrors.Errorf("Failed to create dataset: %v", err)
		}
	*/

	fmt.Printf("inserting %v traces\n", len(traces))
	for _, trace := range traces {
		//j, _ := json.Marshal(trace)
		if err := table.Inserter().Put(ctx, trace); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func fileToReader(path string) (io.Reader, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}
