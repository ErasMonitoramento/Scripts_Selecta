package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"

    _ "github.com/go-sql-driver/mysql"
)

// Estrutura para armazenar cada entrada de descoberta
type DiscoveryItem struct {
    SNMPIndexVLA string `json:"{#SNMPINDEXVLA}"`
    IFAlias      string `json:"{#IFALIAS}"`
    VLAN         string `json:"{#VLAN}"`
}

// Função para buscar os dados da tabela snmp_data e formatá-los como JSON de descoberta
func fetchDiscoveryData(db *sql.DB) ([]DiscoveryItem, error) {
    query := "SELECT snmp_index_vla, if_alias, vlan FROM snmp_data"
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("erro ao consultar a tabela snmp_data: %v", err)
    }
    defer rows.Close()

    discoveryData := []DiscoveryItem{}

    for rows.Next() {
        var item DiscoveryItem
        err := rows.Scan(&item.SNMPIndexVLA, &item.IFAlias, &item.VLAN)
        if err != nil {
            return nil, fmt.Errorf("erro ao ler os dados da linha: %v", err)
        }
        discoveryData = append(discoveryData, item)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("erro ao iterar pelas linhas: %v", err)
    }

    return discoveryData, nil
}

func main() {
    // Verificação dos argumentos
    if len(os.Args) < 4 {
        log.Fatalf("Uso: %s <comunidade> <IP> <porta>", os.Args[0])
    }

    // Configuração de conexão com o banco de dados
    dsnWithDB := "eraszabbix:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/pppoene8000"
    db, err := sql.Open("mysql", dsnWithDB)
    if err != nil {
        log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
    }
    defer db.Close()

    // Buscar dados da tabela e formatá-los para descoberta
    discoveryData, err := fetchDiscoveryData(db)
    if err != nil {
        log.Fatalf("Erro ao buscar dados para descoberta: %v", err)
    }

    // Convertendo para JSON no formato de descoberta
    output := map[string]interface{}{"data": discoveryData}
    jsonData, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        log.Fatalf("Erro ao converter dados para JSON: %v", err)
    }

    // Exibir o JSON formatado
    fmt.Println(string(jsonData))
}
