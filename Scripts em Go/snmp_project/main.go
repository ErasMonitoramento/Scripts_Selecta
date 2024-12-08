package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gosnmp/gosnmp"
)

func snmpGetSingle(oid, host, community string, port uint16) ([]float64, error) {
	gosnmp.Default.Target = host
	gosnmp.Default.Community = community
	gosnmp.Default.Port = port
	gosnmp.Default.Version = gosnmp.Version2c
	gosnmp.Default.Timeout = gosnmp.Default.Timeout

	err := gosnmp.Default.Connect()
	if err != nil {
		return nil, fmt.Errorf("error connecting to host: %v", err)
	}
	defer gosnmp.Default.Conn.Close()

	result, err := gosnmp.Default.Get([]string{oid})
	if err != nil {
		return nil, fmt.Errorf("error getting OID: %v", err)
	}

	var resultList []float64
	for _, variable := range result.Variables {
		switch variable.Type {
		case gosnmp.OctetString:
			valueStr := string(variable.Value.([]byte))
			valueFloat, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("error converting string to float: %v", err)
			}
			resultList = append(resultList, valueFloat)
		case gosnmp.Integer, gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32:
			valueFloat := float64(variable.Value.(int))
			resultList = append(resultList, valueFloat)
		default:
			return nil, fmt.Errorf("unsupported SNMP value type: %T", variable.Value)
		}
	}
	return resultList, nil
}

func main() {
	if len(os.Args) < 6 {
		fmt.Printf("Usage: %s <community> <host> <index> <port> <signal_type>\n", os.Args[0])
		os.Exit(1)
	}

	community := os.Args[1]
	host := os.Args[2]
	index := os.Args[3]
	portStr := os.Args[4]
	signalType := os.Args[5]

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	var oid string
	if signalType == "RX" {
		oid = fmt.Sprintf("1.3.6.1.4.1.2011.5.25.31.1.1.3.1.32.%s", index)
	} else if signalType == "TX" {
		oid = fmt.Sprintf("1.3.6.1.4.1.2011.5.25.31.1.1.3.1.33.%s", index)
	} else {
		log.Fatalf("Invalid signal type: %s", signalType)
	}

	data, err := snmpGetSingle(oid, host, community, uint16(port))
	if err != nil {
		log.Fatalf("Error retrieving SNMP data: %v", err)
	}

	for _, value := range data {
		adjustedValue := value * 0.01
		fmt.Printf("%.2f\n", adjustedValue)
	}
}
