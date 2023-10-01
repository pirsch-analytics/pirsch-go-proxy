package logbuch

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const (
	defaultFiles      = 1
	defaultFileSize   = 1024 * 1024 * 5 // 5 MB
	defaultBufferSize = 4096            // 4 KB
)

// NameSchema is an interface to generate log file names.
// If you implement this interface, make sure the Name() method returns unique file names.
type NameSchema interface {
	// Name returns the next file name used to store log data.
	Name() string
}

// RollingFileAppender is a manager for rolling log files.
// It needs to be closed using the Close() method.
type RollingFileAppender struct {
	files           int
	fileSize        int
	fileName        NameSchema
	fileDir         string
	buffer          []byte
	maxBufferSize   int
	currentFile     *os.File
	currentFileSize int
	fileNames       []string
	m               sync.Mutex
}

// NewRollingFileAppender creates a new RollingFileAppender.
// If you pass values below or equal to 0 for files, size or bufferSize, default values will be used.
// The file output directory is created if required and can be left empty to use the current directory.
// The filename schema is required. Note that the rolling file appender uses the filename schema you provide,
// so if it returns the same name on each call, it will overwrite the existing log file.
// The RollingFileAppender won't clean the log directory on startup. Old log files will stay in place.
func NewRollingFileAppender(files, size, bufferSize int, dir string, filename NameSchema) (*RollingFileAppender, error) {
	if files <= 0 {
		files = defaultFiles
	}

	if size <= 0 {
		size = defaultFileSize
	}

	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	if filename == nil {
		return nil, errors.New("filename schema must be specified")
	}

	if err := os.MkdirAll(dir, 0774); err != nil {
		return nil, err
	}

	appender := &RollingFileAppender{files: files,
		fileSize:      size,
		fileName:      filename,
		fileDir:       dir,
		buffer:        make([]byte, 0, bufferSize),
		maxBufferSize: bufferSize,
		fileNames:     make([]string, 0, files)}

	if err := appender.nextFile(); err != nil {
		return nil, err
	}

	return appender, nil
}

// Write writes given data to the rolling log files.
// This might not happen immediately as the RollingFileAppender uses a buffer.
// If you want the data to be persisted, call Flush().
func (appender *RollingFileAppender) Write(p []byte) (n int, err error) {
	appender.m.Lock()
	defer appender.m.Unlock()

	if len(appender.buffer) >= appender.maxBufferSize {
		if err := appender.flush(); err != nil {
			return 0, err
		}
	}

	appender.buffer = append(appender.buffer, p...)
	return len(p), nil
}

// Flush writes all log data currently in buffer into the currently active log file.
func (appender *RollingFileAppender) Flush() error {
	appender.m.Lock()
	defer appender.m.Unlock()
	return appender.flush()
}

// Close flushes the log data and closes all open file handlers.
func (appender *RollingFileAppender) Close() error {
	appender.m.Lock()
	defer appender.m.Unlock()

	if err := appender.flush(); err != nil {
		return err
	}

	return appender.currentFile.Close()
}

func (appender *RollingFileAppender) flush() error {
	n, err := appender.currentFile.Write(appender.buffer)

	if err != nil {
		return err
	}

	appender.buffer = appender.buffer[:0]
	appender.currentFileSize += n

	if appender.currentFileSize >= appender.fileSize {
		if err := appender.nextFile(); err != nil {
			return err
		}
	}

	return nil
}

func (appender *RollingFileAppender) nextFile() error {
	if appender.currentFile != nil {
		if err := appender.currentFile.Close(); err != nil {
			return err
		}
	}

	path := filepath.Join(appender.fileDir, appender.fileName.Name())
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0664)

	if err != nil {
		return err
	}

	appender.currentFile = f
	appender.currentFileSize = 0
	return appender.updateFiles(path)
}

func (appender *RollingFileAppender) updateFiles(path string) error {
	appender.fileNames = append(appender.fileNames, path)

	if len(appender.fileNames) > appender.files {
		n := len(appender.fileNames) - appender.files
		filesToDelete := appender.fileNames[:n]
		appender.fileNames = appender.fileNames[n:]

		for _, file := range filesToDelete {
			if err := os.Remove(file); err != nil {
				return err
			}
		}
	}

	return nil
}
