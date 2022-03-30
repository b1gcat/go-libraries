# doc
https://rocketmq.apache.org/docs/quick-start/

# 创建mq平台：

## build:
版本号查询 https://archive.apache.org/dist/rocketmq/
git clone https://github.com/apache/rocketmq-docker.git

```bash
cd rocketmq-docker/image-build
sh build-image.sh 4.9.1 centos
sh build-image-dashboard.sh 1.0.0 centos
```

### 生成配置
```
cd rocketmq-docker
export LC_ALL=en_US.UTF-8
sh stage.sh 4.9.1 
```

### 启动

```bash
cd rocketmq-docker/stages/4.9.1/templates
export VERSION=4.9.1
单机模式
修改./play-docker.sh: 变更 sh mqbroker => sh mqbroker -c ../conf/broker.conf
./play-docker.sh centos 
```

### 修改配置

```bash
1）进入broker容器修改../conf/broker.conf文件结尾增加配置
2）进入ns容器修改../conf/play_acl.yml,增加账号信息
3）重启容器broker和ns
```

### 确认服务是否正常

```bash
1）docker ps -a 查看容器状态和端口状态：
nameserver 9876 （存在）
broker 10911/10912 （存在）
2）进入ns容器确认ip是否正确。执行：
sh mqadmin clusterList -n 127.0.0.1:9876
```



### 控制面板

控制面板附属功能可以不启动

```
docker pull apacherocketmq/rocketmq-console:2.0.0
docker run -d --name rocketmq-dashboard -e "JAVA_OPTS=-Drocketmq.namesrv.addr=61.216.34.217:9876 -Drocketmq.config.accessKey=rocketmq2 -Drocketmq.config.secretKey=12345678" -p 6881:8080 -t apacherocketmq/rocketmq-dashboard:latest
访问：
http://x.x.x.x:8080
```



## bug
```bash
使用mqadmin时会报少库错误，则执行命令可修复
cp /usr/lib/jvm/java-1.8.0-openjdk/jre/lib/ext/sunjce_provider.jar \
/home/rocketmq/rocketmq-4.9.1/lib/
```

# 基本操作 
```bash
列出brokers
sh mqadmin clusterList -n 127.0.0.1:9876
列出所有topic
sh mqadmin topicList -n 127.0.0.1:9876
```
