package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Função para conectar ao banco de dados
func connectDB(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}

// Função para criar a tabela se não existir
func createTableIfNotExists(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS trafego_pppoe (
		cliente VARCHAR(255) PRIMARY KEY,
		download BIGINT,
		upload BIGINT,
		3horas BIGINT,
		7d BIGINT,
		15d BIGINT,
		24h BIGINT,
		30d BIGINT,
		latest_timestamp DATETIME
	);`
	_, err := db.Exec(createTableQuery)
	return err
}

// Função para executar a consulta SQL
func queryDB(db *sql.DB, query string) ([]map[string]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare slice para guardar os resultados
	results := make([]map[string]interface{}, 0)

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}

		// Cria um map para armazenar o resultado da linha atual
		result := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			result[colName] = *val
		}

		results = append(results, result)
	}
	return results, nil
}

// Função para inserir os dados no banco de destino
func insertDB(db *sql.DB, data []map[string]interface{}) error {
	insertQuery := `
	INSERT INTO trafego_pppoe (cliente, download, upload, 3horas, 7d, 15d, 24h, 30d, latest_timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
	download = VALUES(download),
	upload = VALUES(upload),
	3horas = VALUES(3horas),
	7d = VALUES(7d),
	15d = VALUES(15d),
	24h = VALUES(24h),
	30d = VALUES(30d),
	latest_timestamp = VALUES(latest_timestamp);
	`

	for _, row := range data {
		_, err := db.Exec(insertQuery,
			row["Cliente_name"],
			row["download_bps"],
			row["upload_bps"],
			row["horas3"],
			row["dias7"],
			row["dias15"],
			row["horas24"],
			row["dias30"],
			row["latest_timestamp"],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	// Configuração para conectar ao banco de dados do Zabbix (origem)
	zabbixDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		"eraszabbix",
		"S@imon09xHGHS@imon@!!9is",
		"10.255.0.10",
		"3306",
		"zabbix",
	)

	// Configuração para conectar ao banco de dados pppoe (destino)
	pppoeDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		"pppoene8000",
		"S@imon09xHGHS@imon@!!9is",
		"10.255.0.10",
		"3306",
		"pppoene8000",
	)

	// Conectando ao banco de dados de origem (Zabbix)
	dbZabbix, err := connectDB(zabbixDSN)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de origem (Zabbix): %v", err)
	}
	defer dbZabbix.Close()

	// Conectando ao banco de dados de destino (PPPoE)
	dbPPPoE, err := connectDB(pppoeDSN)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de destino (PPPoE): %v", err)
	}
	defer dbPPPoE.Close()

	// Criando a tabela trafego_pppoe se não existir
	err = createTableIfNotExists(dbPPPoE)
	if err != nil {
		log.Fatalf("Erro ao criar a tabela trafego_pppoe: %v", err)
	}

	consultaSQL := `
	SELECT
		h.host AS host_name,
		SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING_INDEX(i.name, ' -', 1), 'Cliente ', -1), ':', 1) AS Cliente_name,
		MAX(CASE WHEN i.name LIKE '%: Bits received' THEN hu.value END) AS download_bps,
		MAX(CASE WHEN i.name LIKE '%: Bits sent' THEN hu.value END) AS upload_bps,
		MAX(CASE WHEN i.name LIKE '%- 3 Horas' THEN hu.value END) AS horas3,
		MAX(CASE WHEN i.name LIKE '%- 7 Dias' THEN hu.value END) AS dias7,
		MAX(CASE WHEN i.name LIKE '%- 15 Dias' THEN hu.value END) AS dias15,
		MAX(CASE WHEN i.name LIKE '%- 24 Horas' THEN hu.value END) AS horas24,
		MAX(CASE WHEN i.name LIKE '%- 30 Dias' THEN hu.value END) AS dias30,
		MAX(CONVERT_TZ(FROM_UNIXTIME(hu.clock), '+00:00', '+03:00')) AS latest_timestamp
	FROM
		history_uint AS hu
	INNER JOIN (
		SELECT
			itemid,
			MAX(clock) AS max_clock
		FROM
			history_uint
		GROUP BY
			itemid
	) AS latest_hu ON hu.itemid = latest_hu.itemid AND hu.clock = latest_hu.max_clock
	JOIN
		items AS i ON hu.itemid = i.itemid
	JOIN
		hosts AS h ON i.hostid = h.hostid
	WHERE
		h.host = 'RTR-BNG-CBA-C01-01'
		AND i.name LIKE '%Cliente%'
	GROUP BY
		Cliente_name
	ORDER BY
		latest_timestamp DESC;    
	`

	// Executa a consulta no banco de origem (Zabbix)
	dadosOrigem, err := queryDB(dbZabbix, consultaSQL)
	if err != nil {
		log.Fatalf("Erro ao consultar o banco de origem: %v", err)
	}

	// Insere os dados no banco de destino (PPPoE)
	err = insertDB(dbPPPoE, dadosOrigem)
	if err != nil {
		log.Fatalf("Erro ao inserir no banco de destino: %v", err)
	}

	fmt.Println("Dados transferidos com sucesso!")
}
