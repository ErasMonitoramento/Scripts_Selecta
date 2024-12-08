#!/bin/bash

# URL do repositório
repo_url="https://github.com/ErasMonitoramento/Scripts_Selecta.git"

# Diretórios de trabalho e destino
clone_dir="/tmp/implantacao_zabbix/Scripts_Selecta"
source_dir="script_pasta_extenal"  # Nome corrigido
destination_dir="/usr/lib/zabbix/externalscripts"

# Lista de arquivos que devem receber permissão de execução
files_with_permissions=(
    "asname"
    "run_asnamev6"
    "discovery_hw_interfaces_opticas_debian11.py"
    "executa_pppoe_total"
    "mkcontador"
    "mkdiscovery"
    "ppoe.sh"
    "run_asnname"
    "run_oid_snmp"
    "run_signal"
    "status_asn"
)

# Função para verificar e instalar o Git
ensure_git_installed() {
    echo "Verificando se o Git está instalado..."
    if ! command -v git &>/dev/null; then
        echo "Git não encontrado. Instalando..."
        sudo apt update -y
        sudo apt install git -y
    else
        echo "Git já está instalado."
    fi
}

# Função para clonar o repositório
git_clone() {
    if [ -d "$clone_dir" ]; then
        echo "Diretório $clone_dir já existe. Removendo..."
        rm -rf "$clone_dir"
    fi
    echo "Clonando repositório $repo_url para $clone_dir..."
    git clone "$repo_url" "$clone_dir"
}

# Função para mover os arquivos e ajustar permissões
move_files_and_set_permissions() {
    src_path="$clone_dir/$source_dir"
    if [ -d "$src_path" ]; then
        echo "Movendo todos os arquivos e pastas de $src_path para $destination_dir..."
        cp -r "$src_path/"* "$destination_dir/"

        echo "Ajustando permissões para arquivos específicos..."
        for file in "${files_with_permissions[@]}"; do
            if [ -f "$destination_dir/$file" ]; then
                chmod +x "$destination_dir/$file"
                echo "Permissões ajustadas para $destination_dir/$file"
            else
                echo "Arquivo $file não encontrado em $destination_dir"
            fi
        done
    else
        echo "Pasta $source_dir não encontrada no repositório clonado."
    fi
}

# Fluxo principal
main() {
    # Verificar e instalar Git, se necessário
    ensure_git_installed

    # Clonar o repositório
    git_clone

    # Mover arquivos e ajustar permissões
    move_files_and_set_permissions

    echo "Configuração concluída com sucesso."
}

# Executar o fluxo principal
main
