package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gosnmp/gosnmp"
)

func main() {
	// Verifica se os argumentos foram fornecidos
	if len(os.Args) < 4 {
		log.Fatal("Uso: ./pppoe_total <comunidade> <índice> <porta>")
	}

	// Obtém os argumentos
	community := os.Args[1]
	index := os.Args[2]
	portStr := os.Args[3]

	// Converte a porta para inteiro
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Porta inválida: %v", err)
	}

	// Configuração SNMP
	params := &gosnmp.GoSNMP{
		Target:    "10.0.102.3", // Substitua pelo endereço IP do dispositivo
		Port:      uint16(port),
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   5 * time.Second, // Aumenta o timeout para 5 segundos
		Retries:   5,               // Aumenta o número de tentativas para 5
	}
	err = params.Connect()
	if err != nil {
		log.Fatalf("Falha na conexão SNMP: %v", err)
	}
	defer params.Conn.Close()

	// Prefixo OID base
	oidPrefix := ".1.3.6.1.4.1.2011.5.25.40.1.1.13.1.4."
	oid := oidPrefix + index

	// Faz a consulta SNMP Get
	result, err := params.Get([]string{oid})
	if err != nil {
		log.Fatalf("Erro ao executar consulta SNMP: %v", err)
	}

	// Verifica e imprime o valor do índice
	for _, variable := range result.Variables {
		switch variable.Type {
		case gosnmp.Integer:
			// Apenas o valor é impresso
			fmt.Println(gosnmp.ToBigInt(variable.Value).Int64())
		default:
			fmt.Printf("Tipo de dado não esperado: %v\n", variable.Type)
		}
	}
}
