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

框架提供兩種 parse 方式，依 `Content-Type` header 自動選擇 Codec。

### ParseRequest — 手動處理 error

decode 失敗時回傳 error，由呼叫方自行決定如何回應。

```go
r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    var body CreateUserRequest
    if err := framework.ParseRequest(req, &body); err != nil {
        framework.HandleError(w, req, err)
        return
    }
    // use body ...
})
```

### ParseOrRespond — 自動處理 error

decode 失敗時自動呼叫 `HandleError`（走 ExceptionMapper → `ErrBadRequest` 預設 400 → fallback 500），handler 只需判斷是否 `return`。

```go
r.POST("/api/users/login", func(w http.ResponseWriter, req *http.Request) {
    var body LoginRequest
    if err := framework.ParseOrRespond(w, req, &body); err != nil {
        return  // 回應已由框架寫入，只需停止執行
    }
    // use body ...
})
```

預設支援 `application/json` 與 `text/plain`（見 [🗜️ Codec 擴充](#️-codec-擴充)）。

---

## 📤 Response 序列化

`framework.Respond` 依 `Accept` header 自動選擇 Codec 序列化並設定對應的 response header。

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

依序嘗試以下三層：

1. ⚡ **ExceptionMapperPlugin 自訂規則** — 先查 pointer equality（O(1)），再以 `errors.Is` 遍歷 wrapped error
2. 🛡️ **Framework 預設 mapping** — `framework.ErrBadRequest` → 400 Bad Request（不需要額外設定）
3. 🔥 **Fallback** — 回傳 500 Internal Server Error

```go
r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    if err := userService.Register(body.Email, body.Name, body.Password); err != nil {
        framework.HandleError(w, req, err)   // 自動轉換 err → HTTP status
        return
    }
    framework.Respond(w, req, http.StatusCreated, nil)
})
```

### 🛡️ ErrBadRequest — Framework 預設 sentinel

`framework.ErrBadRequest` 是框架層定義的 sentinel error，代表 request 格式錯誤。
`HandleError` 遇到它會自動回應 400，**不需要在 ExceptionMapperPlugin 額外設定**。

`ParseOrRespond` 在 decode 失敗時就會回傳 `ErrBadRequest`，也可以手動使用：

```go
if someFormatInvalid {
    framework.HandleError(w, r, framework.ErrBadRequest)
    return
}
```

### 🗂️ ExceptionMapperPlugin — 定義業務錯誤映射規則

在組裝路由時安裝插件，一次定義所有業務錯誤的 HTTP 對應。

```go
import "github.com/xchwan/simple-web-framework/framework/plugin"

router.AddPlugin(
    plugin.NewExceptionMapperPlugin().
        On(ErrEmailDuplicate,     http.StatusBadRequest,   "Duplicate email").
        On(ErrCredentialsInvalid, http.StatusBadRequest,   "Credentials invalid").
        On(ErrTokenInvalid,       http.StatusUnauthorized, "Can't authenticate who you are.").
        On(ErrForbidden,          http.StatusForbidden,    "Forbidden"),
)
```

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
    repo := router.Resolve("userRepo").(*UserRepository)
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
router.Bind("userRepo",    func() any { return NewUserRepository() })
router.Bind("userHandler", func() any { return NewUserHandler() })

h := router.Resolve("userHandler").(*UserHandler)
router.GET("/api/users", h.List)
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

Plugin 透過兩個介面擴充框架能力：

```go
// Installer 在 AddPlugin 時執行一次，用於安裝期初始化（例如向 CodecRegistry 註冊 codec）。
type Installer interface {
    Install(ctx PluginContext)
}

// ContextInjector 在每個 request 進來時執行，將資料注入 request context。
type ContextInjector interface {
    Inject(r *http.Request) *http.Request
}
```

一個 plugin 可以只實作其中一個，也可以兩個都實作。

### 安裝 Plugin

```go
router.AddPlugin(myPlugin)
```

- 若 plugin 實作 `Installer` → 立即執行 `Install`，並傳入目前所有已註冊資源（`PluginContext`）
- 若 plugin 實作 `ContextInjector` → 每個 request 進來時自動呼叫 `Inject`

### PluginContext — Plugin 之間的溝通橋樑

`PluginContext` 是一個以型別為 key 的 map，`Install` 時可以從中取出其他已安裝的資源。
這讓 plugin 之間可以互相協作，而 Router 不需要知道任何具體型別。

```go
// XmlCodec 在 Install 時向 CodecRegistry 註冊自己
func (c *XmlCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).Register("application/xml", c)
}
```

### 📦 內建 Plugin

| Plugin | 介面 | 功能 | 預設 |
|--------|------|------|------|
| `CodecRegistry` | `ContextInjector` | JSON + text/plain 序列化，每個 request 注入 context | ✅ 自動安裝 |
| `ExceptionMapperPlugin` | `ContextInjector` | error → HTTP status 映射，每個 request 注入 context | 🔧 手動安裝 |
| `XmlCodec` | `Installer` | 向 CodecRegistry 註冊 application/xml 支援 | 🔧 手動安裝 |

---

## 🗜️ Codec 擴充

### 🗂️ 啟用 XML 支援

框架內建 `XmlCodec`，安裝後即可處理 `application/xml` 請求與回應。

```go
import "github.com/xchwan/simple-web-framework/framework/plugin"

router.AddPlugin(&plugin.XmlCodec{})
```

### 新增自訂 Media Type

實作 `plugin.Codec` 介面，再透過 `Installer` 向 `CodecRegistry` 註冊：

```go
import (
    "io"
    "reflect"
    "github.com/xchwan/simple-web-framework/framework/plugin"
)

type MsgpackCodec struct{}

func (c *MsgpackCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).
        Register("application/msgpack", c)
}

func (c *MsgpackCodec) Encode(w io.Writer, v any) error {
    return msgpack.NewEncoder(w).Encode(v)
}

func (c *MsgpackCodec) Decode(r io.Reader, v any) error {
    return msgpack.NewDecoder(r).Decode(v)
}

// 安裝
router.AddPlugin(&MsgpackCodec{})
```

---

## 📦 完整範例

以下為 `internal/user` 的完整組裝流程，展示框架各功能的協作方式。

### 1. 🐛 定義 Domain Errors

```go
// internal/user/errors.go
var (
    ErrEmailDuplicate        = errors.New("email duplicate")
    ErrRegisterFormatInvalid = errors.New("register format invalid")
    ErrCredentialsInvalid    = errors.New("credentials invalid")
    ErrLoginFormatInvalid    = errors.New("login format invalid")
    ErrTokenInvalid          = errors.New("token invalid")
    ErrForbidden             = errors.New("forbidden")
    ErrNameFormatInvalid     = errors.New("name format invalid")
)
```

### 2. ✍️ 撰寫 Handler

Handler 透過 `framework.Get[T]` 從 container 取得 service，不持有任何依賴。

```go
// internal/user/handler.go
type UserHandler struct{}

func (h *UserHandler) service(r *http.Request) *UserService {
    return framework.Get[*UserService](r, "userService")
}

// Register：body 格式錯誤時讓 service 驗證並回傳 domain error（手動流）
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    framework.ParseRequest(r, &req)  // error 由 service 驗證攔截
    u, err := h.service(r).Register(req.Email, req.Name, req.Password)
    if err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusCreated, userResponse{ID: u.ID, Email: u.Email, Name: u.Name})
}

// Login：body 格式錯誤直接 400，不進 service（自動流）
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := framework.ParseOrRespond(w, r, &req); err != nil {
        return
    }
    u, err := h.service(r).Login(req.Email, req.Password)
    if err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusOK, loginResponse{ID: u.ID, Email: u.Email, Name: u.Name, Token: u.Token})
}
```

### 3. 🔧 組裝路由

```go
// internal/user/register.go
func Register(router *framework.Router) {
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

    router.Bind("userRepo", func() any { return NewUserRepository() })
    router.Bind("userService", func() any {
        repo := router.Resolve("userRepo").(*UserRepository)
        return NewUserService(repo)
    }, scope.NewHttpRequestScope())
    router.Bind("userHandler", func() any { return NewUserHandler() })

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
    user.Register(r)
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
