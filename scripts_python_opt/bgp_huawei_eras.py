import mysql.connector
import subprocess
import sys
import os

def check_whois_availability():
    # Definindo o comando de verificação baseado no sistema operacional
    check_command = "where" if os.name == 'nt' else "which"
    
    try:
        subprocess.run([check_command, "whois"], check=True, stdout=subprocess.DEVNULL)
        print("O pacote 'whois' está instalado.")
    except subprocess.CalledProcessError:
        print("O pacote 'whois' não está instalado.")
        if os.name == 'nt':
            print("Por favor, instale o 'whois' manualmente no Windows.")
        else:
            print("Considere instalar o 'whois' usando o gerenciador de pacotes do seu sistema (e.g., apt, yum).")
            sys.exit(1)

# Verifica a disponibilidade do whois
check_whois_availability()

# Configurações do banco de dados de origem
source_db_config = {
    'host': '45.235.144.28',
    'user': 'eraszabbix',
    'password': 'S@imon09xHGHS@imon@!!9is',
    'database': 'zabbix',
    'port': 3306
}

# Configurações do banco de dados de destino
destination_db_config = {
    'host': '45.235.144.28',
    'user': 'bgphuawei',
    'password': 'S@imon09xHGHS@imon@!!9is',
    'port': 3306
}

try:
    # Conectar ao MySQL Server sem especificar um banco de dados
    conn = mysql.connector.connect(host=destination_db_config['host'],
                                   user=destination_db_config['user'],
                                   password=destination_db_config['password'],
                                   port=destination_db_config['port'])
    cursor = conn.cursor()
    cursor.execute("CREATE DATABASE IF NOT EXISTS bgphuawei")
    cursor.execute("USE bgphuawei")
    
    # Criar tabelas se não existirem
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS bgp_name (
        asn VARCHAR(255),
        host_asn VARCHAR(255),
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    """)
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS bgp_infor (
        host_name VARCHAR(255),
        peer VARCHAR(255),
        ip VARCHAR(255),
        established_time INT,
        operational_status INT,
        total_routes INT,
        latest_timestamp TIMESTAMP
    );
    """)
    cursor.close()
    conn.close()
except mysql.connector.Error as err:
    print(f"Error: {err}")
    sys.exit(1)

# Reconfigurar para incluir o banco de dados no destino
destination_db_config['database'] = 'bgphuawei'

# Conectar ao banco de dados de origem
source_conn = mysql.connector.connect(**source_db_config)
source_cursor = source_conn.cursor()

# Conectar ao banco de dados de destino
destination_conn = mysql.connector.connect(**destination_db_config)
destination_cursor = destination_conn.cursor()

# Consulta SQL 1 e operações de inserção para bgp_infor
query1 = """
    SELECT
        host_name,
        REPLACE(SUBSTRING_INDEX(item_name, ' - ', -1), 'Total routes from peer ', '') AS peer,
        SUBSTRING_INDEX(SUBSTRING_INDEX(item_name, 'peer ', -1), ' - ', 1) AS ip,
        MAX(CASE WHEN item_name LIKE 'Established time%' THEN item_value END) AS established_time,
        MAX(CASE WHEN item_name LIKE 'Operational status%' THEN item_value END) AS operational_status,
        MAX(CASE WHEN item_name LIKE 'Total routes%' THEN item_value END) AS total_routes,
        MAX(timestamp) AS latest_timestamp
    FROM (
        SELECT
            h.host AS host_name,
            i.name AS item_name,
            hu.value AS item_value,
            CONVERT_TZ(FROM_UNIXTIME(hu.clock), '+00:00', '+03:00') AS timestamp
        FROM
            history_uint AS hu
        JOIN
            items AS i ON hu.itemid = i.itemid
        JOIN
            hosts AS h ON i.hostid = h.hostid
        WHERE
            h.host = 'BGP'
            AND i.name LIKE '%peer%'
        ORDER BY
            hu.clock DESC
        LIMIT 100
    ) AS subquery
    GROUP BY
        host_name, peer, ip;
"""
source_cursor.execute(query1)
result1 = source_cursor.fetchall()

destination_cursor.execute("TRUNCATE TABLE bgp_infor;")
for row in result1:
    insert_query = """
        INSERT INTO bgp_infor (host_name, peer, ip, established_time, operational_status, total_routes, latest_timestamp)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
    """
    destination_cursor.execute(insert_query, row)

# Consulta SQL 2 e operações de inserção para bgp_name
query2 = """
    SELECT 
        h.hostid,
        h.name AS host_name,
        ht.value AS host_asn,
        SUBSTRING_INDEX(SUBSTRING_INDEX(i.name, 'AS Name for IPv4 peer ', -1), ' - ', 1) AS ip_number,
        SUBSTRING_INDEX(SUBSTRING_INDEX(i.name, 'AS Name for IPv4 peer ', -1), ' - ', -1) AS asn,
        NOW() AS consulta_timestamp
    FROM 
        history_text ht
    JOIN 
        items i ON ht.itemid = i.itemid
    JOIN 
        hosts h ON i.hostid = h.hostid
    JOIN (
        SELECT itemid, MAX(clock) AS max_clock
        FROM history_text
        GROUP BY itemid
    ) max_clocks ON ht.itemid = max_clocks.itemid AND ht.clock = max_clocks.max_clock
    WHERE 
        h.host = 'BGP'
        AND i.name LIKE '%peer%'
    ORDER BY 
        ht.clock DESC;
"""
source_cursor.execute(query2)
result2 = source_cursor.fetchall()

destination_cursor.execute("TRUNCATE TABLE bgp_name;")
for row in result2:
    insert_query = """
        INSERT INTO bgp_name (host_asn, asn, timestamp)
        VALUES (%s, %s, %s)
    """
    destination_cursor.execute(insert_query, (row[2], row[4], row[5]))

# Commit as transações no banco de dados de destino e fechar todas as conexões
destination_conn.commit()
source_cursor.close()
destination_cursor.close()
source_conn.close()
destination_conn.close()

print("Operações de banco de dados concluídas com sucesso.")
