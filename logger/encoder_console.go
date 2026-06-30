package logger

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// ANSI 颜色常量。
const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
	ansiGray    = "\x1b[90m"
)

// globalColorFlag 0 = auto、1 = on、-1 = off；Init() 自动设置。
var globalColorFlag atomic.Int32

func setColorMode(mode string) {
	switch mode {
	case "on":
		globalColorFlag.Store(1)
	case "off":
		globalColorFlag.Store(-1)
	default:
		if isTerminal(os.Stdout) {
			globalColorFlag.Store(1)
		} else {
			globalColorFlag.Store(-1)
		}
	}
}

func useColor() bool { return globalColorFlag.Load() == 1 }

func color(text, c string) string {
	if !useColor() {
		return text
	}
	return c + text + ansiReset
}

// isTerminal 简单判断 stdout 是否 tty。
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func levelColor(l zapcore.Level) string {
	switch l {
	case zapcore.DebugLevel:
		return ansiGray
	case zapcore.InfoLevel:
		return ansiBlue
	case zapcore.WarnLevel:
		return ansiYellow
	case zapcore.ErrorLevel:
		return ansiRed
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ansiMagenta
	}
	return ansiReset
}

// ConsoleEncoder 自研控制台编码器：彩色 + 单行紧凑 + span 缩进。
// 同时实现 zapcore.Encoder 完整接口（Clone / EncodeEntry / ObjectEncoder）。
type ConsoleEncoder struct {
	pool       buffer.Pool
	theme      string
	timeLayout string
}

// NewConsoleEncoder 构造控制台编码器。
func NewConsoleEncoder(theme, timeLayout string) zapcore.Encoder {
	if theme == "" {
		theme = "default"
	}
	if timeLayout == "" {
		timeLayout = "15:04:05.000"
	}
	return &ConsoleEncoder{
		pool:       buffer.NewPool(),
		theme:      theme,
		timeLayout: timeLayout,
	}
}

// Clone 实现 zapcore.Encoder。
func (e *ConsoleEncoder) Clone() zapcore.Encoder {
	return &ConsoleEncoder{
		pool:       buffer.NewPool(),
		theme:      e.theme,
		timeLayout: e.timeLayout,
	}
}

// EncodeEntry 实现 zapcore.Encoder。
//
// 输出顺序：time | level | indent span | msg-or-span | kvs | caller
func (e *ConsoleEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := e.pool.Get()

	buf.AppendString(ent.Time.Format(e.timeLayout))
	buf.AppendByte(' ')

	levelStr := strings.ToUpper(ent.Level.CapitalString())
	switch e.theme {
	case "minimal":
		buf.AppendString(levelStr)
	default:
		buf.AppendString(color(padRight(levelStr, 5), levelColor(ent.Level)))
	}
	buf.AppendByte(' ')

	depth := int32(0)
	spanName := ""
	for _, f := range fields {
		switch f.Key {
		case "span_depth":
			if f.Type == zapcore.Int32Type {
				depth = int32(f.Integer)
			}
		case "span":
			spanName = f.String
		}
	}

	if spanName != "" {
		if depth > 0 {
			writeIndent(buf, depth)
		}
		if useColor() {
			buf.AppendString(color(spanName, ansiCyan+ansiBold))
		} else {
			buf.AppendString(spanName)
		}
	} else {
		writeMessage(buf, ent.Message)
	}

	for _, f := range fields {
		if isFrameKey(f.Key) {
			continue
		}
		buf.AppendByte(' ')
		writeKV(buf, f)
	}

	// caller: 优先取业务注入的 _log_caller field；其次回退 zap 自动填充的 ent.Caller。
	callerShort := ""
	if cf, ok := extractCaller(fields); ok {
		if cv, ok := cf.Interface.(callerFieldValue); ok {
			callerShort = trimCallerPath(cv.file, cv.line)
		}
	} else if ent.Caller.Defined {
		callerShort = trimCallerPath(ent.Caller.File, ent.Caller.Line)
	}
	if callerShort != "" {
		buf.AppendByte(' ')
		if useColor() {
			buf.AppendString(color(callerShort, ansiDim))
		} else {
			buf.AppendString(callerShort)
		}
	}

	if ent.Stack != "" && ent.Level >= zapcore.ErrorLevel {
		buf.AppendByte('\n')
		buf.AppendString(ent.Stack)
	}

	buf.AppendByte('\n')
	return buf, nil
}

// ===== ObjectEncoder 接口实现（保留给 *zap.Field.Set 链式 API 使用） =====
// console encoder 主要工作在 EncodeEntry 上；这里把 Field 调用"吃掉"避免空指针。

func (e *ConsoleEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	// 暂未用到，预留接口
	return nil
}
func (e *ConsoleEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error { return nil }
func (e *ConsoleEncoder) OpenNamespace(key string)                                      {}
func (e *ConsoleEncoder) AddBinary(key string, value []byte)                            {}
func (e *ConsoleEncoder) AddByteString(key string, value []byte)                        {}
func (e *ConsoleEncoder) AddBool(key string, value bool)                                {}
func (e *ConsoleEncoder) AddComplex128(key string, value complex128)                    {}
func (e *ConsoleEncoder) AddComplex64(key string, value complex64)                      {}
func (e *ConsoleEncoder) AddDuration(key string, value time.Duration)                   {}
func (e *ConsoleEncoder) AddFloat64(key string, value float64)                          {}
func (e *ConsoleEncoder) AddFloat32(key string, value float32)                          {}
func (e *ConsoleEncoder) AddInt(key string, value int)                                  {}
func (e *ConsoleEncoder) AddInt64(key string, value int64)                              {}
func (e *ConsoleEncoder) AddInt32(key string, value int32)                              {}
func (e *ConsoleEncoder) AddInt16(key string, value int16)                              {}
func (e *ConsoleEncoder) AddInt8(key string, value int8)                                {}
func (e *ConsoleEncoder) AddString(key, value string)                                   {}
func (e *ConsoleEncoder) AddTime(key string, value time.Time)                           {}
func (e *ConsoleEncoder) AddUint(key string, value uint)                                {}
func (e *ConsoleEncoder) AddUint64(key string, value uint64)                            {}
func (e *ConsoleEncoder) AddUint32(key string, value uint32)                            {}
func (e *ConsoleEncoder) AddUint16(key string, value uint16)                            {}
func (e *ConsoleEncoder) AddUint8(key string, value uint8)                              {}
func (e *ConsoleEncoder) AddUintptr(key string, value uintptr)                          {}
func (e *ConsoleEncoder) AddReflected(key string, value interface{}) error              { return nil }

// ===== helper =====

func isFrameKey(key string) bool {
	switch key {
	case "span", "span_depth", "span_parent", "span_cost_ms", "span_children",
		"trace_id", "user_id", "tag", "cost_ms", callerFieldKey:
		return true
	}
	return false
}

// writeMessage 把 msg 写到 buf；开启颜色时：
//   - 连续匹配 msg 开头的 [...] 标签用 cyan+bold（强提示）；
//   - 剩余正文用 dim grey（不抢戏）。
//   - 关闭颜色时按原文输出。
//
// 例子：
//   msg = "[downstream] 请求 (claude-code → MoGo)"
//   输出 = "\x1b[1;36m[downstream]\x1b[0m \x1b[90m请求 (...)\x1b[0m"
//
// 这让业务前缀一眼可见，同时 msg 主体仍比 INFO/CYAN tag 浅一档。
func writeMessage(buf *buffer.Buffer, msg string) {
	if !useColor() {
		buf.AppendString(msg)
		return
	}
	i := 0
	// 匹配 msg 开头的 [...] 标签，允许连续多个。
	for i < len(msg) {
		if msg[i] == ' ' || msg[i] == '\t' {
			break
		}
		if msg[i] != '[' {
			break
		}
		end := strings.IndexByte(msg[i:], ']')
		if end < 0 {
			break
		}
		tag := msg[i+1 : i+end]
		if tag == "" || strings.TrimSpace(tag) == "" {
			break
		}
		full := msg[i : i+end+1]
		buf.AppendString(color(full, ansiCyan+ansiBold))
		i += end + 1
		// 标签后允许一个空格分隔（连续标签也能匹配 [a][b]）。
		if i < len(msg) && msg[i] == ' ' {
			i++
		}
	}
	rest := msg[i:]
	if rest != "" {
		buf.AppendString(color(rest, ansiGray))
	}
}

func writeIndent(buf *buffer.Buffer, depth int32) {
	for i := int32(0); i < depth; i++ {
		buf.AppendString("  ")
	}
	buf.AppendString("└─ ")
}

func writeKV(buf *buffer.Buffer, f zapcore.Field) {
	buf.AppendString(f.Key)
	buf.AppendByte('=')
	switch f.Type {
	case zapcore.StringType:
		buf.AppendString(color(quoteIfNeeded(f.String), ansiGreen))
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		buf.AppendString(color(strconv.FormatInt(f.Integer, 10), ansiYellow))
	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type, zapcore.UintptrType:
		buf.AppendString(color(strconv.FormatUint(uint64(f.Integer), 10), ansiYellow))
	case zapcore.Float64Type:
		buf.AppendString(color(floatStringFromBits(f.Integer), ansiYellow))
	case zapcore.BoolType:
		val := "false"
		if f.Integer == 1 {
			val = "true"
		}
		buf.AppendString(color(val, ansiCyan))
	case zapcore.ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			buf.AppendString(color("error="+err.Error(), ansiRed))
		}
	case zapcore.DurationType:
		if d, ok := f.Interface.(time.Duration); ok {
			buf.AppendString(color(d.String(), ansiYellow))
		}
	default:
		if b, err := json.Marshal(anyFromField(f)); err == nil {
			_, _ = buf.Write(b)
			return
		}
		buf.AppendString(fmt.Sprintf("%v", anyFromField(f)))
	}
}

// quoteIfNeeded 仅当 string 含空格或 '="' 时加双引号。
func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c <= ' ' || c == '"' || c == '=' {
			return `"` + s + `"`
		}
	}
	return s
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// trimCallerPath 取出 caller 文件的末两段路径。
func trimCallerPath(file string, line int) string {
	idx := strings.LastIndex(file, "/")
	if idx >= 0 {
		file = file[idx+1:]
		if idx2 := strings.LastIndex(file, "/"); idx2 >= 0 {
			// 保留包名 + 文件
			file = file[idx2+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func floatStringFromBits(bits int64) string {
	f := math.Float64frombits(uint64(bits))
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Sprintf("%f", f)
	}
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

func anyFromField(f zapcore.Field) any {
	if f.Interface != nil {
		return f.Interface
	}
	switch f.Type {
	case zapcore.StringType:
		return f.String
	case zapcore.Int64Type:
		return f.Integer
	case zapcore.Uint64Type:
		return uint64(f.Integer)
	case zapcore.BoolType:
		return f.Integer == 1
	case zapcore.ErrorType:
		return f.Interface
	}
	return nil
}
