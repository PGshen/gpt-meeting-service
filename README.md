<!--
 * @Descripttion: 
 * @version: 
 * @Date: 2023-05-02 21:42:05
 * @LastEditTime: 2023-07-13 00:34:28
-->
# gpt-meeting-service
## 技术选型
- 语言: golang
- 框架: kratos
- 存储: MongoDB

## 建议环境
vscode

## 本地跑起来
```
# 配置configs/config.yaml

# 依赖安装（在Makefile文件中定义了相关命令）
make init

# 运行(调试建议使用vscode)
kratos run

# 导入初始模版
cd cmd/script && go run dataOp.go importData
```

## docker镜像构建
```
docker build -t gpt-meeting-service:v1 .
```

## docker-compose部署
```
cd docker-compose
# 配置文件
mkdir conf && cp ../configs/ ./conf

# 启动
docker-compose up -d

# 停止
docker-compose down
```

## 说明
- 项目采用kratos框架，需要先了解这个 [框架](https://go-kratos.dev/docs/getting-started/start/)
- 由于protoc对sse的不支持，导致了meeting模块没有使用protoc来定义接口
- kratos入门[参考文章](https://learnku.com/articles/64942)
- kratos命令可参考kratos.md

