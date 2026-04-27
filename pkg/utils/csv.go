package utils

import (
	"encoding/csv"
	"os"

	"github.com/leeozaka/gommits/internal/models"
)

func ExportToCSV(commits []models.CommitInfo, csvPath string) error {
	file, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"commit_hash", "author_name", "author_email", "commit_date", "commit_message", "file_path"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, c := range commits {
		if len(c.Files) > 0 {
			for _, f := range c.Files {
				row := []string{c.Hash, c.Author, c.Email, c.Date, c.Message, f}
				if err := writer.Write(row); err != nil {
					return err
				}
			}
		} else {
			row := []string{c.Hash, c.Author, c.Email, c.Date, c.Message, ""}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return writer.Error()
}
