package prometheus

import (
	"fmt"
	"log"
)

type OutPutConfig struct {
	Logger     *log.Logger
}

func (p *OutPutConfig) GetTimeSeriesMetric() {
	fmt.Println("nothing yet")
}