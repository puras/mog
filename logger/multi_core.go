package logger

import (
	"errors"
	"sync"
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// multiCore 把多个 zapcore.Core 合并为单个 Core。
type multiCore struct {
	mu       sync.RWMutex
	cores    []*coreEntry
	minLevel zapcore.Level // 缓存所有 sink 的最低 level
}

type coreEntry struct {
	disabled atomic.Bool
	core     zapcore.Core
}

// Multi 把多个 core 组合为单个 core。空入参返回 nop。
func Multi(cores ...zapcore.Core) zapcore.Core {
	if len(cores) == 0 {
		return zapcore.NewNopCore()
	}
	mc := &multiCore{minLevel: zapcore.DebugLevel}
	for _, c := range cores {
		if c == nil {
			continue
		}
		entry := &coreEntry{core: c}
		if lvl, ok := coreMinLevel(c); ok && lvl > mc.minLevel {
			mc.minLevel = lvl
		}
		mc.cores = append(mc.cores, entry)
	}
	if len(mc.cores) == 0 {
		return zapcore.NewNopCore()
	}
	if len(mc.cores) == 1 {
		return mc.cores[0].core
	}
	return mc
}

// coreMinLevel 尝试取 core 的最低 level；private，仅本文件使用。
func coreMinLevel(c zapcore.Core) (zapcore.Level, bool) {
	type levelEnabler interface {
		LevelEnabler() zapcore.Level
	}
	if l, ok := c.(levelEnabler); ok {
		return l.LevelEnabler(), true
	}
	return zapcore.DebugLevel, false
}

// Enabled 实现 zapcore.LevelEnabler。
// 返回 true 即说明至少一个 sink 在该 level 上生效。
func (m *multiCore) Enabled(lvl zapcore.Level) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.cores {
		if m.cores[i].disabled.Load() {
			continue
		}
		if m.cores[i].core.Enabled(lvl) {
			return true
		}
	}
	return false
}

// With 派生带附加字段的 multiCore。
func (m *multiCore) With(fields []zapcore.Field) zapcore.Core {
	m.mu.RLock()
	sinks := m.cores
	m.mu.RUnlock()

	cloned := make([]*coreEntry, len(sinks))
	for i, e := range sinks {
		cloned[i] = &coreEntry{core: e.core.With(fields)}
	}
	return &multiCore{
		cores:    cloned,
		minLevel: m.minLevel,
	}
}

// Check 把 entry 派发给所有启用且 level 满足的 core。
func (m *multiCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	m.mu.RLock()
	sinks := m.cores
	m.mu.RUnlock()

	for i := range sinks {
		if sinks[i].disabled.Load() {
			continue
		}
		if sinks[i].core.Enabled(ent.Level) {
			ce = ce.AddCore(ent, sinks[i].core)
		}
	}
	return ce
}

// Write 向所有 sink 写日志。
func (m *multiCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	m.mu.RLock()
	sinks := m.cores
	m.mu.RUnlock()

	var errs []error
	for i := range sinks {
		if sinks[i].disabled.Load() {
			continue
		}
		if err := sinks[i].core.Write(ent, fields); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Sync 同步所有 sink。
func (m *multiCore) Sync() error {
	m.mu.RLock()
	sinks := m.cores
	m.mu.RUnlock()

	var errs []error
	for i := range sinks {
		if sinks[i].disabled.Load() {
			continue
		}
		if err := sinks[i].core.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// DisableSink 关闭指定顺序的 sink。
func DisableSink(c zapcore.Core, idx int) bool {
	mc, ok := c.(*multiCore)
	if !ok {
		return false
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if idx < 0 || idx >= len(mc.cores) {
		return false
	}
	before := mc.cores[idx].disabled.Load()
	mc.cores[idx].disabled.Store(true)
	return !before
}

// EnableSink 重新启用指定 sink。
func EnableSink(c zapcore.Core, idx int) bool {
	mc, ok := c.(*multiCore)
	if !ok {
		return false
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if idx < 0 || idx >= len(mc.cores) {
		return false
	}
	before := mc.cores[idx].disabled.Load()
	mc.cores[idx].disabled.Store(false)
	return before
}

// SinkCount 返回 multiCore 的 sink 数量。
func SinkCount(c zapcore.Core) int {
	mc, ok := c.(*multiCore)
	if !ok {
		return 0
	}
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.cores)
}
