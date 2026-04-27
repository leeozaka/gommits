package utils

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/leeozaka/gommits/internal/models"
	"github.com/xuri/excelize/v2"
)

func ExportToExcel(commits []models.CommitInfo, repoPath, repoName string) error {
	f := excelize.NewFile()

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fileName := fmt.Sprintf("%s_commits.xlsx", repoName)

	sheetName := "Commits"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %v", err)
	}

	f.SetActiveSheet(index)

	f.DeleteSheet("Sheet1")

	headers := []string{"Commit Hash", "Author Name", "Author Email", "Commit Date", "Commit Message", "Files Changed"}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create header style: %v", err)
	}

	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create data style: %v", err)
	}

	for i, header := range headers {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	columnWidths := map[string]float64{
		"A": 15, // Commit Hash
		"B": 20, // Author Name
		"C": 25, // Author Email
		"D": 18, // Commit Date
		"E": 40, // Commit Message
		"F": 35, // Files Changed
	}

	for col, width := range columnWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	row := 2
	for _, commit := range commits {
		filesStr := ""
		if len(commit.Files) > 0 {
			for i, file := range commit.Files {
				if i > 0 {
					filesStr += "\n"
				}
				filesStr += file
			}
		} else {
			filesStr = "No files changed"
		}

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), commit.Hash)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), commit.Author)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), commit.Email)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), commit.Date)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), commit.Message)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), filesStr)

		for col := 'A'; col <= 'F'; col++ {
			cell := string(col) + strconv.Itoa(row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}

		row++
	}

	if len(commits) > 0 {
		tableRange := fmt.Sprintf("A1:F%d", len(commits)+1)
		err = f.AddTable(sheetName, &excelize.Table{
			Range:             tableRange,
			Name:              "CommitsTable",
			StyleName:         "TableStyleMedium2",
			ShowFirstColumn:   false,
			ShowLastColumn:    false,
			ShowRowStripes:    &[]bool{true}[0],
			ShowColumnStripes: false,
		})
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	summarySheet := "Summary"
	summaryIndex, err := f.NewSheet(summarySheet)
	if err == nil {
		f.SetCellValue(summarySheet, "A1", "Repository Summary")
		f.SetCellValue(summarySheet, "A2", "Repository Name:")
		f.SetCellValue(summarySheet, "B2", repoName)
		f.SetCellValue(summarySheet, "A3", "Total Commits:")
		f.SetCellValue(summarySheet, "B3", len(commits))
		f.SetCellValue(summarySheet, "A4", "Repository Path:")
		f.SetCellValue(summarySheet, "B4", repoPath)

		titleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
				Size: 14,
			},
		})
		f.SetCellStyle(summarySheet, "A1", "A1", titleStyle)

		labelStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
		})
		f.SetCellStyle(summarySheet, "A2", "A4", labelStyle)

		f.SetColWidth(summarySheet, "A", "A", 20)
		f.SetColWidth(summarySheet, "B", "B", 40)

		f.SetActiveSheet(summaryIndex)
	}

	fullPath := filepath.Join(repoPath, fileName)
	if err := f.SaveAs(fullPath); err != nil {
		return fmt.Errorf("failed to save Excel file: %v", err)
	}

	return nil
}

func WriteExcel(svc interface {
	IsGitRepo(string) bool
	GatherCommits(string, string, string, bool) ([]models.CommitInfo, string, error)
	GetRepositoryName(string) string
}) {
	repoPath := "."

	if !svc.IsGitRepo(repoPath) {
		fmt.Println("Current directory is not a git repository")
		return
	}

	commits, _, err := svc.GatherCommits(repoPath, "", "main", true)
	if err != nil {
		fmt.Printf("Error gathering commits: %v\n", err)
		return
	}

	if len(commits) > 50 {
		commits = commits[:50]
	}

	repoName := svc.GetRepositoryName(repoPath)
	err = ExportToExcel(commits, repoPath, repoName)
	if err != nil {
		fmt.Printf("Error creating Excel file: %v\n", err)
		return
	}
}

// dbaChecklistItems holds the fixed checklist required by the automation system on every DBA sheet.
var dbaChecklistItems = []string{
	"Os scripts (alteração objetos de BD) foram homologados?",
	"Todos os objetos compilaram em desenvolvimento?",
	"Todos os objetos estão atualizados no TFS?",
	"Procedures Novas - Segue o padrão de nomenclatura (Sigla do Objeto)_(Sigla do Sistema)_(Iniciais do Processo)_(Descrição sugestiva da Ação)",
	"Procedures - O cabeçalho está atualizado?",
	"Procedures - Tem SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED?",
	"Procedures - Tem IF OBJECT_ID('[dbo].[<MINHAPROCEDURE>]') IS NOT NULL DROP PROCEDURE [dbo].[<MINHAPROCEDURE>]?",
	"Procedures - Não é permitido utilizar as variáveis NVARCHAR(MAX), NCHAR(MAX), VARCHAR(MAX), CHAR(MAX) > 100",
	"Procedures - Não são permitidos comandos DML (INSERT, UPDATE, DELETE, SELECT) via Linked Server, tem que criar uma Procedure no destino",
	"Procedures - Não são permitidos comandos DML (INSERT, UPDATE, DELETE, SELECT) que estejam fora do escopo de uma procedure",
	"Procedures - Não serão aprovadas procedures que utilizem CURSORES em seu escopo / Se possível não usar Loop/While.",
	"Procedures - Evite realizar funções nas cláusulas where",
	"Procedures - Não utilizar * para o retorno da consulta (Exemplo: SELECT *, COUNT(*))",
	"Procedures - Não utilizar comandos descontinuados a partir da versão 2000 (Exemplo: '!=' ou '=*' para fazer JOIN's, referenciar um Alias de uma coluna na cláusula WHERE)",
	"Procedures - Devem ser identificados explicitamente os nomes das colunas para a tabela que você está inserindo dados com a instrução INSERT",
	"Procedures - Utilize o comando SELECT ao invés de SET para atribuir valores para as variáveis",
	"Procedures - Cuidado com o DISTINCT / Sempre que possível, utilize a instrução UNION ALL (mais rápida) ao invés de UNION",
	"Procedures - Favor criar as Tabelas para trabalhar com tabelas temporarias, evitar usar SELECT Campos INTO #TEMP FROM Tabela",
	"Procedures - Não utilizar SubQuery",
	"Procedures - Ao invés de utilizarmos um COUNT(1), podemos utilizar a cláusula IF EXISTS, fazendo com que o SQL Server retorne sucesso logo no primeiro registro encontrado",
}

// Shared style constructors used by both the Serviços and DBA sheet writers.

func newDotnetTitleStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
}

func newDotnetHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
}

func newDotnetDataStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
}

func newDotnetQAWarnStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Color: "#FF0000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
}

func ExportDotnetExcel(services []models.DotnetEntry, up, down []models.DBAEntry, repoPath, repoName string) error {
	f := excelize.NewFile()

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fileName := fmt.Sprintf("%s_dotnet.xlsx", repoName)

	sheetName := "Serviços"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %v", err)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	titleStyle, err := newDotnetTitleStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create title style: %v", err)
	}

	headerStyle, err := newDotnetHeaderStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create header style: %v", err)
	}

	subLabelStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Italic: true,
			Size:   10,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create sub-label style: %v", err)
	}

	dataStyle, err := newDotnetDataStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create data style: %v", err)
	}

	newEntryStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#008000",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create new-entry style: %v", err)
	}

	f.SetCellValue(sheetName, "A1", "Objetos a serem atualizados")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	headers := []string{"SEQUENCIA", "CAMINHO E OBJETO", "TIPO", "BASE"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + "3"
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	f.SetCellValue(sheetName, "B4", "Lista de Serviços Alterados")
	f.SetCellStyle(sheetName, "B4", "B4", subLabelStyle)
	f.SetCellValue(sheetName, "B5", "Publicado")
	f.SetCellStyle(sheetName, "B5", "B5", subLabelStyle)

	row := 6
	for _, entry := range services {
		rowStr := strconv.Itoa(row)
		f.SetCellValue(sheetName, "A"+rowStr, entry.Sequence)
		f.SetCellValue(sheetName, "B"+rowStr, entry.Path)
		f.SetCellValue(sheetName, "C"+rowStr, entry.Type)

		for col := 'A'; col <= 'D'; col++ {
			cell := string(col) + rowStr
			if entry.Type == "NEW" && col == 'C' {
				f.SetCellStyle(sheetName, cell, cell, newEntryStyle)
			} else {
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			}
		}

		row++
	}

	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 80)
	f.SetColWidth(sheetName, "C", "C", 10)
	f.SetColWidth(sheetName, "D", "D", 10)

	if err := writeDBASheet(f, up, down); err != nil {
		return err
	}

	fullPath := filepath.Join(repoPath, fileName)
	if err := f.SaveAs(fullPath); err != nil {
		return fmt.Errorf("failed to save Excel file: %v", err)
	}

	return nil
}

func writeDBASheet(f *excelize.File, up, down []models.DBAEntry) error {
	sheet := "DBA"
	if _, err := f.NewSheet(sheet); err != nil {
		return fmt.Errorf("failed to create DBA sheet: %v", err)
	}

	titleStyle, err := newDotnetTitleStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create DBA title style: %v", err)
	}

	headerStyle, err := newDotnetHeaderStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create DBA header style: %v", err)
	}

	dataStyle, err := newDotnetDataStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create DBA data style: %v", err)
	}

	qaWarnStyle, err := newDotnetQAWarnStyle(f)
	if err != nil {
		return fmt.Errorf("failed to create DBA QA warn style: %v", err)
	}

	checklistLabelStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	if err != nil {
		return fmt.Errorf("failed to create checklist label style: %v", err)
	}

	checkmarkStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return fmt.Errorf("failed to create checkmark style: %v", err)
	}

	// Row 1: title (merged A1:E1)
	f.MergeCell(sheet, "A1", "E1")
	f.SetCellValue(sheet, "A1", "Objetos a serem atualizados")
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)

	// Row 3: checklist header
	f.MergeCell(sheet, "A3", "D3")
	f.SetCellValue(sheet, "A3", "CHECKLIST BANCO DE DADOS")
	f.SetCellStyle(sheet, "A3", "D3", headerStyle)
	f.SetCellValue(sheet, "E3", "FEITO")
	f.SetCellStyle(sheet, "E3", "E3", headerStyle)

	// Rows 5-24: checklist items
	for i, item := range dbaChecklistItems {
		rowStr := strconv.Itoa(i + 5)
		f.MergeCell(sheet, "A"+rowStr, "D"+rowStr)
		f.SetCellValue(sheet, "A"+rowStr, item)
		f.SetCellStyle(sheet, "A"+rowStr, "D"+rowStr, checklistLabelStyle)
		f.SetCellValue(sheet, "E"+rowStr, "✓")
		f.SetCellStyle(sheet, "E"+rowStr, "E"+rowStr, checkmarkStyle)
	}

	// Row 25: readiness prompt (literal "#REF!" preserved — expected by automation)
	f.MergeCell(sheet, "A25", "D25")
	f.SetCellValue(sheet, "A25", "Você está pronto para encaminhar a lista de objetos?")
	f.SetCellStyle(sheet, "A25", "D25", checklistLabelStyle)
	f.SetCellValue(sheet, "E25", "#REF!")
	f.SetCellStyle(sheet, "E25", "E25", checkmarkStyle)

	// Row 26: QA warning (merged A26:E26)
	f.MergeCell(sheet, "A26", "E26")
	f.SetCellValue(sheet, "A26", "Por favor, qualquer dúvida consultar a equipe de QA.")
	f.SetCellStyle(sheet, "A26", "E26", qaWarnStyle)

	// Row 28: data headers
	dbaHeaders := []string{"SEQUEN", "CAMINHO E OBJETO", "TIPO", "SERVIDOR", "BASE"}
	for i, h := range dbaHeaders {
		cell := string(rune('A'+i)) + "28"
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Row 29: "Scripts de Up" section label
	f.MergeCell(sheet, "A29", "E29")
	f.SetCellValue(sheet, "A29", "Scripts de Up")
	f.SetCellStyle(sheet, "A29", "E29", headerStyle)

	// Up entries starting at row 30
	row := 30
	for _, e := range up {
		rowStr := strconv.Itoa(row)
		f.SetCellValue(sheet, "A"+rowStr, e.Sequence)
		f.SetCellValue(sheet, "B"+rowStr, e.Path)
		f.SetCellValue(sheet, "C"+rowStr, "script")
		for col := 'A'; col <= 'E'; col++ {
			f.SetCellStyle(sheet, string(col)+rowStr, string(col)+rowStr, dataStyle)
		}
		row++
	}

	// "Scripts de Down" section label
	downLabelRow := strconv.Itoa(row)
	f.MergeCell(sheet, "A"+downLabelRow, "E"+downLabelRow)
	f.SetCellValue(sheet, "A"+downLabelRow, "Scripts de Down")
	f.SetCellStyle(sheet, "A"+downLabelRow, "E"+downLabelRow, headerStyle)
	row++

	// Down entries
	for _, e := range down {
		rowStr := strconv.Itoa(row)
		f.SetCellValue(sheet, "A"+rowStr, e.Sequence)
		f.SetCellValue(sheet, "B"+rowStr, e.Path)
		f.SetCellValue(sheet, "C"+rowStr, "script")
		for col := 'A'; col <= 'E'; col++ {
			f.SetCellStyle(sheet, string(col)+rowStr, string(col)+rowStr, dataStyle)
		}
		row++
	}

	f.SetColWidth(sheet, "A", "A", 10)
	f.SetColWidth(sheet, "B", "B", 80)
	f.SetColWidth(sheet, "C", "C", 10)
	f.SetColWidth(sheet, "D", "D", 15)
	f.SetColWidth(sheet, "E", "E", 10)

	return nil
}
