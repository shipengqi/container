package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
// __attribute__((constructor)) 表示引用包之后会自动执行
__attribute__((constructor)) void enter_namespace(void) {
	char *container_pid;
	container_pid = getenv("container_pid");
	if (container_pid) {
		//fprintf(stdout, "got container_pid=%s\n", container_pid);
	} else {
		//fprintf(stdout, "missing container_pid env skip nsenter");
		return;
	}
	char *container_cmd;
	container_cmd = getenv("container_cmd");
	if (container_cmd) {
		//fprintf(stdout, "got container_cmd=%s\n", container_cmd);
	} else {
		//fprintf(stdout, "missing container_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", container_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
        // 进入 namespace
		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
    // 执行指定的命令
	int res = system(container_cmd);
	exit(0);
	return;
}
*/
import "C"

// 上面的代码带来一个问题，就是只要这个包被导入，它就会在所有 Go 代码前执行，那么即
// 使那些不需要使用 exec 这段代码的地方也会运行这段程序。举例来说，使用 container run 来
// 创建容器，但是这段 C 代码依然会执行，这就会影响前面己经完成的功能。
// 因此需要在这段 C 代码前面一开始的位置就指定环境变量，对于不使用 exec 功能的 Go 代码，只要不设置对应
// 的环境变量，那么当 C 程序检测到没有这个环境变量时，就会直接退出，继续执行原来的代码，
// 并不会影响原来的逻辑。