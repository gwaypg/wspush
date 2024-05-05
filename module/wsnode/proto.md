
# 协议

## 登录

通过http登录服务换取websocket连接的token

### 请求 
* 请求方法:HTTP-POST
* SPEC:限制增加的个数 
* 请求地址: /github.com/gwaypg/wspush/login
* 请求鉴权: 登录
* 请求参数 

```text
acc=昵称/手机号/邮件&pwd=sha256(原始密码)
```
| 字段 | 描述  |
| --- | --- |
| acc |昵称/手机号/邮件 |
| pwd | 密码的sha256值 |

### 响应
```text
200 请求成功
```
| 字段 | 描述  |
| --- | --- |
| 200 | http状态码 | 
| 请求成功 | 状态说明 |

