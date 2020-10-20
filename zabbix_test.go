package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"testing"
)

type ZabbixRequestData struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

type ZabbixRequest struct {
	Request      string              `json:"request"`
	Data         []ZabbixRequestData `json:"data"`
	Clock        int                 `json:"clock"`
	Host         string              `json:"host"`
	HostMetadata string              `json:"host_metadata"`
}

func TestSendActiveMetric(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		conn, err := listener.Accept()
		if err != nil {
			errs <- err
		}

		// Obtain request from the mock zabbix server
		// Read protocol header and version
		header := make([]byte, 5)
		_, err = conn.Read(header)
		if err != nil {
			errs <- err
		}

		// Read data length
		dataLengthRaw := make([]byte, 8)
		_, err = conn.Read(dataLengthRaw)
		if err != nil {
			errs <- err
		}

		dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

		// Read data content
		content := make([]byte, dataLength)
		_, err = conn.Read(content)
		if err != nil {
			errs <- err
		}

		// The zabbix output checks that there are not errors
		// Zabbix header length not used, set to 1
		resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"info\":\"processed: 1; failed: 0; total: 1; seconds spent: 0.000030\"}")
		_, err = conn.Write(resp)
		if err != nil {
			errs <- err
		}

		// Close connection after reading the client data
		conn.Close()

		// Strip zabbix header and get JSON request
		var request ZabbixRequest
		err = json.Unmarshal(content, &request)
		if err != nil {
			errs <- err
		}

		expectedRequest := "agent data"
		if expectedRequest != request.Request {
			errs <- fmt.Errorf("Incorrect request field received, expected '%s'", expectedRequest)
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	m := NewMetric("zabbixAgent1", "ping", "13", true)

	s := NewSender(zabbixHost, zabbixPort)
	resActive, errActive, resTrapper, errTrapper := s.SendMetrics([]*Metric{m})
	if errActive != nil {
		t.Fatalf("error sending active metric: %v", errActive)
	}
	if errTrapper != nil {
		t.Fatalf("trapper error should be nil, we are not sending trapper metrics: %v", errTrapper)
	}

	raInfo, err := resActive.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if raInfo.Failed != 0 {
		t.Errorf("Failed error expected 0 got %d", raInfo.Failed)
	}
	if raInfo.Processed != 1 {
		t.Errorf("Processed error expected 1 got %d", raInfo.Processed)
	}
	if raInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", raInfo.Total)
	}

	_, err = resTrapper.GetInfo()
	if err == nil {
		t.Fatalf("No response trapper expected: %v", err)
	}

	// Wait for zabbix server emulator to finish
	err = <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}

func TestSendTrapperMetric(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		conn, err := listener.Accept()
		if err != nil {
			errs <- err
		}

		// Obtain request from the mock zabbix server
		// Read protocol header and version
		header := make([]byte, 5)
		_, err = conn.Read(header)
		if err != nil {
			errs <- err
		}

		// Read data length
		dataLengthRaw := make([]byte, 8)
		_, err = conn.Read(dataLengthRaw)
		if err != nil {
			errs <- err
		}

		dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

		// Read data content
		content := make([]byte, dataLength)
		_, err = conn.Read(content)
		if err != nil {
			errs <- err
		}

		// The zabbix output checks that there are not errors
		resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"info\":\"processed: 1; failed: 0; total: 1; seconds spent: 0.000030\"}")
		_, err = conn.Write(resp)
		if err != nil {
			errs <- err
		}

		// Close connection after reading the client data
		conn.Close()

		// Strip zabbix header and get JSON request
		var request ZabbixRequest
		err = json.Unmarshal(content, &request)
		if err != nil {
			errs <- err
		}

		expectedRequest := "sender data"
		if expectedRequest != request.Request {
			errs <- fmt.Errorf("Incorrect request field received, expected '%s'", expectedRequest)
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	m := NewMetric("zabbixAgent1", "ping", "13", false)

	s := NewSender(zabbixHost, zabbixPort)
	resActive, errActive, resTrapper, errTrapper := s.SendMetrics([]*Metric{m})
	if errTrapper != nil {
		t.Fatalf("error sending trapper metric: %v", errTrapper)
	}
	if errActive != nil {
		t.Fatalf("active error should be nil, we are not sending zabbix agent metrics: %v", errActive)
	}

	rtInfo, err := resTrapper.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if rtInfo.Failed != 0 {
		t.Errorf("Failed error expected 0 got %d", rtInfo.Failed)
	}
	if rtInfo.Processed != 1 {
		t.Errorf("Processed error expected 1 got %d", rtInfo.Processed)
	}
	if rtInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", rtInfo.Total)
	}

	_, err = resActive.GetInfo()
	if err == nil {
		t.Fatalf("No response active expected: %v", err)
	}

	// Wait for zabbix server emulator to finish
	err = <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}

func TestSendActiveAndTrapperMetric(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		for i := 0; i < 2; i++ {
			conn, err := listener.Accept()
			if err != nil {
				errs <- err
			}

			// Obtain request from the mock zabbix server
			// Read protocol header and version
			header := make([]byte, 5)
			_, err = conn.Read(header)
			if err != nil {
				errs <- err
			}

			// Read data length
			dataLengthRaw := make([]byte, 8)
			_, err = conn.Read(dataLengthRaw)
			if err != nil {
				errs <- err
			}

			dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

			// Read data content
			content := make([]byte, dataLength)
			_, err = conn.Read(content)
			if err != nil {
				errs <- err
			}

			// Strip zabbix header and get JSON request
			var request ZabbixRequest
			err = json.Unmarshal(content, &request)
			if err != nil {
				errs <- err
			}

			resp := []byte("")

			if request.Request == "sender data" {
				resp = []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"info\":\"processed: 1; failed: 0; total: 1; seconds spent: 0.000030\"}")
			} else if request.Request == "agent data" {
				resp = []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"info\":\"processed: 1; failed: 0; total: 1; seconds spent: 0.111111\"}")
			}

			// The zabbix output checks that there are not errors
			_, err = conn.Write(resp)
			if err != nil {
				errs <- err
			}

			// Close connection after reading the client data
			conn.Close()
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	m := NewMetric("zabbixAgent1", "ping", "13", true)
	m2 := NewMetric("zabbixTrapper1", "pong", "13", false)

	s := NewSender(zabbixHost, zabbixPort)
	resActive, errActive, resTrapper, errTrapper := s.SendMetrics([]*Metric{m, m2})

	if errActive != nil {
		t.Fatalf("error sending active metric: %v", errActive)
	}
	if errTrapper != nil {
		t.Fatalf("error sending trapper metric: %v", errTrapper)
	}

	raInfo, err := resActive.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if raInfo.Failed != 0 {
		t.Errorf("Failed error expected 0 got %d", raInfo.Failed)
	}
	if raInfo.Processed != 1 {
		t.Errorf("Processed error expected 1 got %d", raInfo.Processed)
	}
	if raInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", raInfo.Total)
	}

	rtInfo, err := resTrapper.GetInfo()
	if err != nil {
		t.Fatalf("error in response Trapper: %v", err)
	}

	if rtInfo.Failed != 0 {
		t.Errorf("Failed error expected 0 got %d", rtInfo.Failed)
	}
	if rtInfo.Processed != 1 {
		t.Errorf("Processed error expected 1 got %d", rtInfo.Processed)
	}
	if rtInfo.Total != 1 {
		t.Errorf("Total error expected 1 got %d", rtInfo.Total)
	}

	// Wait for zabbix server emulator to finish
	err = <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}

func TestRegisterHostOK(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		for i := 0; i < 2; i++ {
			conn, err := listener.Accept()
			if err != nil {
				errs <- err
			}

			// Obtain request from the mock zabbix server
			// Read protocol header and version
			header := make([]byte, 5)
			_, err = conn.Read(header)
			if err != nil {
				errs <- err
			}

			// Read data length
			dataLengthRaw := make([]byte, 8)
			_, err = conn.Read(dataLengthRaw)
			if err != nil {
				errs <- err
			}

			dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

			// Read data content
			content := make([]byte, dataLength)
			_, err = conn.Read(content)
			if err != nil {
				errs <- err
			}

			// Strip zabbix header and get JSON request
			var request ZabbixRequest
			err = json.Unmarshal(content, &request)
			if err != nil {
				errs <- err
			}

			expectedRequest := "active checks"
			if expectedRequest != request.Request {
				errs <- fmt.Errorf("Incorrect request field received, expected '%s'", expectedRequest)
			}

			// If the host does not exist, the first response will be an error
			resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"failed\",\"info\": \"host [prueba] not found\"}")

			// Next response is the valid one
			if i == 1 {
				resp = []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"data\": [{\"key\":\"net.if.in[eth0]\",\"delay\":60,\"lastlogsize\":0,\"mtime\":0}]}")
			}

			_, err = conn.Write(resp)
			if err != nil {
				errs <- err
			}

			// Close connection after reading the client data
			conn.Close()
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	s := NewSender(zabbixHost, zabbixPort)
	err := s.RegisterHost("prueba", "prueba")
	if err != nil {
		t.Fatalf("register host error: %v", err)
	}

	// Wait for zabbix server emulator to finish
	err = <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}

func TestRegisterHostError(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		for i := 0; i < 2; i++ {
			conn, err := listener.Accept()
			if err != nil {
				errs <- err
			}

			// Obtain request from the mock zabbix server
			// Read protocol header and version
			header := make([]byte, 5)
			_, err = conn.Read(header)
			if err != nil {
				errs <- err
			}

			// Read data length
			dataLengthRaw := make([]byte, 8)
			_, err = conn.Read(dataLengthRaw)
			if err != nil {
				errs <- err
			}

			dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

			// Read data content
			content := make([]byte, dataLength)
			_, err = conn.Read(content)
			if err != nil {
				errs <- err
			}

			// Strip zabbix header and get JSON request
			var request ZabbixRequest
			err = json.Unmarshal(content, &request)
			if err != nil {
				errs <- err
			}

			expectedRequest := "active checks"
			if expectedRequest != request.Request {
				errs <- fmt.Errorf("Incorrect request field received, expected '%s'", expectedRequest)
			}

			// Simulate error always
			resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"failed\",\"info\": \"host [prueba] not found\"}")
			_, err = conn.Write(resp)
			if err != nil {
				errs <- err
			}

			// Close connection after reading the client data
			conn.Close()
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	s := NewSender(zabbixHost, zabbixPort)
	err := s.RegisterHost("prueba", "prueba")
	if err == nil {
		t.Fatalf("should return an error: %v", err)
	}

	// Wait for zabbix server emulator to finish
	err = <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}

func TestInvalidResponseHeader(t *testing.T) {
	zabbixHost := "127.0.0.1"
	zabbixPort := 10051

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", zabbixHost, zabbixPort))
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer listener.Close()

	errs := make(chan error, 1)

	go func(chan error) {
		conn, err := listener.Accept()
		if err != nil {
			errs <- err
		}

		// Obtain request from the mock zabbix server
		// Read protocol header and version
		header := make([]byte, 5)
		_, err = conn.Read(header)
		if err != nil {
			errs <- err
		}

		// Read data length
		dataLengthRaw := make([]byte, 8)
		_, err = conn.Read(dataLengthRaw)
		if err != nil {
			errs <- err
		}

		dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

		// Read data content
		content := make([]byte, dataLength)
		_, err = conn.Read(content)
		if err != nil {
			errs <- err
		}

		// The zabbix output checks that there are not errors
		// Zabbix header length not used, set to 1
		resp := []byte("BXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\":\"success\",\"info\":\"processed: 1; failed: 0; total: 1; seconds spent: 0.000030\"}")
		_, err = conn.Write(resp)
		if err != nil {
			errs <- err
		}

		// Close connection after reading the client data
		conn.Close()

		// Strip zabbix header and get JSON request
		var request ZabbixRequest
		err = json.Unmarshal(content, &request)
		if err != nil {
			errs <- err
		}

		expectedRequest := "agent data"
		if expectedRequest != request.Request {
			errs <- fmt.Errorf("Incorrect request field received, expected '%s'", expectedRequest)
		}

		// End zabbix fake backend
		errs <- nil
	}(errs)

	m := NewMetric("zabbixAgent1", "ping", "13", true)

	s := NewSender(zabbixHost, zabbixPort)
	_, errActive, _, _ := s.SendMetrics([]*Metric{m})
	if errActive == nil {
		t.Fatal("Expected an error because an incorrect Zabbix protocol header")
	}

	// Wait for zabbix server emulator to finish
	err := <-errs
	if err != nil {
		t.Fatalf("Fake zabbix backend should not produce any errors: %v", err)
	}
}
