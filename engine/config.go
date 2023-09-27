package engine

type (
	Config struct {
		Engine EngineFlag
	}
	IConfig interface {
		SetEngineFlags(flags EngineFlag)
		GetEngineFlags() EngineFlag
	}
)

type EngineFlag = uint32

const (
	EngineHotReload EngineFlag = 1 << iota
	EngineSetupAfterReload
	EngineLogging
)

func (c *Config) Setup() error {
	c.Engine = EngineHotReload | EngineLogging
	return nil
}

func (c *Config) SetEngineFlags(flags EngineFlag) {
	c.Engine = flags
}

func (c *Config) GetEngineFlags() EngineFlag {
	return c.Engine
}

var _ IConfig = (*Config)(nil)
