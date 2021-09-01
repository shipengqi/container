## 容器后台运行

在 Docker 早期版本，所有的容器 init 进程都是从 docker daemon 这个进程 fork 出来的，这也就会导致一个众所周知的问题，如果 docker daemon 挂掉，
那么所有的容器都会挂掉，这给升级 docker daemon 带来很大的风险。后来，Docker 使用了 containerd ，也就是现在的 runC，实现了即使 daemon 挂掉，
容器进程不挂掉的功能。

容器，其实就是一个进程。当前运行命令的 `./container` 是主进程，容器是被当前 `./container` 进程 fork 出来的子进程。子进程的结束和父进程
的运行是一个异步的过程，即父进程永远不知道子进程到底什么时候结束。如果创建子进程的父进程退出，那么这个子进程就成了没人管的孩子，俗称孤儿进程。
为了避免孤儿进程退出时无法释放所占用的资源而僵死，进程号为 1 的 init 进程就会接受这些孤儿进程。这就是父进程退出而容器进程依然运行的原理。
虽然容器刚开始是由当前运行的 `./container` 进程创建的，但是当 `./container` 进程退出后，容器进程就会被进程号为 1 的 init 进程接管，这
时容器进程还是运行着的，这样就实现了 `./container` 退出、容器不岩掉的功能。

## `logs` 命令实现

```go
cmd := exec.Command("/proc/self/exe", "init")
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
        syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
}
if tty {
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
} else {
    // 将容器进程的标准输出挂载到 /var/run/q.container/<containerId>/container.log 文件中
    dirURL := fmt.Sprintf(DefaultInfoLocation, containerId)
    if utils.IsNotExist(dirURL) {
        if err := os.MkdirAll(dirURL, 0622); err != nil {
            log.Errorf("NewInitProcess mkdir %s error %v", dirURL, err)
            return nil, nil, err
        }
    }
    stdLogFilePath := dirURL + LogFileName
    stdLogFile, err := os.Create(stdLogFilePath)
    if err != nil {
        log.Errorf("NewInitProcess create file %s error %v", stdLogFilePath, err)
        return nil, nil, err
    }
    cmd.Stdout = stdLogFile
}
```

`logs` 命令输出 `/var/run/q.container/<containerId>/container.log` 文件的内容：

```go
func (a *logA) Run() error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, a.containerId)
	logFileLocation := dirURL + container.LogFileName
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		return errors.Errorf("open: %s, %v", logFileLocation, err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Errorf("read: %s, %v", logFileLocation, err)
	}
	_, _ = fmt.Fprint(os.Stdout, string(content))
	return nil
}
```
