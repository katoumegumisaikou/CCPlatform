package service

import (
	"ccplatform/internal/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupService 数据备份服务。
type BackupService struct {
	backupPath string
}

func NewBackupService() *BackupService {
	path := "backups"
	if config.Cfg.Backup.BackupPath != "" {
		path = config.Cfg.Backup.BackupPath
	}
	os.MkdirAll(path, 0755)
	return &BackupService{backupPath: path}
}

// BackupFile 备份文件信息。
type BackupFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Time     string `json:"time"`
}

// Create 执行数据库备份（mysqldump）。
func (s *BackupService) Create() (*BackupFile, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(s.backupPath, filename)

	dsn := config.Cfg.Database.DSN()
	parts := strings.Split(dsn, "/")
	dbName := "ccplatform"
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		dbName = strings.Split(lastPart, "?")[0]
	}

	cmd := exec.Command("mysqldump", "-h", "127.0.0.1", dbName,
		"--result-file="+filepath, "--single-transaction", "--quick")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("mysqldump failed: %v, output: %s", err, output)
	}

	info, _ := os.Stat(filepath)
	return &BackupFile{
		Filename: filename,
		Size:     info.Size(),
		Time:     timestamp,
	}, nil
}

// List 列出备份文件。
func (s *BackupService) List() ([]BackupFile, error) {
	entries, err := os.ReadDir(s.backupPath)
	if err != nil {
		return nil, err
	}
	var files []BackupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		files = append(files, BackupFile{
			Filename: entry.Name(),
			Size:     info.Size(),
			Time:     info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Time > files[j].Time
	})
	return files, nil
}

// Restore 从备份文件恢复。
func (s *BackupService) Restore(filename string) error {
	filepath := filepath.Join(s.backupPath, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", filename)
	}

	dsn := config.Cfg.Database.DSN()
	parts := strings.Split(dsn, "/")
	dbName := "ccplatform"
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		dbName = strings.Split(lastPart, "?")[0]
	}

	cmd := exec.Command("mysql", "-h", "127.0.0.1", dbName)
	input, _ := os.ReadFile(filepath)
	cmd.Stdin = strings.NewReader(string(input))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore failed: %v, output: %s", err, output)
	}
	return nil
}

// Delete 删除备份文件。
func (s *BackupService) Delete(filename string) error {
	filepath := filepath.Join(s.backupPath, filename)
	return os.Remove(filepath)
}
