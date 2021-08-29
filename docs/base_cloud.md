# OCI 和 CRI

## OCI

Linux 基金会在 2015 年6 月成立了 OCI ( Open Container Initiative ）组织，旨在围绕容器格式定义和运行时的配置制定一个开放的工业化标准。

OCI 是一个标准化的容器规范，包括运行时规范和镜像规范。runC 是基于此规范的一个参考实现，Docker 贡献了 runC 的主要代码。
containerd 比 runC 的层次更高，containerd 可以使用 runC 启动容器，还可以下载镜像，管理网络。

## CRI
CRI 是 Container Runtime Interface ，容器运行时接口的简称。CRI 是一组接口规范，这一插件规范让 Kubernetes 无须重新编译就可以使用更多
的容器运行时。

CRI 包含 Protocol Buffers 、gRPC API 及运行库支持，还有尚在开发的标准规范和工具。

kubelet 通过 gRPC 框架与 CRI shim 进行通信，CRI shim 通过 Unix Socket 启动一个 gRPC server 提供容器运行时服务，kubelet 作为 gRPC client，
通过 Unix Socket 与 CRI shim 通信。gRPC server 使用 protocol buffers 提供两类 gRPC service: ImageService 和 RuntimeService。
ImageService 提供从镜像仓库拉取镜像、删除镜像、查询镜像信息的 RPC 调用功能。
RuntimeService 提供容器的相关生命周期管理（容器创建、修改、销毁等）及容器的交互操作。


CRI 最核心的两个概念就是 PodSandbox 和 container。Pod 由一组应用容器组成，这些应用容器共享相同的环境与资源约束，这个共同的环境与资源约束被
称为 PodSandbox 。由于不同的容器运行时对 PodSandbox 的实现不一样，因此 CRI 留有一组接口给不同的容器运行时自主发挥，例如 Hypervisor 将 PodSandbox 
实现成一个虚拟机，Docker 则将 PodSandbox 实现成一个 Linux 命名空间。

### RuntimeService
Kubelet 在创建一个 Pod 前首先需要调用 RuntimeService。RunPodSandbox 为 Pod 创建一个 PodSandbox ，这个过程包含初始化 Pod 网络、分配 IP 、
激活沙箱等。然后，Kubelet 会调用 CreateContainer、StartContainer、StopContainer、RemoveContainer 对容器进行创建、启动、停止、
删除等操作。当 Pod 被删除时，会调用 StopPodSandbox 、RemovePodSandbox 来销毁 Pod。

Kubelet 的职责在于 Pod 的生命周期的管理，包含健康监测与重启策略控制，井且实现容器生命周期管理中的各种钩子。

### ImageService

为了启动一个容器，CRI 还需要执行镜像相关的操作，比如镜像的拉取、查看、移除等操作，因此 CRI 也定义了一组 ImageService 接口。但是容器的运行不需要
镜像构建操作，所以 CRI 接口并不包含 buildImage 相关操作，镜像的构建需要使用外部工具如 Docker 来完成。

### LogService

LogService 定义了容器的 stdout/stderr 应该如何被 CRI 处理的相关规范。非 stdout/stderr 的日志处理不在 CRI 处理范围之内。