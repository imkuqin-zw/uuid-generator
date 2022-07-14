# uuid-generator 全局唯一ID生成服务

一个开箱及用的全局唯一ID生成器服务

## 特点
* 已支持
    * 实现美团leaf的segment模式
    * 实现美团leaf的snowflake模式，使用etcd管理worker ID，配合运维可实现worker ID的回收

* 即将实现
    * 微信序列号生成模式
