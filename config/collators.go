package config

type CollatorsPoolConfig struct {
	Address       string
	HistoryRounds uint32
	Revokes       bool
}

func DefaultCollatorsPoolConfig() CollatorsPoolConfig {
	return CollatorsPoolConfig{
		Address:       "",
		HistoryRounds: 8,
		Revokes:       true,
	}
}
