package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL
	"golang.org/x/crypto/ssh"
)

// PPPoEData representa os dados de uma entrada PPPoE
type PPPoEData struct {
	Name    string
	Address string
}

// Função para conectar ao SSH
func connectSSH(host, port, username, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor SSH: %v", err)
	}

	return client, nil
}

// Função para executar um comando SSH
func executeSSHCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("erro ao criar sessão SSH: %v", err)
	}
	defer session.Close()

	output, err := session.Output(command)
	if err != nil {
		return "", fmt.Errorf("erro ao executar comando SSH: %v", err)
	}

	return string(output), nil
}

// Função para processar a saída e filtrar NAME e ADDRESS
func parseOutput(output string) []PPPoEData {
	scanner := bufio.NewScanner(strings.NewReader(output))
	results := []PPPoEData{}

	for scanner.Scan() {
		line := scanner.Text()
		// Ignorar cabeçalhos ou linhas vazias
		if strings.HasPrefix(line, "Flags") || strings.HasPrefix(line, "Columns") || strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Dividir a linha em campos
		fields := strings.Fields(line)

		// Garantir que há pelo menos 5 campos (ignorando o campo Flags)
		if len(fields) >= 5 {
			name := fields[2]                  // O terceiro campo é o NAME
			address := fields[len(fields)-1]  // O último campo é o ADDRESS
			results = append(results, PPPoEData{Name: name, Address: address})
		}
	}

	return results
}

// Função para configurar o banco de dados
func setupDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS pppoemk_ativos (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			address VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar a tabela pppoemk_ativos: %v", err)
	}
	return nil
}

// Função para salvar os dados na tabela
func saveToDatabase(db *sql.DB, data []PPPoEData) error {
	// Limpar tabela antes de salvar os novos dados
	_, err := db.Exec("TRUNCATE TABLE pppoemk_ativos;")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %v", err)
	}

	// Inserir os dados
	query := "INSERT INTO pppoemk_ativos (name, address, created_at) VALUES (?, ?, ?);"
	for _, entry := range data {
		_, err := db.Exec(query, entry.Name, entry.Address, time.Now())
		if err != nil {
			return fmt.Errorf("erro ao salvar no banco de dados: %v", err)
		}
	}
	return nil
}

func main() {
	// Configurações do SSH
	sshHost := "10.0.104.2"
	sshPort := "8224"
	sshUser := "Monitoramento-zabbix"
	sshPassword := "ZABBIX104070"

	// Configurações do Banco de Dados
	dbDSN := "pppoene8000:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/pppoene8000"

	// Comando a ser executado
	command := "ppp active print"

	// Conectar ao SSH
	sshClient, err := connectSSH(sshHost, sshPort, sshUser, sshPassword)
	if err != nil {
		log.Fatalf("Erro ao conectar ao SSH: %v", err)
	}
	defer sshClient.Close()
	fmt.Println("Conexão SSH estabelecida com sucesso!")

	// Executar o comando SSH
	output, err := executeSSHCommand(sshClient, command)
	if err != nil {
		log.Fatalf("Erro ao executar comando SSH: %v", err)
	}

	// Processar a saída do comando
	results := parseOutput(output)

	// Conectar ao Banco de Dados
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("Erro ao conectar ao Banco de Dados: %v", err)
	}
	defer db.Close()

	// Configurar a tabela
	err = setupDatabase(db)
	if err != nil {
		log.Fatalf("Erro ao configurar a tabela no banco de dados: %v", err)
	}

	// Salvar os dados no banco de dados
	err = saveToDatabase(db, results)
	if err != nil {
		log.Fatalf("Erro ao salvar os dados no banco de dados: %v", err)
	}

	fmt.Println("Dados salvos no banco de dados com sucesso!")
}
