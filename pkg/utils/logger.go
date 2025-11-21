package utils

import (
    "fmt"
    "log"
    "os"
    "time"
)

type Logger struct {
    *log.Logger
    debug bool
    useColors bool
}

func NewLogger(debug bool) *Logger {
    logger := &Logger{
        Logger: log.New(os.Stdout, "", 0),
        debug:  debug,
        useColors: true, // Enable colors by default
    }
    
    // Disable colors jika running di environment tanpa support
    if os.Getenv("TERM") == "" || os.Getenv("NO_COLOR") != "" {
        logger.useColors = false
    }
    
    return logger
}

// Color constants
const (
    colorReset  = "\033[0m"
    colorRed    = "\033[31m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorBlue   = "\033[34m"
    colorPurple = "\033[35m"
    colorCyan   = "\033[36m"
    colorWhite  = "\033[37m"
    colorGray   = "\033[90m"
)

func (l *Logger) Info(format string, v ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf(format, v...)
    
    if l.useColors {
        l.Printf("%s[%s INFO]%s %s", colorCyan, timestamp, colorReset, message)
    } else {
        l.Printf("[%s INFO] %s", timestamp, message)
    }
}

func (l *Logger) Debug(format string, v ...interface{}) {
    if l.debug {
        timestamp := time.Now().Format("15:04:05")
        message := fmt.Sprintf(format, v...)
        
        if l.useColors {
            l.Printf("%s[%s DEBUG]%s %s", colorGray, timestamp, colorReset, message)
        } else {
            l.Printf("[%s DEBUG] %s", timestamp, message)
        }
    }
}

func (l *Logger) Warn(format string, v ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf(format, v...)
    
    if l.useColors {
        l.Printf("%s[%s WARN]%s %s", colorYellow, timestamp, colorReset, message)
    } else {
        l.Printf("[%s WARN] %s", timestamp, message)
    }
}

func (l *Logger) Error(format string, v ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf(format, v...)
    
    if l.useColors {
        l.Printf("%s[%s ERROR]%s %s", colorRed, timestamp, colorReset, message)
    } else {
        l.Printf("[%s ERROR] %s", timestamp, message)
    }
}

func (l *Logger) Fatal(format string, v ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf(format, v...)
    
    if l.useColors {
        l.Printf("%s[%s FATAL]%s %s", colorRed, timestamp, colorReset, message)
    } else {
        l.Printf("[%s FATAL] %s", timestamp, message)
    }
    os.Exit(1)
}

// Special method untuk banner/logos
func (l *Logger) Banner(message string) {
    if l.useColors {
        l.Printf("%s%s%s", colorGreen, message, colorReset)
    } else {
        l.Print(message)
    }
}

func (l *Logger) Success(format string, v ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    message := fmt.Sprintf(format, v...)
    
    if l.useColors {
        l.Printf("%s[%s SUCCESS]%s %s", colorGreen, timestamp, colorReset, message)
    } else {
        l.Printf("[%s SUCCESS] %s", timestamp, message)
    }
}