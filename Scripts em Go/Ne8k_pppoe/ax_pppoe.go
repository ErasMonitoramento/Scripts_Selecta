package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
)

// Função para verificar e remover todos os usuários que voltaram a estar online
func removeAllOnlineUsersFromHistory(db *sql.DB) error {
    // Obter todos os logins da tabela pppoe_offline_history
    rows, err := db.Query("SELECT login FROM pppoe_offline_history")
    if err != nil {
        return fmt.Errorf("erro ao buscar logins na tabela de histórico: %v", err)
    }
    defer rows.Close()

    // Iterar sobre cada login e verificar se está online
    for rows.Next() {
        var login string
        if err := rows.Scan(&login); err != nil {
            return fmt.Errorf("erro ao ler login: %v", err)
        }

        // Verificar se o login está na tabela pppoe_sessions com status 'online'
        var isOnline bool
        query := "SELECT EXISTS(SELECT 1 FROM pppoe_sessions WHERE login = ? AND status = 'online')"
        err := db.QueryRow(query, login).Scan(&isOnline)
        if err != nil {
            return fmt.Errorf("erro ao verificar se o usuário está online: %v", err)
        }

        // Se o login estiver online, removê-lo da tabela de histórico
        if isOnline {
            deleteQuery := "DELETE FROM pppoe_offline_history WHERE login = ?"
            _, err := db.Exec(deleteQuery, login)
            if err != nil {
                return fmt.Errorf("erro ao excluir o usuário da tabela de histórico: %v", err)
            }
            log.Printf("Usuário %s removido da tabela pppoe_offline_history.\n", login)
        }
    }

    return nil
}

func main() {
    // Configuração de conexão com o banco de dados
    dsn := "pppoene8000:S@imon09xHGHS@imon@!!9is@tcp(10.255.0.10:3306)/pppoene8000?parseTime=true"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
    }
    defer db.Close()

    // Executar a função para verificar e remover logins conforme necessário
    err = removeAllOnlineUsersFromHistory(db)
    if err != nil {
        log.Fatalf("Erro ao executar a verificação: %v", err)
    }
}
