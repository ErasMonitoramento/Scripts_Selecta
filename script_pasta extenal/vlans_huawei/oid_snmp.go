package main

import (
    "bufio"
    "bytes"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"

    _ "github.com/go-sql-driver/mysql"
)

// Estrutura para armazenar cada entrada de descoberta
type DiscoveryItem struct {
    SNMPIndexVLA string `json:"{#SNMPINDEXVLA}"`
    IFAlias      string `json:"{#IFALIAS}"`
    VLAN         string `json:"{#VLAN}"`
}

// Função para criar o banco de dados e a tabela caso não existam
func setupDatabase() (*sql.DB, error) {
    // Conexão inicial ao MariaDB para criação do banco de dados
    dsn := "eraszabbix:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar ao MariaDB: %v", err)
    }

    // Criar o banco de dados se ele não existir
    _, err = db.Exec("CREATE DATABASE IF NOT EXISTS pppoene8000")
    if err != nil {
        return nil, fmt.Errorf("erro ao criar o banco de dados: %v", err)
    }

    // Conectar ao banco de dados `pppoene8000`
    db.Close()
    dsnWithDB := "eraszabbix:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/pppoene8000"
    db, err = sql.Open("mysql", dsnWithDB)
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar ao banco de dados pppoene8000: %v", err)
    }

    // Criar a tabela `snmp_data` se ela não existir
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS snmp_data (
        id INT AUTO_INCREMENT PRIMARY KEY,
        snmp_index_vla VARCHAR(50) NOT NULL,
        if_alias VARCHAR(100) NOT NULL,
        vlan VARCHAR(20) NOT NULL
    )`
    _, err = db.Exec(createTableQuery)
    if err != nil {
        return nil, fmt.Errorf("erro ao criar a tabela: %v", err)
    }

    return db, nil
}

// Função para salvar os dados no banco de dados MariaDB
func saveToDatabase(db *sql.DB, data []DiscoveryItem) error {
    // Inserindo dados na tabela snmp_data
    for _, item := range data {
        _, err := db.Exec("INSERT INTO snmp_data (snmp_index_vla, if_alias, vlan) VALUES (?, ?, ?)", item.SNMPIndexVLA, item.IFAlias, item.VLAN)
        if err != nil {
            return fmt.Errorf("erro ao inserir dados: %v", err)
        }
    }
    return nil
}

func main() {
    // Configurar o banco de dados e criar a tabela se necessário
    db, err := setupDatabase()
    if err != nil {
        log.Fatalf("Erro ao configurar o banco de dados: %v", err)
    }
    defer db.Close()

    // Definindo os valores diretamente no código
    comunidade := "public@dude"
    ip := "10.0.102.3"
    porta := "161"

    // Apenas os OIDs selecionados
    targetOIDs := map[string]bool{
        "115.0.1000": true,
        "117.0.1001": true,
        "123.0.3100": true,
        "155.0.2100": true,
        "172.0.100":  true,
        "202.0.2002": true,
        "203.0.2003": true,
        "205.0.230":  true,
        "206.0.250":  true,
        "208.0.550":  true,
        "214.0.200":  true,
        "215.0.261":  true,
        "216.0.262":  true,
    }

    descriptions, err := loadDescriptions("snmp_descriptions.txt")
    if err != nil {
        log.Fatalf(`{"data":[],"error":"Erro ao carregar as descrições: %v"}`, err)
    }

    command := "/usr/bin/snmpwalk"
    args := []string{"-v2c", "-c", comunidade, "-t", "30", "-r", "3", ip + ":" + porta, "1.3.6.1.4.1.2011.5.25.40.1.1.13.1.4"}

    cmd := exec.Command(command, args...)

    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    err = cmd.Run()
    if err != nil {
        log.Fatalf(`{"data":[],"error":"Erro ao executar o comando SNMP: %v, detalhes: %s"}`, err, stderr.String())
    }

    discoveryData := []DiscoveryItem{}
    scanner := bufio.NewScanner(&out)

    for scanner.Scan() {
        line := scanner.Text()

        parts := strings.Split(line, " = ")
        if len(parts) == 2 {
            oidPart := strings.Split(parts[0], ".")
            if len(oidPart) >= 4 {
                index := strings.Join(oidPart[len(oidPart)-3:], ".")
                if targetOIDs[index] {
                    description := descriptions[strings.Split(index, ".")[0]]
                    vlan := oidPart[len(oidPart)-1]
                    discoveryData = append(discoveryData, DiscoveryItem{
                        SNMPIndexVLA: index,
                        IFAlias:      fmt.Sprintf("\"%s\"", description),
                        VLAN:         vlan,
                    })
                }
            }
        }
    }

    if err := scanner.Err(); err != nil {
        log.Fatalf(`{"data":[],"error":"Erro ao processar a saída: %v"}`, err)
    }

    // Salvar os dados no banco de dados
    if err := saveToDatabase(db, discoveryData); err != nil {
        log.Fatalf("Erro ao salvar dados no banco de dados: %v", err)
    }

    jsonData, err := json.MarshalIndent(map[string]interface{}{"data": discoveryData}, "", "  ")
    if err != nil {
        log.Fatalf(`{"data":[],"error":"Erro ao converter para JSON: %v"}`, err)
    }

    fmt.Println(string(jsonData))
}

// Função para carregar as descrições de cada índice do arquivo
func loadDescriptions(filename string) (map[string]string, error) {
    descriptions := make(map[string]string)
    
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, " = ", 2)
        if len(parts) == 2 {
            index := strings.TrimSpace(parts[0])
            description := strings.TrimSpace(parts[1])
            descriptions[index] = description
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return descriptions, nil
}
