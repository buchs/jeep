package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"read_creds"
	"strings"
)

type DocProps struct {
	Author     string `xml:"Author"`
	LastAuthor string `xml:"LastAuthor"`
	Created    string `xml:"Created"`
}

type Cell struct {
	Data string `xml:"Data"`
}

type Row struct {
	Cells []Cell `xml:"Cell"`
}

type TableEntry struct {
	Columns []string `xml:"Column"`
	Rows    []Row    `xml:"Row"`
}

type Worksheet struct {
	Name             string     `xml:",ss:Name,attr"`
	Table            TableEntry `xml:"Table"`
	WorksheetOptions string     `xms:"WorksheetOptions"`
}

type Workbook struct {
	XMLName    xml.Name    `xml:"Workbook"`
	DC1        string      `xml:"DocumentProperties"`
	DC2        string      `xml:"ExcelWorkbook"`
	DC3        string      `xml:"Styles"`
	WorkSheets []Worksheet `xml:"Worksheet"`
}

func DumpWorksheet(ws Worksheet, ofp *os.File) {
	foundHeader := false
	keyColumns := []int{-1, -1, -1, -1, -1, -1, -1, -1}
	keyValues := []string{"", "", "", "", "", "", "", ""}
	// columns: Lan Id, Name, Past Supervisor Name, Present Supervisor Name,
	// Job Title, Past Work Unit Desc, Present Workd Unit Desc
	for _, r := range ws.Table.Rows {
		if !foundHeader {
			for idx, c := range r.Cells {
				if idx == 0 && c.Data == "Person Id" {
					foundHeader = true
				}
				if foundHeader {
					switch c.Data {
					case "Lan Id":
						{
							keyColumns[0] = idx
							keyValues[0] = c.Data
						}
					case "User Id":
						{
							keyColumns[1] = idx
							keyValues[1] = c.Data
						}
					case "Name":
						{
							keyColumns[2] = idx
							keyValues[2] = c.Data
						}
					case "Past Supervisor Name":
						{
							keyColumns[3] = idx
							keyValues[3] = c.Data
						}
					case "Present  Supervisor Name":
						{ // yes, double space is in format
							keyColumns[4] = idx
							keyValues[4] = c.Data
						}
					case "Job Title":
						{
							keyColumns[5] = idx
							keyValues[5] = c.Data
						}
					case "Past Work Unit Desc":
						{
							keyColumns[6] = idx
							keyValues[6] = c.Data
						}
					case "Present Work Unit Desc":
						{
							keyColumns[7] = idx
							keyValues[7] = c.Data
						}
					}
				}
			}
			ofp.WriteString(strings.Join(keyValues, ",") + "\n")
		} else {
			for idx, c := range r.Cells {
				for idx2, kCol := range keyColumns {
					if idx == kCol {
						keyValues[idx2] = c.Data
					}
				}
			}
			ofp.WriteString(strings.Join(keyValues, ",") + "\n")
		}
	}
	ofp.WriteString("\n")
}

func main() {
	var lanid, passwd string
	creds := strings.Split(read_creds.ReadCreds(), ",")
	lanid = creds[0]
	passwd = creds[1]

	filename := "HRTransfersReport.xml"
	xmlFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	bites, _ := ioutil.ReadAll(xmlFile)

	var q Workbook
	err = xml.Unmarshal(bites, &q)
	if err != nil {
		fmt.Printf("Unmarshall error: %v", err)
		return
	}

	ofp, ofperr := os.Create("worksheet_dump.csv")
	if ofperr != nil {
		panic(ofperr)
	}
	defer ofp.Close()
	for i, _ := range q.WorkSheets {
		if q.WorkSheets[i].Name == "Company (Summary)" {
			DumpWorksheet(q.WorkSheets[i], ofp)
		}
	}
}
