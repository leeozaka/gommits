package utils

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
	"github.com/xuri/excelize/v2"
)

func ExportToExcel(commits []models.CommitInfo, repoPath string) error {
	f := excelize.NewFile()

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	repoName := git.GetRepositoryName(repoPath)
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

func WriteExcel() {
	repoPath := "."

	if !git.IsGitRepo(repoPath) {
		fmt.Println("Current directory is not a git repository")
		return
	}

	commits, _, err := git.GatherCommits(repoPath, "", "main", true)
	if err != nil {
		fmt.Printf("Error gathering commits: %v\n", err)
		return
	}

	if len(commits) > 50 {
		commits = commits[:50]
	}

	err = ExportToExcel(commits, repoPath)
	if err != nil {
		fmt.Printf("Error creating Excel file: %v\n", err)
		return
	}
}
