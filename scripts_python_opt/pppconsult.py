import mysql.connector
from mysql.connector import Error
import os
from dotenv import load_dotenv

# Carrega as configurações do arquivo .env
load_dotenv()

# Configurações do Banco de Dados de Origem (Zabbix)
db_config_origem = {
    'host': os.getenv('ORIGEM_HOST'),
    'user': os.getenv('ORIGEM_USER'),
    'password': os.getenv('ORIGEM_PASSWORD'),
    'database': os.getenv('ORIGEM_DATABASE'),
    'port': int(os.getenv('ORIGEM_PORT'))
}

# Configurações do Banco de Dados de Destino
db_config_destino = {
    'host': os.getenv('DESTINO_HOST'),
    'user': os.getenv('DESTINO_USER'),
    'password': os.getenv('DESTINO_PASSWORD'),
    'database': os.getenv('DESTINO_DATABASE'),
    'port': int(os.getenv('DESTINO_PORT'))
}

# Consulta SQL para o Banco de Dados de Origem
# Nova consulta SQL para o Banco de Dados de Origem
consulta_sql = """
SELECT
    h.host AS host_name,
    SUBSTRING_INDEX(SUBSTRING_INDEX(i.name, '@', 1), 'Cliente ', -1) AS Cliente_name,
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
    h.host = 'VS-BRAS-PPPOE'
    AND i.name LIKE '%Cliente%'
GROUP BY
    Cliente_name
ORDER BY
    latest_timestamp DESC;    
"""


# Função para conectar ao banco de dados e executar a consulta
def consultar_db(config, consulta):
    try:
        conexao = mysql.connector.connect(**config)
        cursor = conexao.cursor(dictionary=True)
        cursor.execute(consulta)
        resultados = cursor.fetchall()
        cursor.close()
        conexao.close()
        return resultados
    except Error as e:
        print(f"Erro ao conectar ao MySQL: {e}")
        return None

# Função para inserir os dados no banco de destino
def inserir_db(config, dados):
    inserir_sql = """
    INSERT INTO clienteppp (cliente, download, upload, 3horas, 7d, 15d, 24h, 30d, latest_timestamp)
    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
    ON DUPLICATE KEY UPDATE
    download = VALUES(download),
    upload = VALUES(upload),
    3horas = VALUES(3horas),
    7d = VALUES(7d),
    15d = VALUES(15d),
    24h = VALUES(24h),
    30d = VALUES(30d),
    latest_timestamp = VALUES(latest_timestamp);
    """
    try:
        conexao = mysql.connector.connect(**config)
        cursor = conexao.cursor()
        for dado in dados:
            cursor.execute(inserir_sql, (
                dado['Cliente_name'],
                dado['download_bps'],  # Nome da coluna ajustado
                dado['upload_bps'],    # Nome da coluna ajustado
                dado['horas3'],        # Nome da coluna parece correto
                dado['dias7'],         # Nome da coluna parece correto
                dado['dias15'],        # Nome da coluna parece correto
                dado['horas24'],       # Nome da coluna parece correto
                dado['dias30'],        # Nome da coluna parece correto
                dado['latest_timestamp']
            ))
        conexao.commit()
        cursor.close()
        conexao.close()
    except Error as e:
        print(f"Erro ao inserir no MySQL: {e}")


# Executar a consulta e inserção
dados_origem = consultar_db(db_config_origem, consulta_sql)
if dados_origem:
    inserir_db(db_config_destino, dados_origem)
