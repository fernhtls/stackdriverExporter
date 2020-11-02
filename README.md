# Stackdriver exporter

Command line application to extract stackdiver metrics and export then into different formats like json, prometheus, bigquery ....

### Notes

Depends on the following for connectivity:

```
export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/my-key.json"
```

```
	// Sets your Google Cloud Platform project ID.
	projectID := "YOUR_PROJECT_ID"
```

Or `gcloud auth application-default`.

### Still to come

More output formats like PROMETHEUS, import to BQ and so on

### Examples calling to get multiple metrics

```
go run main.go --project_id "deployments-metrics" \
  --metric_type "storage.googleapis.com/storage/total_bytes|*/10 * * * *" \
  --metric_type "storage.googleapis.com/storage/object_count|*/10 * * * *" \
  --output_type "json" \
  --output_path "/tmp"
```
