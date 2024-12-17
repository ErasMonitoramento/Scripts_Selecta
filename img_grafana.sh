#!/bin/bash

# URL do repositório
repo_url="https://github.com/ErasMonitoramento/Scripts_Selecta.git"

# Diretórios de trabalho
clone_dir="/tmp/implantacao_zabbix/Scripts_Selecta"
source_dir="Imagens Grafana/logos"  # Caminho para a pasta logos
destination_dir="/usr/share/grafana/public/img/logos"

# Lista de plugins do Grafana para instalar
grafana_plugins=(
    "alexanderzobnin-zabbix-app"
    "marcusolsson-dynamictext-panel"
    "grafana-clock-panel"
    "volkovlabs-echarts-panel"
    "volkovlabs-variable-panel"
    "gapit-htmlgraphics-panel"
    "heywesty-trafficlight-panel"
    "mesak-imagesave-panel"
    "orchestracities-map-panel"
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

# Função para verificar e instalar o grafana-cli
ensure_grafana_cli_installed() {
    echo "Verificando se o grafana-cli está instalado..."
    if ! command -v grafana-cli &>/dev/null; then
        echo "grafana-cli não encontrado. Certifique-se de que o Grafana está instalado corretamente."
        exit 1
    else
        echo "grafana-cli encontrado."
    fi
}

# Função para instalar os plugins do Grafana
install_grafana_plugins() {
    ensure_grafana_cli_installed
    echo "Instalando plugins do Grafana..."
    for plugin in "${grafana_plugins[@]}"; do
        echo "Instalando plugin: $plugin"
        sudo grafana-cli plugins install "$plugin"
        if [ $? -ne 0 ]; then
            echo "Erro ao instalar o plugin: $plugin. Verifique o grafana-cli."
        fi
    done
    echo "Reiniciando o serviço do Grafana para aplicar as alterações..."
    sudo systemctl restart grafana-server
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

# Função para atualizar ou copiar a pasta logos
update_logos() {
    src_path="$clone_dir/$source_dir"
    if [ -d "$src_path" ]; then
        if [ -d "$destination_dir" ]; then
            echo "A pasta logos já existe no destino. Atualizando o conteúdo..."
            sudo rsync -av --delete "$src_path/" "$destination_dir/"
        else
            echo "A pasta logos não existe no destino. Copiando pela primeira vez..."
            sudo cp -r "$src_path" "$destination_dir"
        fi
        echo "Operação concluída com sucesso."
    else
        echo "Pasta logos não encontrada no repositório clonado."
    fi
}

# Fluxo principal
main() {
    # Verificar e instalar Git, se necessário
    ensure_git_installed

    # Clonar o repositório
    git_clone

    # Atualizar ou copiar a pasta logos para o destino
    update_logos

    # Instalar plugins do Grafana
    install_grafana_plugins

    echo "Configuração concluída com sucesso."
}

# Executar o fluxo principal
main
