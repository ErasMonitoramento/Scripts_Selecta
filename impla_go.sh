#!/bin/bash

# URL do repositório
repo_url="https://github.com/ErasMonitoramento/Scripts_Selecta.git"

# Diretórios de trabalho e destino
clone_dir="/tmp/implantacao_zabbix/Scripts_Selecta"
eras_dir="/etc/eras"
zabbix_dir="/usr/lib/zabbix/externalscripts"
exclude_dir="vlans_huawei"

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

# Função para criar a pasta eras caso não exista
ensure_eras_dir_exists() {
    if [ ! -d "$eras_dir" ]; then
        echo "Criando diretório $eras_dir..."
        sudo mkdir -p "$eras_dir"
        echo "Diretório $eras_dir criado."
    else
        echo "Diretório $eras_dir já existe."
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

# Função para mover arquivos para os diretórios apropriados
move_files() {
    echo "Movendo arquivos e pastas para os diretórios de destino..."

    # Itera sobre as subpastas e arquivos no repositório
    for item in "$clone_dir/Scripts em Go/"*; do
        base_item=$(basename "$item")
        if [ "$base_item" == "$exclude_dir" ]; then
            echo "Movendo $base_item para $zabbix_dir..."
            sudo mv "$item" "$zabbix_dir/"
        else
            echo "Movendo $base_item para $eras_dir..."
            sudo mv "$item" "$eras_dir/"
        fi
    done

    echo "Arquivos e pastas movidos com sucesso."
}

# Fluxo principal
main() {
    # Verificar e instalar Git, se necessário
    ensure_git_installed

    # Criar o diretório eras, se necessário
    ensure_eras_dir_exists

    # Clonar o repositório
    git_clone

    # Mover arquivos e pastas
    move_files

    echo "Configuração concluída com sucesso."
}

# Executar o fluxo principal
main
