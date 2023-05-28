package liblog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var Log *Logger
var LogFileName = "webrtcMedia.log"
var LogLIBVideoFileName = "webrtcMedia_video.log"

const (
	LOG_ERROR int = 1
	LOG_WARN  int = 2
	LOG_FLOW  int = 3
	LOG_INFO  int = 4
	LOG_DEUBG int = 5
)

type Logger struct {
	LogName  string
	LogFile  string
	LogLevel int
	IsPrint  bool
	LogChan  chan string
	Done     chan bool
	RWLock   *sync.RWMutex
}

func CreateLogger(name string, logfile string, loglevel int, isprint bool) *Logger {
	logger := &Logger{LogName: name, LogFile: logfile, LogLevel: loglevel, IsPrint: isprint}
	err := logger.logInit()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	logger.RWLock = new(sync.RWMutex)

	if Log == nil {
		Log = logger
	}

	return logger
}

func (self *Logger) logInit() error {
	if self.LogChan != nil {
		return fmt.Errorf("[%s] has log init. so return.\n", self.LogName)
	}

	log, err := os.OpenFile(self.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("open log file[%s] fail.\n", self.LogFile)
	}

	_ = log.Close()

	self.LogChan = make(chan string, 20)
	self.Done = make(chan bool, 1)

	go self.WriteLogGo()

	//fmt.Println("log init succ.")
	return nil
}

func (self *Logger) close() {
	self.Done <- true
	close(self.LogChan)
}

func (self *Logger) SetLogLevel(level int) {
	if LOG_ERROR <= level && level <= LOG_DEUBG {
		self.RWLock.Lock()
		self.LogLevel = level
		self.RWLock.Unlock()
	}
}

func (self *Logger) GetLogLevel() int {
	self.RWLock.RLock()
	level := self.LogLevel
	self.RWLock.RUnlock()
	return level
}

func (self *Logger) WriteLogGo() {
	for {
		select {
		case <-self.Done:
			close(self.Done)
			return
		case logContent, ok := <-self.LogChan:
			if ok {
				if self.IsPrint {
					fmt.Printf(logContent)
				}

				if self.LogFile == "" {
					continue
				}

				logger, err := os.OpenFile(self.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Printf("open log file[%s] fail.\n", self.LogFile)
					continue
				}

				_, _ = logger.WriteString(logContent)

				_ = logger.Close()
				//
				//if fileinfo, err:= os.Stat(self.LogFile); err == nil {
				//	filesize := fileinfo.Size() / 1024 / 1024
				//	if filesize > 20 {
				//
				//	}
				//
				//}

			}
		default:
			continue
		}
	}
}

func getCaller() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknow", -1
	}

	return file, line
}

func (self *Logger) writeLog(level string, msg string, args ...interface{}) {
	log_time := time.Now().Format("2006-01-02 15:04:05.000")
	log_file_name, log_file_line := getCaller()
	log_file_name_list := strings.Split(log_file_name, "/")
	pathLen := len(log_file_name_list)
	if pathLen > 1 {
		log_file_name = log_file_name_list[pathLen-1]
	}
	log_content := fmt.Sprintf("[%s] %s <%s><%s:%d>\n\n", level, fmt.Sprintf(msg, args...), log_time, log_file_name, log_file_line)
	self.LogChan <- log_content
}

func (self *Logger) Elog(msg string, args ...interface{}) {

	if self.GetLogLevel() < LOG_ERROR {
		return
	}
	self.writeLog("ERROR", msg, args...)
}

func (self *Logger) Wlog(msg string, args ...interface{}) {

	if self.GetLogLevel() < LOG_WARN {
		return
	}
	self.writeLog("WARN ", msg, args...)
}

func (self *Logger) Flog(msg string, args ...interface{}) {

	if self.GetLogLevel() < LOG_FLOW {
		return
	}
	self.writeLog("FLOW ", msg, args...)
}

func (self *Logger) Ilog(msg string, args ...interface{}) {

	if self.GetLogLevel() < LOG_INFO {
		return
	}
	self.writeLog("INFO ", msg, args...)
}

func (self *Logger) Dlog(msg string, args ...interface{}) {

	if self.GetLogLevel() < LOG_DEUBG {
		return
	}
	self.writeLog("DEUBG", msg, args...)
}

func Init_log(loglevel int, isPrint bool) {
	sysType := runtime.GOOS

	if sysType == "linux" {
		LogFileName = "/home/log/webrtcMedia.log"
		LogLIBVideoFileName = "/home/log/webrtcMedia_video.log"
	}

	CreateLogger("main", LogFileName, loglevel, isPrint)
}

func GetLIBVideoLogFile() string {
	return LogLIBVideoFileName
}
