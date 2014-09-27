// Package syslog provides easy to use interface for syslog logging system
package syslog

// #include <stdlib.h>
// #include "syslog_wrapper.h"
import "C"
import (
	"fmt"
	"unsafe"
)

type Priority int
type Option int

const (
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	// Facility.

	// From /usr/include/sys/syslog.h.
	// These are the same up to LOG_FTP on Linux, BSD, and OS X.
	LOG_KERN Priority = iota << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_ // unused
	_ // unused
	_ // unused
	_ // unused
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

const (
	//Option
	LOG_PID    Option = 0x01
	LOG_CONS   Option = 0x02
	LOG_NDELAY Option = 0x08
	LOG_NOWAIT Option = 0x10
	LOG_PERROR Option = 0x20
)

type internal_log struct {
	p   Priority
	msg string
	end bool
}

var log_buff chan internal_log
var doLog bool

// A dedicated async log routine so the logging will be asynchonied via the APP
func run_log() {
	var log_msg internal_log
	test := true
	for test == true {
		log_msg = <-log_buff
		if !log_msg.end {
			write_to_syslog(log_msg)
		} else {
			break
		}
	}
	close(log_buff)
	C.closelog()
}

// Actual syslog writer function this will call the C code
func write_to_syslog(log internal_log) {
	message := C.CString(log.msg)
	C.go_syslog(C.int(log.p), message)
	C.free(unsafe.Pointer(message))
}

// Opens or reopens a connection to Syslog in preparation for submitting messages.
// See http://www.gnu.org/software/libc/manual/html_node/openlog.html
// for parameters description
func Openlog(ident string, o Option, p Priority) {
	cs := C.CString(ident)
	C.go_openlog(cs, C.int(o), C.int(p))
	C.free(unsafe.Pointer(cs))
	log_buff = make(chan internal_log)
	doLog = true
	go run_log()
}

// Writes msg to syslog with facility and priority indicated by parameter "p"
// You can combine facility and priority with bitwise or operator, e.g. :
// syslog.Syslog( syslog.LOG_INFO | syslog.LOG_USER, "Hello syslog")
func Syslog(p Priority, msg string) {
	if doLog {
		log := internal_log{p: p, msg: msg, end: false}
		log_buff <- log
	}
}

// Formats according to a format specifier and writes to syslog with
// facility and priority indicated by parameter "p"
func Syslogf(p Priority, format string, a ...interface{}) {
	Syslog(p, fmt.Sprintf(format, a...))
}

// Closes the current Syslog connection, if there is one.
// This includes closing the /dev/log socket, if it is open.
// Closelog also sets the identification string for Syslog messages back to the default,
func Closelog() {
	doLog = false
	log := internal_log{p: 0, msg: "", end: true}
	log_buff <- log
}

func setlogmask(logmask int) int {
	i := C.setlogmask(C.int(logmask))
	return int(i)
}
