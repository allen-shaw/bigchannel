# TestCase

## 1. 空启动场景
1. producer和consumer同时启动
2. producer先发送
    - producer 先发送后consumer启动
    - 随后producer也发送
3. consumer先消费
    - consumer先启动后producer再发送

## 2. 消息累积场景

## 3. NoAck场景
1. 发送10条消费，ack前5条，
2. 重连，后5条重复（按序）

