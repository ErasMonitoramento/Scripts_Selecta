package main

import (
    "bufio"
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

func main() {
    // Define o comando SNMP
    command := "snmpwalk"
    args := []string{"-v2c", "-c", "public@dude", "10.0.102.3", "1.3.6.1.2.1.31.1.1.1.18"}

    // Executa o comando SNMP e captura a saída
    cmd := exec.Command(command, args...)
    var out bytes.Buffer
    cmd.Stdout = &out

    err := cmd.Run()
    if err != nil {
        fmt.Printf("Erro ao executar o comando SNMP: %v\n", err)
        return
    }

    // Cria o arquivo para salvar as descrições
    file, err := os.Create("snmp_descriptions.txt")
    if err != nil {
        fmt.Printf("Erro ao criar o arquivo: %v\n", err)
        return
    }
    defer file.Close()

    // Processa a saída e grava no arquivo apenas as descrições não vazias
    scanner := bufio.NewScanner(&out)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, " = STRING: ")
        if len(parts) == 2 {
            index := parts[0][strings.LastIndex(parts[0], ".")+1:] // Pega o último número como índice
            description := strings.TrimSpace(parts[1])
            
            if description != `""` { // Ignora descrições vazias
                description = strings.Trim(description, `"`)
                _, err := file.WriteString(fmt.Sprintf("%s = %s\n", index, description))
                if err != nil {
                    fmt.Printf("Erro ao escrever no arquivo: %v\n", err)
                    return
                }
            }
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Printf("Erro ao processar a saída: %v\n", err)
    } else {
        fmt.Println("As descrições SNMP foram salvas em 'snmp_descriptions.txt'.")
    }
}
