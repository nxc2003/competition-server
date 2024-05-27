# 后端功能测试
## 登录功能测试
### 验证码接收功能
```bash
    http://localhost:3000/auth/captcha
```
### 登录功能
```bash
    http://localhost:3000/auth/captcha
    cookie=captchaAnswer=captcha; Path=/;
```
注意系统初始账号密码```admin/123```
```json
{
"account": "admin",
"password": "123",
"identity": "student",
"code": "03795"
}
```
## 用户操作功能
### 添加用户
