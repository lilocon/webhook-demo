配置中心
======================

## 入门指南

#### 修改.env 文件 自定义自己的配置
```shell script
vi .env
```


#### 编译
```shell script
hack/build.sh
```

#### 构建镜像

```shell script
hack/build.sh
```

#### 部署

```shell script
hack/gen-k8s-resource.sh
kubectl apply -f .run/resource.yaml
```
