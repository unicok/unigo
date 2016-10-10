# auth(鉴权) 

## 设计理念
用户中心，支持各种第三方登陆。

## 凭证规范
### uuid
<pre>
二进制uuid, eg: CA761232-ED42-11CE-BACD-00AA0057B223 
</pre>
### plain
<pre>
{
	"username":1,
	"password":2
}
</pre>