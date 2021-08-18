package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// Mount Namespace 是 Linux 第一个 Namespace 类型， 因此，它的系统调用参数是 NEWNS (New Namespace 的缩写）。
func main() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}


// [root@shcCDFrh75vm7 4_mount]# go run main.go
// sh-4.2# ls /proc/
// 1      1049   1178   1219  1259  1309  17    1903  1999  2495  2682  2821  2855  31   331  400  53    5967  76   925  9594       diskstats    kcore       mounts        swaps
// 10     10492  1179   1221  1260  1310  1747  1904  2     2531  27    2824  2856  310  337  41   5328  6     77   926  96         dma          keys        mtrr          sys
// 1004   1050   1180   1223  1264  1318  18    1916  2001  2534  2709  2826  2857  311  34   42   54    6003  78   937  9983       driver       key-users   net           sysrq-trigger
// 1023   1056   1181   1224  1265  1324  1866  1923  2003  2540  2772  2827  2860  314  36   43   55    6010  79   938  acpi       execdomains  kmsg        pagetypeinfo  sysvipc
// 10330  1057   1182   1234  1266  1332  1867  1927  2004  2580  2779  2829  2861  315  360  44   56    6021  792  939  buddyinfo  fb           kpagecount  partitions    timer_list
// 10336  1068   1183   1248  127   1339  1870  1971  21    2581  2785  2831  29    316  37   46   5749  6066  8    940  bus        filesystems  kpageflags  sched_debug   timer_stats
// 10351  11     11846  1249  1294  1341  1873  1981  22    2593  2788  2832  2900  317  371  48   5758  63    80   941  cgroups    fs           loadavg     schedstat     tty
// 1037   11093  11922  1250  1295  1366  1875  1983  2242  26    2793  2836  2946  318  38   49   5768  64    81   942  cmdline    interrupts   locks       scsi          uptime
// 10381  11402  12     1253  13    1374  1887  1994  23    2600  28    2843  3     319  389  5    5797  65    82   943  consoles   iomem        mdstat      self          version
// 10429  1163   12016  1255  1300  14    1892  1995  24    2603  2814  2845  30    32   39   50   58    66    9    944  cpuinfo    ioports      meminfo     slabinfo      vmallocinfo
// 10436  1176   12021  1257  1306  1495  1895  1997  2475  2651  2815  2850  308   33   390  51   5880  7     913  945  crypto     irq          misc        softirqs      vmstat
// 10446  1177   12026  1258  1308  16    19    1998  2479  2668  2819  2854  309   330  391  52   5962  74    914  95   devices    kallsyms     modules     stat          zoneinfo
// proc 是一个文件系统，提供额外的机制，可以通过内核和内核模块将信息发送给进程
// 注意这里的 /proc 还是宿主机的
// [root@shcCDFrh75vm7 4_mount]# go run main.go
// sh-4.2# ps -ef // 使用 ps 来查看系统的进程
// Error, do this: mount -t proc proc /proc
// 将 /proc mount 到自己的 Namespace 下面来
// sh-4.2# mount -t proc proc /proc
// sh-4.2# ls /proc/
// 1          bus       cpuinfo    dma          filesystems  ioports   keys        kpageflags  meminfo  mtrr          sched_debug  slabinfo  sys            timer_stats  vmallocinfo
// 5          cgroups   crypto     driver       fs           irq       key-users   loadavg     misc     net           schedstat    softirqs  sysrq-trigger  tty          vmstat
// acpi       cmdline   devices    execdomains  interrupts   kallsyms  kmsg        locks       modules  pagetypeinfo  scsi         stat      sysvipc        uptime       zoneinfo
// buddyinfo  consoles  diskstats  fb           iomem        kcore     kpagecount  mdstat      mounts   partitions    self         swaps     timer_list     version
// sh-4.2# ps -ef
// UID        PID  PPID  C STIME TTY          TIME CMD
// root         1     0  0 13:25 pts/2    00:00:00 sh
// root         6     1  0 14:32 pts/2    00:00:00 ps -ef
// sh 进程是PID 为 1。这就说明，当前的 Mount Namespace 中的 mount 和外部空间是隔离的，mount 操作并没有影响到外部。
// Docker volume 也是利用了这个特性。
