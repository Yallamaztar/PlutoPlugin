package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/term"
)

type Logger struct {
	slug string
	base *log.Logger
	file *os.File
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[38;5;135m"
	colorCyan   = "\033[36m"
)

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func supportsColor() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func colorize(s, color string) string {
	if !supportsColor() {
		return s
	}
	return color + s + colorReset
}

func enableWindowsANSI() {
	if runtime.GOOS != "windows" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	const enableVirtualTerminalProcessing = 0x0004

	var mode uint32
	handle := os.Stdout.Fd()

	getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	setConsoleMode.Call(handle, uintptr(mode|enableVirtualTerminalProcessing))
}

func levelTag(level string) string {
	switch level {
	case "INFO":
		return colorize("[INFO]", colorGreen)
	case "DEBUG":
		return colorize("[DEBUG]", colorCyan)
	case "WARN":
		return colorize("[WARN]", colorYellow)
	case "ERROR":
		return colorize("[ERROR]", colorRed)
	case "FATAL":
		return colorize("[FATAL]", colorRed)
	case "PANIC":
		return colorize("[PANIC]", colorRed)
	default:
		return "[" + level + "]"
	}
}

func New(slug, logPath string) *Logger {
	enableWindowsANSI()
	writers := []io.Writer{os.Stdout}

	var file *os.File
	if logPath != "" {
		var err error
		file, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			writers = append(writers, file)
		} else {
			fmt.Printf("Failed to open log file %s: %v\n", logPath, err)
		}
	}

	multi := io.MultiWriter(writers...)
	return &Logger{
		slug: slug,
		base: log.New(multi, "", log.LstdFlags|log.Lshortfile),
		file: file,
	}
}

func (l *Logger) prefix() string {
	return colorize(fmt.Sprintf("[%s]", l.slug), colorPurple)
}

func (l *Logger) Println(v ...any) {
	l.base.Println(append([]any{l.prefix()}, v...)...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.base.Printf(l.prefix()+" "+format, v...)
}

func (l *Logger) Infoln(v ...any) {
	l.base.Println(append([]any{l.prefix(), levelTag("INFO")}, v...)...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.base.Printf(l.prefix()+" "+levelTag("INFO")+" "+format, v...)
}

func (l *Logger) Warnln(v ...any) {
	l.base.Println(append([]any{l.prefix(), levelTag("WARN")}, v...)...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.base.Printf(l.prefix()+" "+levelTag("WARN")+" "+format, v...)
}

func (l *Logger) Errorln(v ...any) {
	l.base.Println(append([]any{l.prefix(), levelTag("ERROR")}, v...)...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.base.Printf(l.prefix()+" "+levelTag("ERROR")+" "+format, v...)
}

func (l *Logger) Fatal(v ...any) {
	l.base.Fatal(append([]any{l.prefix(), levelTag("FATAL")}, v...)...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.base.Fatalf(l.prefix()+" "+levelTag("FATAL")+" "+format, v...)
}
