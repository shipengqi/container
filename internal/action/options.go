package action

type RunActionOptions struct {
	Interactive bool
	TTY         bool
	Detach      bool
	CpuSet      string
	MemoryLimit string
	CpuShare    string
	Volume      string
	Name        string
	Network     string
	Publish     []string
	Envs        []string
}

type NetWorkCreateActionOptions struct {
	Subnet string
	Driver string
}
