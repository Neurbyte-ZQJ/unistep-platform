# 项目规则

## 环境配置

- GCC 路径：`D:\msys64\mingw64\bin`，运行 Go 测试或编译（CGO）前需先设置 PATH：
  ```powershell
  $env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH
  ```
- 后端技术栈：Go + Gin + GORM + SQLite（需 CGO）
- 前端技术栈：Vue 3 + Element Plus + Pinia + TypeScript + Vite

## 开发规范

- 后端 Handler 直接操作 GORM，未引入 Service 层
- 统一响应格式：`response.OK()` / `response.Created()` / `response.Fail()`
- 文件上传通过 `Uploader` 接口抽象，便于测试替换
- 前端使用 Composition API + `<script setup lang="ts">`
- API 调用封装在 `frontend/src/api/` 目录下
