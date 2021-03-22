package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("conversor do dicionário de entrada SAS e TXT da PNAD para DCT (Stata).\n")
		fmt.Print("os microdados estão disponíveis em:\n")
		fmt.Print("ftp://ftp.ibge.gov.br/Trabalho_e_Rendimento/Pesquisa_Nacional_por_Amostra_de_Domicilios_anual/microdados/\n\n")

		fmt.Print("Ex.: pnadstata.exe \"pasta do arquivo\\input PES2015.txt\"\n\n")

		fmt.Print("este programa foi escrito por Guilherme Higashi\n")

		os.Exit(0)
	}

	// regex objects
	rgxFileExt := regexp.MustCompile(`(\.\w{3})$`)
	rgxLine := regexp.MustCompile(
		`(?:\w*\s*@)(\d{5})(?:\s+)(\w+)(?:\s+\$?)(\d+\.?\d*)(?:\s*/\*\s*)([^\s].*[^\s])(?:\s*\*/)`)
	rgxIntPart := regexp.MustCompile(`([0-9]+)(?:\.*\d*)`)
	rgxDecPart := regexp.MustCompile(`(?:\d+\.)([0-9]+)`)

	// open input file
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("erro na leitura do arquivo: %s", err)
	} else {
		fmt.Printf("arquivo de entrada: \"%s\"\n", os.Args[1])
	}
	// decode windows1252 encoding
	decodedFile := charmap.Windows1252.NewDecoder().Reader(inFile)

	readBuffer := new(bytes.Buffer)
	_, _ = readBuffer.ReadFrom(decodedFile)
	inFileString := readBuffer.String()
	inFileSubMatches := rgxLine.FindAllStringSubmatch(inFileString, -1)
	err = inFile.Close()
	if err != nil {
		log.Fatalf("erro no fechamento do arquivo de entrada: %s", err)
	}

	// create output file
	outFile, err := os.Create(strings.TrimSuffix(os.Args[1], rgxFileExt.FindString(os.Args[1])) + ".dct")
	if err != nil {
		log.Fatalf("erro na criação do novo arquivo: %s", err)
	} else {
		fmt.Printf("arquivo de saída \"%s\"\n", outFile.Name())
	}

	// create write buffer
	w := bufio.NewWriter(outFile)

	// write header
	_, _ = fmt.Fprint(w, "dictionary {\n")

	nVars := 0
	for _, m := range inFileSubMatches {
		pos, _ := strconv.Atoi(m[1])
		varName := strings.ToLower(m[2])
		varDesc := strings.Replace(strings.Replace(strings.ToLower(m[4]),
			"  ", " ", -1),
			" ?", "?", -1)
		intPart := rgxIntPart.FindStringSubmatch(m[3])
		decPart := rgxDecPart.FindStringSubmatch(m[3])

		var varType string
		var varSize string

		if intPart != nil && decPart == nil {
			//integer
			intPartSize, _ := strconv.Atoi(intPart[1])
			varSize = "%" + intPart[1] + "f"

			if intPartSize <= 2 {
				varType = "byte"
			} else if intPartSize <= 6 {
				varType = "int"
			} else if intPartSize <= 12 {
				varType = "long"
			} else {
				varType = "double"
			}
		} else if intPart != nil && decPart != nil {
			//decimal
			varSize = "%" + m[3] + "f"
			varType = "double"
		}

		_, err = fmt.Fprintf(w, "_column(%d) %7s %7s %7s  \"%s\"\n",
			pos, varType, varName, varSize, varDesc)

		if err != nil {
			log.Fatalf("erro na escrita da linha: %s", err)
		} else {
			nVars++
		}
	}

	// write footer
	_, err = fmt.Fprintf(w, "}\n")

	// flush the write buffer
	_ = w.Flush()

	fmt.Printf("escreveu %d variáveis.\n", nVars)

	err = outFile.Close()
	if err != nil {
		log.Fatalf("erro no fechamento do arquivo de saída: %s", err)
	}

	os.Exit(0)
}
