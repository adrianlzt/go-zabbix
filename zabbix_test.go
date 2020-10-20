package zabbix

import (
	"fmt"
	"testing"
)

func TestSendMetric(t *testing.T) {
	m := NewMetric("zabbixAgent1", "active", "13", true)
	m2 := NewMetric("zabbixAgent1", "trapper", "12", false)
	s := NewSender("10.20.20.179", 10051)
	resActive, errActive, resTrapper, errTrapper := s.SendMetrics([]*Metric{m, m2})
	if errActive != nil {
		t.Fatalf("error enviando la metrica Active: %v", errActive)
	}
	if errTrapper != nil {
		t.Fatalf("error enviando la metrica Trapper: %v", errTrapper)
	}
	t.Logf("ACTIVE: %+v\n", resActive)
	t.Logf("TRAPPER: %+v\n", resTrapper)

	raInfo, err := resActive.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if raInfo.Failed != 1 {
		t.Errorf("Failed error expected 1 got %d", raInfo.Failed)
	}
	if raInfo.Processed != 0 {
		t.Errorf("Processed error expected 0 got %d", raInfo.Processed)
	}
	if raInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", raInfo.Total)
	}

	rtInfo, err := resTrapper.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if rtInfo.Failed != 1 {
		t.Errorf("Failed error expected 1 got %d", rtInfo.Failed)
	}
	if rtInfo.Processed != 0 {
		t.Errorf("Processed error expected 0 got %d", rtInfo.Processed)
	}
	if rtInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", rtInfo.Total)
	}

}

func ExampleSendMetric() {
	m := NewMetric("zabbixAgent1", "active", "13", true)
	m2 := NewMetric("zabbixAgent1", "trapper", "12", false)
	s := NewSender("10.20.20.179", 10051)
	resActive, errActive, resTrapper, errTrapper := s.SendMetrics([]*Metric{m, m2})
	if errActive != nil {
		fmt.Errorf("error enviando la metrica Active: %v", errActive)
		return
	}
	if errTrapper != nil {
		fmt.Errorf("error enviando la metrica Trapper: %v", errTrapper)
		return
	}
	iA, err := resActive.GetInfo()
	if err != nil {
		fmt.Errorf("error en respuesta Active: %s", err)
		return
	}
	iT, err := resTrapper.GetInfo()
	if err != nil {
		fmt.Errorf("error en respuesta Trapper: %s", err)
		return
	}
	fmt.Printf("ACTIVE: %+v\n", iA.Failed)
	fmt.Printf("TRAPPER: %+v\n", iT.Failed)
	// Output:
	//ACTIVE: 1
	//TRAPPER: 1
}
