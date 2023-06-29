package models

import (
	"strconv"
	"time"

	"github.com/bytedance/sonic"
)

type ANY = interface{}

type RawData struct {
	Time      string `json:"datetime"`
	Value     string `json:"value"`
	Partition string `json:"partition"`
}

type Data struct {
	time      time.Time
	value     int
	partition string
}

func (data *Data) UnmarshalJSON(bytes []byte) error {
	var rawData RawData
	err := sonic.Unmarshal(bytes, &rawData)
	if err != nil {
		return err
	}

	time, err := time.Parse("2006-01-02 15:04:05.999999.999999", rawData.Time)
	if err != nil {
		return err
	}
	value, err := strconv.Atoi(rawData.Value)
	if err != nil {
		return err
	}
	data.time = time
	data.value = value
	data.partition = rawData.Partition
	return nil
}

func (data *Data) MarshalJSON() ([]byte, error) {
	rawData := RawData{
		Time:      data.time.Format("2006-01-02 15:04:05.999999.999999"),
		Value:     strconv.Itoa(data.value),
		Partition: data.partition,
	}
	return sonic.Marshal(&rawData)
}
