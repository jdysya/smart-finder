# MD5文件定位系统 API 文档

## 服务端接口

### GET /md5?hash={md5}
返回智能处理页面，自动检测客户端状态并决定处理方式。

**参数:**
- `hash` (string, 必需): 32位MD5哈希值

**响应:** HTML页面

### GET /api/md5?hash={md5}
服务端直接处理MD5查询，重定向到文件管理器。

**参数:**
- `hash` (string, 必需): 32位MD5哈希值

**响应:** 重定向到文件管理器

## 客户端接口

### GET /api/health
健康检查接口，返回客户端状态信息。

**响应:**
```json
{
    "status": "ok",
    "timestamp": 1234567890,
    "version": "1.0.0"
}
```

### GET /md5?hash={md5}
文件定位接口，在文件管理器中定位文件。

**参数:**
- `hash` (string, 必需): 32位MD5哈希值

**响应:** 文本消息

### GET /md5?hash={md5} (带 X-Check-Request: true 头)
文件存在性检查接口，只检查文件是否存在，不执行定位操作。

**参数:**
- `hash` (string, 必需): 32位MD5哈希值
- `X-Check-Request` (header, 必需): "true"

**响应:** HTTP状态码 (200: 存在, 404: 不存在)

## CORS配置

客户端已配置CORS支持：
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, X-Check-Request`
