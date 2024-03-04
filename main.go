package main

import (
    "bytes"
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "os"
    "github.com/SebastiaanKlippert/go-wkhtmltopdf"
    "github.com/google/uuid"
    "strings"
)

type Parcela struct {
    DataParcela   string  `json:"DataParcela"`
    Idade         int     `json:"Idade"`
    Mes           int     `json:"Mes"`
    Juros         float64 `json:"Juros"`
    Amortizacao   float64 `json:"Amortizacao"`
    Dfi           float64 `json:"Dfi"`
    Mip           float64 `json:"Mip"`
    IOFAdicional  float64 `json:"IOFAdicional"`
    IOFMensal     float64 `json:"IOFMensal"`
    TAC           float64 `json:"TAC"`
    Prestacao     float64 `json:"Prestacao"`
    Saldo         float64 `json:"Saldo"`
}

func main() {
    parcelas, err := lerParcelasDoArquivo("JsonSimulacao.json")
    if err != nil {
        log.Fatal("Erro ao ler as parcelas do arquivo:", err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        uniqueID := geraPdf()
        w.Write([]byte(uniqueID))
    })

    http.HandleFunc("/template", func(w http.ResponseWriter, r *http.Request) {
        tmpl, err := template.ParseFiles("c_a41.html")
        if err != nil {
            http.Error(w, "Erro ao analisar o modelo HTML", http.StatusInternalServerError)
            return
        }

        var tpl bytes.Buffer
        if err := tmpl.Execute(&tpl, struct{ Parcelas []Parcela }{Parcelas: parcelas}); err != nil {
            http.Error(w, "Erro ao executar o modelo HTML", http.StatusInternalServerError)
            return
        }

        w.Write(tpl.Bytes())
    })

    http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

    log.Fatal(http.ListenAndServe(":8081", nil))
}

func lerParcelasDoArquivo(nomeArquivo string) ([]Parcela, error) {
    file, err := os.Open(nomeArquivo)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var parcelas []Parcela
    if err := json.NewDecoder(file).Decode(&parcelas); err != nil {
        return nil, err
    }

    return parcelas, nil
}

func geraPdf() string {
    converter, err := wkhtmltopdf.NewPDFGenerator()
    if err != nil {
        log.Fatalf("Erro ao criar o conversor: %v", err)
    }

    url := "http://localhost:8081/template"

    converter.AddPage(wkhtmltopdf.NewPage(url))

    err = converter.Create()
    if err != nil {
        log.Fatalf("Erro ao converter HTML para PDF: %v", err)
    }

    uniqueID := uuid.New().String()

    outputPath := "pdf/"
    if err := os.MkdirAll(outputPath, 0755); err != nil {
        log.Fatalf("Erro ao criar a pasta de saída: %v", err)
    }

    fileName := strings.Join([]string{outputPath, uniqueID, ".pdf"}, "")

    err = converter.WriteFile(fileName)
    if err != nil {
        log.Fatalf("Erro ao salvar o PDF em um arquivo: %v", err)
    }

    log.Printf("O arquivo PDF %s foi criado dentro da pasta 'pdf.'", fileName)

    return uniqueID
}
