package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "github.com/gosnmp/gosnmp"
    _ "github.com/go-sql-driver/mysql"
)

type UserInfo struct {
    Login          string
    IPAddress      string
    VLAN           string
    ConnectionTime string
    Status         string
}

func executeSnmpWalk(oid string) (map[string]string, error) {
    g := &gosnmp.GoSNMP{
        Target:    "10.0.102.3",
        Port:      161,
        Version:   gosnmp.Version2c,
        Community: "public@dude",
        Timeout:   time.Duration(10) * time.Second,
        Retries:   5,
    }
    err := g.Connect()
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar ao dispositivo SNMP: %v", err)
    }
    defer g.Conn.Close()

    result := make(map[string]string)
    oids, err := g.WalkAll(oid)
    if err != nil {
        return nil, fmt.Errorf("erro ao executar o SNMP Walk: %v", err)
    }

    for _, variable := range oids {
        oidParts := strings.Split(variable.Name, ".")
        key := oidParts[len(oidParts)-1]
        var value string
        switch v := variable.Value.(type) {
        case []byte:
            value = string(v)
        default:
            value = fmt.Sprintf("%v", v)
        }
        result[key] = value
    }

    return result, nil
}

func ensureDatabaseExists(dsn string, dbName string) error {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return fmt.Errorf("erro ao conectar ao MySQL: %v", err)
    }
    defer db.Close()

    query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_bin", dbName)
    _, err = db.Exec(query)
    if err != nil {
        return fmt.Errorf("erro ao criar banco de dados %s: %v", dbName, err)
    }

    log.Printf("Banco de dados %s garantido com sucesso.\n", dbName)
    return nil
}

func ensureTablesExist(db *sql.DB) error {
    onlineTableQuery := `
    CREATE TABLE IF NOT EXISTS pppoe_sessions (
        id INT AUTO_INCREMENT PRIMARY KEY,
        snmp_index INT NOT NULL,
        login VARCHAR(255) NOT NULL,
        ip_address VARCHAR(45) NOT NULL,
        vlan INT NOT NULL,
        connection_time BIGINT NOT NULL,
        status VARCHAR(10) NOT NULL DEFAULT 'online',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
    _, err := db.Exec(onlineTableQuery)
    if err != nil {
        return fmt.Errorf("erro ao criar tabela pppoe_sessions: %v", err)
    }
    log.Println("Tabela pppoe_sessions garantida com sucesso.")

    offlineHistoryTableQuery := `
    CREATE TABLE IF NOT EXISTS pppoe_offline_history (
        id INT AUTO_INCREMENT PRIMARY KEY,
        login VARCHAR(255),
        offline_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        vlan VARCHAR(20)
    )`
    _, err = db.Exec(offlineHistoryTableQuery)
    if err != nil {
        return fmt.Errorf("erro ao criar tabela pppoe_offline_history: %v", err)
    }
    log.Println("Tabela pppoe_offline_history garantida com sucesso.")

    return nil
}

func saveUserListToFile(filePath string, userList map[string]UserInfo) error {
    var data []string
    for index, user := range userList {
        line := fmt.Sprintf("%s,%s,%s,%s,%s", index, user.Login, user.IPAddress, user.VLAN, user.ConnectionTime)
        data = append(data, line)
    }
    return os.WriteFile(filePath, []byte(strings.Join(data, "\n")), 0644)
}

func loadOnlineUsers(db *sql.DB) (map[string]struct {
    Login     string
    CreatedAt time.Time
}, error) {
    rows, err := db.Query("SELECT snmp_index, login, created_at FROM pppoe_sessions WHERE status='online'")
    if err != nil {
        return nil, fmt.Errorf("erro ao buscar dados anteriores: %v", err)
    }
    defer rows.Close()

    onlineUsers := make(map[string]struct {
        Login     string
        CreatedAt time.Time
    })
    for rows.Next() {
        var index string
        var login string
        var createdAt time.Time
        err = rows.Scan(&index, &login, &createdAt)
        if err != nil {
            return nil, fmt.Errorf("erro ao ler dados anteriores: %v", err)
        }
        onlineUsers[index] = struct {
            Login     string
            CreatedAt time.Time
        }{Login: login, CreatedAt: createdAt}
    }
    return onlineUsers, nil
}

func loadAllUsers(db *sql.DB) (map[string]UserInfo, error) {
    rows, err := db.Query("SELECT snmp_index, login, ip_address, vlan, connection_time, status FROM pppoe_sessions")
    if err != nil {
        return nil, fmt.Errorf("erro ao buscar dados dos usuários: %v", err)
    }
    defer rows.Close()

    allUsers := make(map[string]UserInfo)
    for rows.Next() {
        var index string
        var user UserInfo
        err = rows.Scan(&index, &user.Login, &user.IPAddress, &user.VLAN, &user.ConnectionTime, &user.Status)
        if err != nil {
            return nil, fmt.Errorf("erro ao ler dados dos usuários: %v", err)
        }
        allUsers[index] = user
    }
    return allUsers, nil
}

func main() {
    filePath := "./pppoe_users.txt"
    log.Println("Iniciando nova consulta SNMP e atualização da lista de usuários...")

    logins, err := executeSnmpWalk("1.3.6.1.4.1.2011.5.2.1.15.1.3")
    if err != nil {
        log.Fatalf("Erro ao obter logins: %v", err)
    }
    ips, err := executeSnmpWalk("1.3.6.1.4.1.2011.5.2.1.15.1.15")
    if err != nil {
        log.Fatalf("Erro ao obter IPs: %v", err)
    }
    vlans, err := executeSnmpWalk("1.3.6.1.4.1.2011.5.2.1.15.1.11")
    if err != nil {
        log.Fatalf("Erro ao obter VLANs: %v", err)
    }
    connectionTimes, err := executeSnmpWalk("1.3.6.1.4.1.2011.5.2.1.16.1.18")
    if err != nil {
        log.Fatalf("Erro ao obter tempos de conexão: %v", err)
    }

    userList := make(map[string]UserInfo)
    for index, login := range logins {
        userList[index] = UserInfo{
            Login:          login,
            IPAddress:      ips[index],
            VLAN:           vlans[index],
            ConnectionTime: connectionTimes[index],
            Status:         "online",
        }
    }

    err = saveUserListToFile(filePath, userList)
    if err != nil {
        log.Fatalf("Erro ao salvar a lista de usuários no arquivo de texto: %v", err)
    }

    rootDSN := "pppoene8000:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/?parseTime=true"
    dbName := "pppoene8000"
    if err := ensureDatabaseExists(rootDSN, dbName); err != nil {
        log.Fatalf("Erro ao garantir a existência do banco de dados: %v", err)
    }

    dsn := fmt.Sprintf("pppoene8000:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/%s?parseTime=true", dbName)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
    }
    defer db.Close()

    if err := ensureTablesExist(db); err != nil {
        log.Fatalf("Erro ao garantir a existência das tabelas: %v", err)
    }

    onlineUsers, err := loadOnlineUsers(db)
    if err != nil {
        log.Fatalf("Erro ao carregar usuários online do banco de dados: %v", err)
    }

    allUsers, err := loadAllUsers(db)
    if err != nil {
        log.Fatalf("Erro ao carregar todos os usuários do banco de dados: %v", err)
    }

    _, err = db.Exec("DELETE FROM pppoe_sessions")
    if err != nil {
        log.Fatalf("Erro ao limpar a tabela pppoe_sessions: %v", err)
    }

    for index, user := range userList {
        if user.Status == "online" {
            connectionTime := user.ConnectionTime
            if connectionTime == "" {
                connectionTime = "0"
            }

            _, err := db.Exec(`
                INSERT INTO pppoe_sessions (snmp_index, login, ip_address, vlan, connection_time, status)
                VALUES (?, ?, ?, ?, ?, ?)`,
                index, user.Login, user.IPAddress, user.VLAN, connectionTime, user.Status)
            if err != nil {
                log.Printf("Erro ao inserir dados online no banco de dados (Index %s): %v", index, err)
            }
        }
    }

    for index, user := range onlineUsers {
        if _, exists := userList[index]; !exists {
            vlan := ""
            if offlineUser, found := allUsers[index]; found {
                vlan = offlineUser.VLAN
            }

            _, err = db.Exec(`
                INSERT INTO pppoe_offline_history (login, offline_start, vlan)
                VALUES (?, ?, ?)`,
                user.Login, user.CreatedAt, vlan)
            if err != nil {
                log.Printf("Erro ao registrar histórico de usuário offline (Login %s, Index %s): %v", user.Login, index, err)
            }
            log.Printf("Usuário %s (Index: %s) registrado no histórico de offline com VLAN %s.\n", user.Login, index, vlan)
        }
    }

    fmt.Println("Status dos usuários atualizado com sucesso.")
}
