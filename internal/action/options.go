package action

type RunActionOptions struct {
	Interactive bool
	TTY         bool
	Detach      bool
	CpuSet      string
	MemoryLimit string
	CpuShare    string
	Volume      string
}
