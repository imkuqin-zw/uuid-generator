# uuid-generator 全局唯一ID生成服务

全局唯一ID生成器服务

## 特点
* 已支持
    * 美团leaf的segment模式
    * 美团leaf的snowflake模式，使用etcd管理worker ID，配合运维可实现worker ID的回收
    * 微信序列号模式,已实现基本的alloc服务和grpc客户端,store借助etcd实现
* 待实现
    * 微信序列号仲裁服务
