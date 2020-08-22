package jsonoutput

import (
	"strconv"

	"testing"

	"github.com/fernhtls/stackdriverExporter/utils"

	"github.com/stretchr/testify/assert"
)

func TestBuildFileName(t *testing.T) {
	startTime, endTime, err := utils.GetStartAndEndTimeJobs("*/5 * * * *")
	if err != nil {
		t.Error("error on generating test file name")
	}
	fileName := buildFileName("/tmp", "storage.googleapis.com/storage/object_count", startTime, endTime)
	startTimeUnixString := strconv.FormatInt(startTime.AsTime().Unix(), 10)
	endTimeUnixString := strconv.FormatInt(endTime.AsTime().Unix(), 10)
	fileNameCompare := "/tmp/storage-googleapis-com-storage-object_count_" + startTimeUnixString + "_" + endTimeUnixString + ".json"
	t.Log("fileName: ", fileName)
	t.Log("fileNameCompare: ", fileNameCompare)
	assert.Equal(t, fileName, fileNameCompare)
}
