import sys
import subprocess

def snmpwalk(ip, community, port, oid):
    try:
        result = subprocess.check_output(
            ["snmpwalk", "-v", "2c", "-c", community, f"{ip}:{port}", oid],
            universal_newlines=True
        )
        return result.splitlines()
    except subprocess.CalledProcessError:
        return []

if len(sys.argv) != 4:
    print(f"Uso: {sys.argv[0]} <comunidade> <porta> <ip_especifico>")
    sys.exit(1)

comunidade = sys.argv[1]
porta = sys.argv[2]
ip_especifico = sys.argv[3]

ip = "10.64.100.1"  # Pode ajustar isso conforme necessário

ips = snmpwalk(ip, comunidade, porta, "1.3.6.1.4.1.2011.5.25.177.1.1.2.1.4.0")
asns = snmpwalk(ip, comunidade, porta, "1.3.6.1.4.1.2011.5.25.177.1.1.2.1.2.0")
statuses = snmpwalk(ip, comunidade, porta, "1.3.6.1.4.1.2011.5.25.177.1.1.2.1.5.0")

ip_array = [line.split("STRING: ")[1].strip('"') for line in ips]
asn_array = [line.split("Gauge32: ")[1].strip('"') for line in asns]
status_array = [line.split("INTEGER: ")[1].strip('"') for line in statuses]

status = -1

if len(ip_array) == len(asn_array) == len(status_array):
    for i in range(len(ip_array)):
        current_ip = ip_array[i]
        if current_ip == ip_especifico:
            status = status_array[i]
            break

try:
    status = int(status)
except ValueError:
    status = -1

print(status)
