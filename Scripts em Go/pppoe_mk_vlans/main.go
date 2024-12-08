package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL
	"golang.org/x/crypto/ssh"
)

// PPPoEData representa os dados de uma entrada PPPoE
type PPPoEData struct {
	Name      string
	Interface string
	Uptime    string
	MTU       string
}

// Função para processar os dados recebidos
func parsePPPoEData(data string) []PPPoEData {
	var result []PPPoEData
	lines := strings.Split(data, "\n")
	regex := regexp.MustCompile(`(\w+)=(<[^\>]+>|[^\s]+)`)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		matches := regex.FindAllStringSubmatch(line, -1)
		entry := PPPoEData{}

		for _, match := range matches {
			key := match[1]
			value := strings.Trim(match[2], "<>")
			switch key {
			case "name":
				entry.Name = strings.TrimPrefix(value, "pppoe-")
			case "interface":
				entry.Interface = value
			case "uptime":
				entry.Uptime = value
			case "mtu":
				entry.MTU = value
			}
		}
		result = append(result, entry)
	}

	return result
}

// Função para gerar a lista de PPPoEs a partir do comando SSH
func generatePPPoEList(host, port, username, password, command string) ([]PPPoEData, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor SSH: %v", err)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("erro ao criar a sessão SSH: %v", err)
	}
	defer session.Close()

	output, err := session.Output(command)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar o comando no servidor SSH: %v", err)
	}

	return parsePPPoEData(string(output)), nil
}

// Função para configurar o banco de dados
func setupDatabase(db *sql.DB, dbName string) error {
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;", dbName))
	if err != nil {
		return fmt.Errorf("erro ao criar o banco de dados: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("USE %s;", dbName))
	if err != nil {
		return fmt.Errorf("erro ao selecionar o banco de dados: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pppoemk_online (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			vlan VARCHAR(255) NOT NULL,
			uptime VARCHAR(255) NOT NULL,
			mtu VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'offline',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar a tabela pppoemk_online: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pppoemk_offline (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			vlan VARCHAR(255) NOT NULL,
			uptime VARCHAR(255) NOT NULL,
			mtu VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar a tabela pppoemk_offline: %v", err)
	}

	return nil
}

// Função para atualizar listas online e offline
func updateLists(db *sql.DB, currentList []PPPoEData) error {
	currentMap := make(map[string]bool)
	for _, entry := range currentList {
		currentMap[entry.Name] = true
	}

	// Obter a lista online anterior
	rows, err := db.Query("SELECT name, vlan, uptime, mtu FROM pppoemk_online;")
	if err != nil {
		return fmt.Errorf("erro ao carregar a lista online: %v", err)
	}
	defer rows.Close()

	previousOnline := make(map[string]PPPoEData)
	for rows.Next() {
		var entry PPPoEData
		err := rows.Scan(&entry.Name, &entry.Interface, &entry.Uptime, &entry.MTU)
		if err != nil {
			return fmt.Errorf("erro ao escanear a lista online: %v", err)
		}
		previousOnline[entry.Name] = entry
	}

	// Remover logins que estão online da tabela offline
	for _, entry := range currentList {
		_, err := db.Exec("DELETE FROM pppoemk_offline WHERE name = ?;", entry.Name)
		if err != nil {
			return fmt.Errorf("erro ao remover login da tabela offline: %v", err)
		}
	}

	// Iniciar transação para atualizar tabelas
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %v", err)
	}

	// Atualizar tabela online com status
	_, err = tx.Exec("TRUNCATE TABLE pppoemk_online;")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("erro ao limpar tabela online: %v", err)
	}

	for _, entry := range currentList {
		_, err = tx.Exec("INSERT INTO pppoemk_online (name, vlan, uptime, mtu, status) VALUES (?, ?, ?, ?, 'online');",
			entry.Name, entry.Interface, entry.Uptime, entry.MTU)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("erro ao atualizar tabela online com status: %v", err)
		}
		delete(previousOnline, entry.Name)
	}

	// Adicionar logins offline
	for _, entry := range previousOnline {
		_, err = tx.Exec("INSERT INTO pppoemk_offline (name, vlan, uptime, mtu) VALUES (?, ?, ?, ?);",
			entry.Name, entry.Interface, entry.Uptime, entry.MTU)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("erro ao atualizar tabela offline: %v", err)
		}
	}

	// Finalizar transação
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("erro ao salvar alterações: %v", err)
	}

	return nil
}

// Função principal
func main() {
	host := "10.0.104.2"
	port := "8224"
	username := "Monitoramento-zabbix"
	password := "ZABBIX104070"
	command := "interface pppoe-server print terse"
	dsn := "pppoene8000:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/"
	dbName := "pppoene8000"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	err = setupDatabase(db, dbName)
	if err != nil {
		log.Fatalf("Erro ao configurar o banco de dados: %v", err)
	}

	currentList, err := generatePPPoEList(host, port, username, password, command)
	if err != nil {
		log.Fatalf("Erro ao gerar a lista atual: %v", err)
	}

	err = updateLists(db, currentList)
	if err != nil {
		log.Fatalf("Erro ao atualizar listas: %v", err)
	}

	fmt.Println("Processo concluído com sucesso!")
}
