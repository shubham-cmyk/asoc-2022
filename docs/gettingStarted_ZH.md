## 描述
Controller-Design 包含控制器的架构流程。

用户实现
1.无服务器CR
2.供应商的秘密
3.功能码

无服务器 CR 被控制器识别，并创建一个作业以在提到的云提供商上执行 s-deploy。在机密中管理的凭证信息在作业执行期间作为 env 变量加载。

## s-JOB的执行流程
1. 创建 Serverless-CR 时会执行 s-deploy，删除 Serverles-CR 时会执行 s-remove。
2. 执行类型由控制器管理。
3. 一个通用的 Job Created Volumes 会创建 3 个安装在 pod 上的卷。
4. 一个用于控制供应商凭据的托管机密文件。
5. pod 从 docker-registry 下载 Docker 映像，该映像是在 node.js 上构建的，并且无服务器二进制文件是预安装的 init。
6. 当 Pod 执行 s-deploy 时，指示 s 二进制文件执行 s-deploy，反之亦然。
7. 执行 s-deploy 后，pod 将被删除。