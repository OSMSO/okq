# okq
--------

### okq = clock + queue
======

### 下一步计划

- [ ] 用户界面
- [ ] Swagger
- [ ] Job落盘
- [ ] 分布式
- [ ] 粒度更小的锁
- [ ] 取消读锁

想想一下,你为客户提供提醒服务.

客户A:每隔30分钟提醒我喝水
客户B:每隔20分钟提醒我看书
客户C:每隔10分钟提醒我上床

你可以这样使用clock

客户A提示
```
curl -X "POST" "http://localhost:3000/api/v1/timers/remind" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "interval": 180,
  "content": {
    "A": "喝水"
  },
  "poptimes": 1,
  "repeat": 0
}'
```

客户B提示
```
curl -X "POST" "http://localhost:3000/api/v1/timers/remind" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "interval": 120,
  "content": {
    "B": "看书"
  },
  "poptimes": 1,
  "repeat": 0
}'
```

客户C提示
```
curl -X "POST" "http://localhost:3000/api/v1/timers/remind" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "interval": 120,
  "content": {
    "C": "上床"
  },
  "poptimes": 1,
  "repeat": 0
}'
```

然后,当时间到了之后,在nsq队列中,就会有相应的任务生成
你可以通过查询

默认,是将数据保存到内存中,这样的模式只供测试服务使用.
如果在生产环境中使用,请将config.json
    UseDB设置为true