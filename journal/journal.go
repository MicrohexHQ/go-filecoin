package journal

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Journal defines an interface for recording operations and their metadata
type Journal interface {
	// Record records an operation and its metadata to a Journal accepting variadic key-value
	// pairs.
	Record(operation string, meta ...interface{})
}

// Builder defines a method for creating Journals with a topic.
type Builder func(topic string) Journal

// NewZapJournalBuilder returns a Builder backed by a zap logger.
func NewZapJournalBuilder(filepath string) (Builder, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Encoding = "json"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.LevelKey = ""
	zapCfg.EncoderConfig.CallerKey = ""
	zapCfg.EncoderConfig.MessageKey = "operation"
	zapCfg.EncoderConfig.NameKey = "topic"
	zapCfg.OutputPaths = []string{filepath}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	global, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return func(topic string) Journal {
		return &ZapJournal{
			logger: global.Sugar().Named(topic),
		}
	}, nil
}

// ZapJournal implemented the Journal interface.
type ZapJournal struct {
	logger *zap.SugaredLogger
}

// Record records an operation and its metadata to a Journal accepting variadic key-value
// pairs.
func (zj *ZapJournal) Record(operation string, kv ...interface{}) {
	zj.logger.Infow(operation, kv...)
}
