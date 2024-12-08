#!/bin/bash

# URL do repositório
repo_url="https://github.com/ErasMonitoramento/Scripts_Selecta.git"

# Diretórios de trabalho e destino
clone_dir="/tmp/implantacao_zabbix/Scripts_Selecta"
source_subdir="script_pasta_externo/vlans_huawei"
destination_dir="/usr/lib/zabbix/externalscripts"

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

# Função para mover a pasta vlans_huawei para o destino
move_subfolder() {
    src_path="$clone_dir/$source_subdir"
    if [ -d "$src_path" ]; then
        echo "Movendo a pasta $source_subdir para $destination_dir..."
        mv "$src_path" "$destination_dir/"
        echo "Ajustando permissões da pasta e arquivos em $destination_dir/vlans_huawei..."
        chmod -R +x "$destination_dir/vlans_huawei"
    else
        echo "Subpasta $source_subdir não encontrada no repositório clonado."
    fi
}

# Fluxo principal
main() {
    # Verificar e instalar Git, se necessário
    ensure_git_installed

    # Clonar o repositório
    git_clone

    # Mover a pasta para o destino
    move_subfolder

    echo "Configuração concluída com sucesso."
}

# Executar o fluxo principal
main
