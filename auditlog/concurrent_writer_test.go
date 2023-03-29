// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !tinygo
// +build !tinygo

package auditlog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestConcurrentWriterNoop(t *testing.T) {
	config := NewConfig()
	writer := &concurrentWriter{}
	if err := writer.Init(config); err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	if err := writer.Close(); err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestConcurrentWriterFailsOnInit(t *testing.T) {
	config := NewConfig()
	config.File = "/unexisting.log"
	config.Dir = t.TempDir()
	config.FileMode = fs.FileMode(0777)
	config.DirMode = fs.FileMode(0777)
	config.Formatter = jsonFormatter

	writer := &concurrentWriter{}
	if err := writer.Init(config); err == nil {
		t.Error("expected error")
	}
}

func TestConcurrentWriterWrites(t *testing.T) {
	dir := t.TempDir()
	file, err := os.Create(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Error("failed to create concurrent logger file")
	}
	config := Config{
		File:      file.Name(),
		Dir:       dir,
		FileMode:  fs.FileMode(0777),
		DirMode:   fs.FileMode(0777),
		Formatter: jsonFormatter,
	}
	ts := time.Now()
	expectedLog := &Log{
		Transaction: Transaction{
			UnixTimestamp: ts.UnixNano(),
			ID:            "123",
			Request: &TransactionRequest{
				Method:      "GET",
				URI:         "/test",
				HTTPVersion: "HTTP/1.1",
			},
			Response: &TransactionResponse{
				Status: 201,
			},
		},
	}
	writer := &concurrentWriter{}
	if err := writer.Init(config); err != nil {
		t.Error("failed to init concurrent logger", err)
	}
	if err := writer.Write(expectedLog); err != nil {
		t.Error("failed to write to logger: ", err)
	}

	fileName := fmt.Sprintf("/%s-%s", ts.Format("20060102-150405"), expectedLog.Transaction.ID)
	logFile := path.Join(dir, ts.Format("20060102"), ts.Format("20060102-1504"), fileName)

	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Error("failed to create audit file for concurrent logger")
		return
	}
	actualLog := &Log{}
	if err := json.Unmarshal(logData, actualLog); err != nil {
		t.Errorf("failed to parse json from concurrent audit log: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedLog, actualLog) {
		expectedLogStr, _ := json.Marshal(expectedLog)
		t.Errorf("unexpected log entry, want:\n%s, have:\n%s", expectedLogStr, logData)
	}
}