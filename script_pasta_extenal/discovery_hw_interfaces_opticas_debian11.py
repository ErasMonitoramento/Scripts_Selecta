#!/usr/bin/env python3

# Desenvolvido por: Bee Solutions
# Autor: Fernando Almondes
# Data: 27/01/2024 - 22:13

import subprocess
import re
import json
import sys

# Verifica os argumentos de linha de comando
if len(sys.argv) < 4:
    print("Uso: python discovery_hw_interfaces_opticas.py <IP> <community> <porta>")
    sys.exit(1)

ip = sys.argv[1]
community = sys.argv[2]
porta = sys.argv[3]

# Função para remover aspas extras
def remove_extra_quotes(value):
    if isinstance(value, str):
        return value.strip('"')
    return value

# Executa o comando snmpwalk para obter a Lista 1 - IFNAME
cmd_lista1 = f"snmpwalk -On -Cc -v2c -c {community} {ip}:{porta} 1.3.6.1.2.1.31.1.1.1.1 | head -80 | grep -E '40GE|100GE|XG|Gi|25GE'"
output_lista1 = subprocess.check_output(cmd_lista1, shell=True).decode()

# Executa o comando snmpwalk para obter a Lista 2 - IFALIAS
cmd_lista2 = f"snmpwalk -On -Cc -v2c -c {community} {ip}:{porta} 1.3.6.1.2.1.31.1.1.1.18 | head -80"
output_lista2 = subprocess.check_output(cmd_lista2, shell=True).decode()

# Executa o comando snmpwalk para obter a Lista 3 - ENTPHYSICALNAME
cmd_lista3 = f"snmpwalk -On -Cc -v2c -c {community} {ip}:{porta} 1.3.6.1.2.1.47.1.1.1.1.7 | head -80 | grep -E '40GE|100GE|XG|Gi|25GE'"
output_lista3 = subprocess.check_output(cmd_lista3, shell=True).decode()

# Processa a Lista 1
matches1 = re.findall(r'\.1\.3\.6\.1\.2\.1\.31\.1\.1\.1\.1\.(\d+)\s=\sSTRING:\s*(.*)', output_lista1)
lista1 = [{"{#SNMPINDEXOLD}": index.strip(), "{#ENTPHYSICALNAME}": name.strip()} for index, name in matches1[:80] if name.strip()]

# Captura a lista de ifindex da lista 1 para comparacao
index_lista_1 = re.findall(r'\.1\.3\.6\.1\.2\.1\.31\.1\.1\.1\.1\.(\d+)', output_lista1)

# Processa a Lista 2
matches2 = re.sub(r'\"\"', 'STRING: \"\"', output_lista2)
matches2 = re.findall(r'\.1\.3\.6\.1\.2\.1\.31\.1\.1\.1\.18\.(\d+)\s=\sSTRING:\s*(.*)', matches2)
lista2_tmp = [{"{#SNMPINDEXOLD}": index.strip(), "{#IFALIAS}": name.strip()} for index, name in matches2[:80] if name.strip()]

# Cria uma nova lista para inserir somente as interfaces fisicas
lista2 = []
for j in index_lista_1:
    for i in lista2_tmp:
        if i['{#SNMPINDEXOLD}'] == j:
            lista2.append(i)

# Processa a Lista 3
matches3 = re.findall(r'\.1\.3\.6\.1\.2\.1\.47\.1\.1\.1\.1\.7\.(\d+)\s=\sSTRING:\s*(.*)', output_lista3)
lista3 = [{"{#SNMPINDEX}": index.strip(), "{#IFALIASOLD}": name.strip()} for index, name in matches3[:80] if name.strip()]

# Combina as listas em um único JSON com limpeza de aspas
output = []
for item1, item2, item3 in zip(lista1, lista2, lista3):
    combined = {**item1, **item2, **item3}
    cleaned = {key: remove_extra_quotes(value) for key, value in combined.items()}
    output.append(cleaned)

# Converte o JSON em uma string formatada
json_output = json.dumps(output, indent=4)

# Imprime o resultado
print(json_output)
