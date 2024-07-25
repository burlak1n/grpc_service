package slogger

import (
	"log/slog"
	"os"

	"context"
	"encoding/json"
	"fmt"
	"grpc_service/internal/entities"
	"io"
	"sync"

	"github.com/fatih/color"
)

// Лучше везде импортировать slog или slogger?
// slogger: его в любом случае придётся импортировать для работы с логгером
// slog: для уменьшения шансов редактирования оригинального slog
// var (
// 	Any = slog.Any
// 	String = slog.String
// )

func SetupLogger(env string) *slog.Logger {
	var h slog.Handler
	switch env {
	case entities.EnvLocal:
		h = NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case entities.EnvDev:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case entities.EnvProd:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(h)
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

type commonHandler struct {
	json              bool // true => output JSON; false => output text
	opts              slog.HandlerOptions
	groups      []string // all groups started from WithGroup
	nOpenGroups int      // the number of groups opened in preformattedAttrs
	mu          *sync.Mutex
	w           io.Writer
}
type handleState struct {
	h       *commonHandler
	buf     *Buffer
	freeBuf bool      // should buf be freed?
	sep     string    // separator to write before next key
	prefix  *Buffer   // for text: key prefix
	groups  *[]string // pool-allocated slice of active groups, for ReplaceAttr
}

var groupPool = sync.Pool{New: func() any {
	s := make([]string, 0, 10)
	return &s
}}

func (h *commonHandler) newHandleState(buf *Buffer, freeBuf bool, sep string) handleState {
	s := handleState{
		h:       h,
		buf:     buf,
		freeBuf: freeBuf,
		sep:     sep,
		prefix:  New(),
	}
	if h.opts.ReplaceAttr != nil {
		s.groups = groupPool.Get().(*[]string)
		*s.groups = append(*s.groups, h.groups[:h.nOpenGroups]...)
	}
	return s
}

func (s *handleState) free() {
	if s.freeBuf {
		s.buf.Free()
	}
	if gs := s.groups; gs != nil {
		*gs = (*gs)[:0]
		groupPool.Put(gs)
	}
	s.prefix.Free()
}

type JSONHandler struct {
	*commonHandler
}

func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions) *JSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &JSONHandler{
		&commonHandler{
			json: true,
			w:    w,
			opts: *opts,
			mu:   &sync.Mutex{},
		},
	}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *JSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.commonHandler.enabled(level)
}

// WithAttrs returns a new [JSONHandler] whose attributes consists
// of h's attributes followed by attrs.
func (h *JSONHandler) WithAttrs(attrs []slog.Attr) (a slog.Handler) {
	return a
}

func (h *JSONHandler) WithGroup(name string) (a slog.Handler) {
	return a
}

func LevelLog(l slog.Level) (level string) {
	level = l.String()[:3] + ":"

	switch l {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}
	return level
}

func (h *commonHandler) Handle(_ context.Context, r slog.Record) error {
	state := h.newHandleState(New(), true, "")
	defer state.free()

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	var (
		timeStr = r.Time.Format("[15:05:05.000]")
		level   = LevelLog(r.Level)
		msg     = color.CyanString(r.Message)
		json    = color.BlackString(string(b))
	)

	fmt.Fprintf(state.buf, "%s %s %s %s\n", timeStr, level, msg, json)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.w.Write(*state.buf)
	return err
}

// enabled reports whether l is greater than or equal to the
// minimum level.
func (h *commonHandler) enabled(l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return l >= minLevel
}
