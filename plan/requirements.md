# 用户需求记录

## 2026-06-09 - Bug 修复需求

### 用户原话

学校服务器今天进行了一次重启，变更了API，然后这个软件就暴露了几个问题：
1. 自动登录器的配置，向导功能的配置，读取出来API无法保存，新配置无法生效
2. 而且只能解析bash格式的CURL，不能解析cmd，要么改成双格式支持，要么指定说明一个格式
3. 默认API配置改成http://210.44.114.32:801/eportal/portal/login?callback=dr1003&login_method=1&user_account=2023405021&user_password=28287X&wlan_user_ip=10.51.243.198&wlan_user_ipv6=&wlan_user_mac=000000000000&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.2.1&terminal_type=1&lang=zh-cn&v=8870&lang=zh
4. 检测间隔无法生效
5. 将检测地址和登录API显示在设置页面，可以进行设置或者是查看

### 提取的需求清单

1. **配置向导API保存修复** - 解析出的自定义登录API必须正确保存到配置文件，并在登录时生效
2. **cURL双格式解析** - 同时支持 bash 格式和 cmd 格式的 cURL 命令解析
3. **更新默认登录API** - 默认API更新为新格式（包含wlan_user_ip等参数）
4. **检测间隔生效修复** - 修改检测间隔后必须立即生效，不需要重启程序
5. **设置页面显示API** - 在设置页面显示当前的认证服务器地址和登录API，可查看和编辑
