# GitHub SSH日常工作流

本文档用于记录本项目后续最常用、最稳妥的 GitHub（SSH）日常操作流程。

## 适用前提

- 已完成 Git 初始化
- 已配置 SSH 密钥并通过 `ssh -T git@github.com` 验证
- 当前仓库已关联远程 `origin`

## 日常标准流程（推荐顺序）

### 1）同步远程最新代码

```bash
git pull --rebase
```

### 2）查看当前变更

```bash
git status
```

### 3）添加要提交的文件

```bash
git add .
```

更稳妥的做法（推荐）：

```bash
git add 路径/文件名
```

### 4）提交变更

```bash
git commit -m "feat: xxx"
```

常见提交前缀：

- `feat`：新功能
- `fix`：修复问题
- `refactor`：重构
- `docs`：文档更新
- `chore`：杂项维护

### 5）推送到 GitHub

```bash
git push
```

## 每次提交前的 3 个检查（强烈建议）

### 检查 1：确认工作区状态

```bash
git status
```

用途：避免把不应提交的文件带入本次提交。

### 检查 2：确认暂存区内容

```bash
git diff --staged
```

用途：确认“即将提交”的内容是否准确。

### 检查 3：查看最近提交风格

```bash
git log --oneline -n 5
```

用途：保持提交信息风格一致，便于团队协作与回溯。

## 常见问题排查

### 推送被拒绝（non-fast-forward）

先同步后再推送：

```bash
git pull --rebase
git push
```

### 提示有冲突

1. 按提示编辑冲突文件并解决冲突  
2. 标记已解决文件：`git add 路径/文件名`  
3. 继续 rebase：`git rebase --continue`  
4. 完成后推送：`git push`

## 快速复制版（最小闭环）

```bash
git pull --rebase
git status
git add .
git commit -m "feat: xxx"
git push
```

