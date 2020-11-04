package jsonoutput

import (
	"os"
	"strconv"

	"testing"

	"github.com/fernhtls/stackdriverExporter/utils"

	"github.com/stretchr/testify/assert"
)

func TestBuildFileName(t *testing.T) {
	j := JSONOutput{
		OutputPath: "/tmp",
	}
	startTime, endTime, err := utils.GetStartAndEndTimeCronJobs("*/5 * * * *")
	if err != nil {
		t.Error("error on generating test file name")
	}
	fileName := j.buildFileName("storage.googleapis.com/storage/object_count", startTime, endTime)
	startTimeUnixString := strconv.FormatInt(startTime.AsTime().Unix(), 10)
	endTimeUnixString := strconv.FormatInt(endTime.AsTime().Unix(), 10)
	fileNameCompare := "/tmp/storage-googleapis-com-storage-object_count_" + startTimeUnixString + "_" + endTimeUnixString + ".json"
	t.Log("fileName: ", fileName)
	t.Log("fileNameCompare: ", fileNameCompare)
	assert.Equal(t, fileName, fileNameCompare)
}

func TestValidateOutputMethod(t *testing.T) {
	jExists := JSONOutput{
		OutputPath: "/tmp",
	}
	err := jExists.ValidateOutputPath()
	assert.NoError(t, err)
	// Path does not exist
	jDoesNotExist := JSONOutput{
		OutputPath: "/zzzz",
	}
	err = jDoesNotExist.ValidateOutputPath()
	assert.Error(t, err)
	// Passing a file as path
	f, err := os.Create("/tmp/test_file.txt")
	f.WriteString("nothing")
	jFileError := JSONOutput{
		OutputPath: "/tmp/test_file.txt",
	}
	f.Close()
	err = jFileError.ValidateOutputPath()
	assert.Error(t, err)
	err = os.Remove("/tmp/test_file.txt")
	if err != nil {
		t.Error(err)
	}
}
