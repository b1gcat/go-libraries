# build

## 开始
```bash
docker run -d --name mysql --env MARIADB_USER=kickYouAbc --env MARIADB_PASSWORD=kickYouAbc \
	--env MARIADB_ROOT_PASSWORD=kickYouAbc  \
	-p 2345:3306  mariadb:latest
	
```

## 进入容器root登陆（密码：kickYouAbc）mysql执行：
```bash
GRANT ALL ON *.* TO 'kickYouAbc'@'%' WITH GRANT OPTION;
flush privileges;
create database unknown default character set utf8 collate utf8_general_ci;
```

## 测试：
远程连接mysql测试联通即可

## 安全性要求

* 防火墙增加白名单访问控制，仅允许代理和指定管理者访问
* 生产环境和测试环境密码不能相同
