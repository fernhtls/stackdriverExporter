# Stackdriver exporter

Command line application to extract stackdiver metrics and export then into different formats like bigquery ....

### Notes

Depends on the following:

```
export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/my-key.json"
```

```
	// Sets your Google Cloud Platform project ID.
	projectID := "YOUR_PROJECT_ID"
```

### Still to come

More output formats like CSV, JSON, PROMETHEUS and so on

### Examples calling to get multiple metrics

```
go run main.go --project_id "deployments-metrics" \
  --metric_type "storage.googleapis.com/storage/total_bytes|*/10 * * * *" \
  --metric_type "storage.googleapis.com/storage/object_count|*/10 * * * *" \
  --output_type "json" \
  --output_path "/tmp"
```

### Next steps
* interface for function outputs
* unique jobs only