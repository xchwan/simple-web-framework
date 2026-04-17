# 🕸️ Simple Web Framework

以 Go 標準函式庫為基礎打造的輕量 HTTP 框架，示範如何用 **IoC Container**、**Plugin 系統**、**Codec Registry** 等設計模式實作一個可擴充的 Web 框架。

---

## 📋 目錄

- [🚀 快速開始](#-快速開始)
- [🗺️ 路由](#️-路由)
- [📥 Request 解析](#-request-解析)
- [📤 Response 序列化](#-response-序列化)
- [🚨 錯誤處理](#-錯誤處理)
- [🧩 IoC Container 與依賴注入](#-ioc-container-與依賴注入)
- [♻️ Scope（生命週期）](#️-scope生命週期)
- [🔌 Plugin 系統](#-plugin-系統)
- [🗜️ Codec 擴充](#️-codec-擴充)
- [📦 完整範例](#-完整範例)
- [🛠️ 開發指令](#️-開發指令)

---

## 🚀 快速開始

```go
package main

import (
    "net/http"
    "github.com/xchwan/simple-web-framework/framework"
)

func main() {
    r := framework.NewRouter()

    r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
        framework.Respond(w, r, http.StatusOK, map[string]string{"message": "Hello, World!"})
    })

    r.Run(":8080")
}
```

---

## 🗺️ 路由

### 基本路由

支援 `GET`、`POST`、`PUT`、`DELETE`、`PATCH` 五種 HTTP 方法。

```go
r := framework.NewRouter()

r.GET("/api/users", listUsers)
r.POST("/api/users", createUser)
r.PUT("/api/users/{id}", updateUser)
r.PATCH("/api/users/{id}", patchUser)
r.DELETE("/api/users/{id}", deleteUser)
```

### 🔖 Path Parameters

在路徑中以 `{name}` 宣告路徑參數，再用 `framework.PathParam` 取出。

```go
r.GET("/api/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
    id := framework.PathParam(r, "userId")
    // id == "42" (when request path is /api/users/42)
})
```

### 🔍 Query Parameters

直接使用標準函式庫取出 query string。

```go
r.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    keyword := r.URL.Query().Get("keyword")
    // GET /api/users?keyword=alice  =>  keyword == "alice"
})
```

### ⚠️ 路由找不到時的行為

| 狀況 | HTTP 狀態碼 |
|------|------------|
| 路徑不存在 | 404 Not Found |
| 路徑存在但方法不符 | 405 Method Not Allowed |

---

## 📥 Request 解析

`framework.ParseRequest` 依 `Content-Type` header 自動選擇對應的 Codec 反序列化 request body。

```go
type CreateUserRequest struct {
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"password"`
}

r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    var body CreateUserRequest
    if err := framework.ParseRequest(req, &body); err != nil {
        framework.HandleError(w, req, err)
        return
    }
    // use body.Email, body.Name, body.Password ...
})
```

預設支援 `application/json` 與 `text/plain`（見 [🗜️ Codec 擴充](#️-codec-擴充)）。

---

## 📤 Response 序列化

`framework.Respond` 依 `Content-Type` header 自動選擇 Codec 序列化並設定對應的 response header。

```go
// 200 OK with JSON body
framework.Respond(w, r, http.StatusOK, body)

// 201 Created
framework.Respond(w, r, http.StatusCreated, body)

// 204 No Content（不輸出 body）
framework.Respond(w, r, http.StatusNoContent, nil)
```

**錯誤回應格式**：框架統一使用 `framework.ErrorBody` 作為錯誤的 JSON 結構。

```go
// {"message": "something went wrong"}
framework.Respond(w, r, http.StatusBadRequest, framework.Error("something went wrong"))
```

---

## 🚨 錯誤處理

### HandleError — 把 Go error 轉成 HTTP 回應

若有安裝 `ExceptionMapperPlugin`，`HandleError` 會查表轉換成對應的 HTTP status code 與訊息；找不到規則時回傳 500。

```go
r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    if err := userService.Register(body.Email, body.Name, body.Password); err != nil {
        framework.HandleError(w, req, err)   // 自動轉換 err -> HTTP status
        return
    }
    framework.Respond(w, req, http.StatusCreated, nil)
})
```

### 🗂️ ExceptionMapperPlugin — 定義錯誤映射規則

在組裝路由時安裝插件，一次定義所有業務錯誤的 HTTP 對應。

```go
import "github.com/xchwan/simple-web-framework/framework/plugin"

router.AddPlugin(
    plugin.NewExceptionMapperPlugin().
        On(ErrEmailDuplicate,        http.StatusBadRequest,   "Duplicate email").
        On(ErrCredentialsInvalid,    http.StatusBadRequest,   "Credentials invalid").
        On(ErrTokenInvalid,          http.StatusUnauthorized, "Can't authenticate who you are.").
        On(ErrForbidden,             http.StatusForbidden,    "Forbidden"),
)
```

呼叫 `HandleError` 時的查找策略：
1. ⚡ 直接比對 error（pointer equality，O(1)）
2. 🔄 找不到時以 `errors.Is` 遍歷，支援 wrapped error

### 🎨 自訂預設錯誤處理器

覆蓋路由層（404 / 405）的預設回應格式。

```go
router.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "error": http.StatusText(statusCode),
        "path":  r.URL.Path,
    })
})
```

---

## 🧩 IoC Container 與依賴注入

### 注冊依賴

用 `router.Bind` 向容器注冊，省略 scope 時預設為 Singleton。

```go
// 🔒 Singleton（整個應用程式共用同一個 instance）
router.Bind("userRepo", func() any {
    return NewUserRepository()
})

// 🌐 明確指定 scope
router.Bind("userService", func() any {
    repo := framework.Get[*UserRepository](r, "userRepo")
    return NewUserService(repo)
}, scope.NewHttpRequestScope())
```

### 在 Handler 中取出依賴

使用 `framework.Get[T]` 泛型函式，型別安全地取出依賴。

```go
r.GET("/api/users", func(w http.ResponseWriter, req *http.Request) {
    svc := framework.Get[*UserService](req, "userService")
    users := svc.SearchUsers("")
    framework.Respond(w, req, http.StatusOK, users)
})
```

### 🏗️ 啟動時解析依賴

`router.Resolve` 可在路由組裝階段（非 request 期間）取出 Singleton 依賴，用來初始化 handler。

```go
router.Bind("userRepo", func() any { return NewUserRepository() })
router.Bind("userHandler", func() any {
    repo := router.Resolve("userRepo").(*UserRepository)
    return NewUserHandler(repo)
})

handler := router.Resolve("userHandler").(*UserHandler)
router.GET("/api/users", handler.List)
```

---

## ♻️ Scope（生命週期）

| Scope | 說明 | 建立方式 |
|-------|------|----------|
| 🔒 `SingletonScope`（預設）| 整個應用程式只建立一次 | `scope.NewSingletonScope()` |
| 🆕 `PrototypeScope` | 每次 `Resolve` 都建立新 instance | `scope.NewPrototypeScope()` |
| 🌐 `HttpRequestScope` | 同一個 HTTP request 內共用同一個 instance | `scope.NewHttpRequestScope()` |

```go
import "github.com/xchwan/simple-web-framework/framework/scope"

// 每個 request 共享同一個 service instance
router.Bind("userService", func() any {
    return NewUserService()
}, scope.NewHttpRequestScope())

// 每次取都是全新的
router.Bind("tempBuffer", func() any {
    return &bytes.Buffer{}
}, scope.NewPrototypeScope())
```

---

## 🔌 Plugin 系統

Plugin 實作 `plugin.Plugin` 介面的 `Install` 方法，在安裝時向框架注冊能力（例如 Codec）。若需要在每個 request 注入 context，可額外實作 `plugin.RequestPreparer`。

```go
type Plugin interface {
    Install(r Registrar)
}

type RequestPreparer interface {
    PrepareRequest(r *http.Request) *http.Request
}
```

### 安裝 Plugin

```go
router.AddPlugin(myPlugin)
```

框架在每次請求進來時，會自動呼叫所有實作 `RequestPreparer` 的 plugin 的 `PrepareRequest`，讓 plugin 有機會向 context 注入資料。

### 📦 內建 Plugin

| Plugin | 功能 | 預設 |
|--------|------|------|
| `CodecRegistry` | JSON + text/plain 序列化 | ✅ 自動安裝 |
| `ExceptionMapperPlugin` | error → HTTP status 映射 | 🔧 手動安裝 |
| `XmlMediaTypePlugin` | application/xml 支援 | 🔧 手動安裝 |

---

## 🗜️ Codec 擴充

### 新增 Media Type 實作

若只是要替換或新增 codec，用 `router.RegisterCodec` 直接注冊：

```go
router.RegisterCodec("application/msgpack", &MsgpackCodec{})
```

### 透過 Plugin 新增 Media Type

實作 `plugin.Plugin` 介面，在 `Install` 時呼叫 `Registrar.RegisterCodec`：

```go
type MsgpackPlugin struct{}

func (p *MsgpackPlugin) Install(r plugin.Registrar) {
    r.RegisterCodec("application/msgpack", &msgpackCodec{})
}

type msgpackCodec struct{}

func (c *msgpackCodec) Encode(w io.Writer, v any) error {
    return msgpack.NewEncoder(w).Encode(v)
}

func (c *msgpackCodec) Decode(r io.Reader, v any) error {
    return msgpack.NewDecoder(r).Decode(v)
}

// 安裝
router.AddPlugin(&MsgpackPlugin{})
```

### 🗂️ 啟用 XML 支援

框架內建 `XmlMediaTypePlugin`，安裝後即可處理 `application/xml` 請求與回應。

```go
import "github.com/xchwan/simple-web-framework/framework/plugin"

router.AddPlugin(&plugin.XmlMediaTypePlugin{})
```

---

## 📦 完整範例

以下為 `internal/user` 的完整組裝流程，展示框架各功能的協作方式。

### 1. 🐛 定義 Domain Errors

```go
// internal/user/errors.go
var (
    ErrEmailDuplicate         = errors.New("email duplicate")
    ErrRegisterFormatInvalid  = errors.New("register format invalid")
    ErrCredentialsInvalid     = errors.New("credentials invalid")
    ErrLoginFormatInvalid     = errors.New("login format invalid")
    ErrTokenInvalid           = errors.New("token invalid")
    ErrForbidden              = errors.New("forbidden")
    ErrNameFormatInvalid      = errors.New("name format invalid")
)
```

### 2. ✍️ 撰寫 Handler

```go
// internal/user/handler.go
type UserHandler struct {
    service *UserService
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    if err := framework.ParseRequest(r, &req); err != nil {
        framework.HandleError(w, r, ErrRegisterFormatInvalid)
        return
    }
    if err := h.service.Register(req.Email, req.Name, req.Password); err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusCreated, nil)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := framework.ParseRequest(r, &req); err != nil {
        framework.HandleError(w, r, ErrLoginFormatInvalid)
        return
    }
    token, err := h.service.Login(req.Email, req.Password)
    if err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusOK, loginResponse{Token: token})
}
```

### 3. 🔧 組裝路由

```go
// internal/user/register.go
func RegisterRoutes(router *framework.Router) {
    // 安裝錯誤映射插件
    router.AddPlugin(
        plugin.NewExceptionMapperPlugin().
            On(ErrEmailDuplicate,        http.StatusBadRequest,   "Duplicate email").
            On(ErrRegisterFormatInvalid, http.StatusBadRequest,   "Registration's format incorrect.").
            On(ErrCredentialsInvalid,    http.StatusBadRequest,   "Credentials Invalid").
            On(ErrLoginFormatInvalid,    http.StatusBadRequest,   "Login's format incorrect.").
            On(ErrTokenInvalid,          http.StatusUnauthorized, "Can't authenticate who you are.").
            On(ErrForbidden,             http.StatusForbidden,    "Forbidden").
            On(ErrNameFormatInvalid,     http.StatusBadRequest,   "Name's format invalid."),
    )

    // 注冊依賴
    router.Bind("userRepo", func() any { return NewUserRepository() })

    router.Bind("userService", func() any {
        repo := router.Resolve("userRepo").(*UserRepository)
        return NewUserService(repo)
    }, scope.NewHttpRequestScope())

    router.Bind("userHandler", func() any {
        repo := router.Resolve("userRepo").(*UserRepository)
        return NewUserHandler(NewUserService(repo))
    })

    // 注冊路由
    h := router.Resolve("userHandler").(*UserHandler)
    router.POST("/api/users",           h.Register)
    router.POST("/api/users/login",     h.Login)
    router.PATCH("/api/users/{userId}", h.UpdateName)
    router.GET("/api/users",            h.SearchUsers)
}
```

### 4. 🚀 啟動

```go
// cmd/main/main.go
func main() {
    r := framework.NewRouter()
    user.RegisterRoutes(r)
    r.Run(":8080")
}
```

---

## 🛠️ 開發指令

所有指令皆在 Docker 容器內執行，不需本地安裝 Go 環境。

```bash
make all          # ✅ staticcheck + format + test + build（CI 完整流程）
make test         # 🧪 執行 ./test/... 下的整合測試
make build        # 🏗️ 編譯 binary
make staticcheck  # 🔍 靜態分析
make format       # 🎨 gofmt 格式化
make tidy         # 📦 go mod tidy
make shell        # 🐚 進入容器互動 shell
make clean        # 🧹 清除 binary 與 build cache
```
