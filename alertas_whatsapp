#!/bin/bash


# URL do repositório

repo_url="https://github.com/ErasMonitoramento/Scripts_Selecta.git"


# Diretórios de trabalho e destino

clone_dir="/tmp/implantacao_zabbix/Scripts_Selecta"

alert_scripts_dir="/usr/lib/zabbix/alertscripts"

external_scripts_dir="/usr/lib/zabbix/externalscripts"

image_base_dir="/var/tmp/zabbix"


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


# Função para clonar repositório

git_clone() {

    if [ -d "$clone_dir" ]; then

        echo "Diretório $clone_dir já existe. Removendo..."

        rm -rf "$clone_dir"

    fi

    echo "Clonando repositório $repo_url para $clone_dir..."

    git clone "$repo_url" "$clone_dir"

}


# Ajustar fuso horário para São Paulo

set_timezone() {

    echo "Ajustando fuso horário para São Paulo..."

    sudo timedatectl set-timezone America/Sao_Paulo

}


# Criar o diretório base para as imagens

create_image_base_dir() {

    if [ ! -d "$image_base_dir" ]; then

        echo "Criando diretório base para as imagens $image_base_dir..."

        mkdir -p "$image_base_dir"

        echo "Ajustando permissão da pasta $image_base_dir..."

        sudo chown -R zabbix:zabbix "$image_base_dir"

    else

        echo "Diretório $image_base_dir já existe."

    fi

}


# Copiar scripts para os diretórios de destino e ajustar permissões

process_scripts() {

    echo "Processando arquivos dos diretórios Alertas_whts e Externalscripts..."

    declare -A specific_dirs=(

        ["Alertas_whts"]="$alert_scripts_dir"

        ["Externalscripts"]="$external_scripts_dir"

    )


    for repo_subdir in "${!specific_dirs[@]}"; do

        src_dir="$clone_dir/$repo_subdir"

        dest_dir="${specific_dirs[$repo_subdir]}"

        if [ -d "$src_dir" ]; then

            for src_file in "$src_dir"/*; do

                if [ -f "$src_file" ]; then

                    dest_file="$dest_dir/$(basename "$src_file")"

                    echo "Copiando $src_file para $dest_file..."

                    cp "$src_file" "$dest_file"

                    echo "Ajustando permissões de execução para $dest_file..."

                    chmod +x "$dest_file"

                fi

            done

        else

            echo "Diretório $repo_subdir não encontrado em $clone_dir."

        fi

    done

}


# Fluxo principal

main() {

    # Verificar e instalar Git, se necessário

    ensure_git_installed


    # Clonar o repositório

    git_clone


    # Ajustar fuso horário

    set_timezone


    # Criar o diretório base para as imagens

    create_image_base_dir


    # Processar os scripts

    process_scripts


    echo "Configuração concluída com sucesso."

}


# Executar o fluxo principal

main

