# Windows SSH 仓库权限怎么配置

## Q：为什么 `ssh -T git@github-simpletrack` 会报 `Bad owner or permissions`？

A：这是 Windows OpenSSH 对 `$env:USERPROFILE\.ssh\config` 的权限检查失败，不是 GitHub 仓库不存在，也不是 SSH Host 写错。

OpenSSH 会拒绝读取权限过宽或所有者异常的 SSH 配置文件。当前报错里出现了：

```text
Bad permissions. Try removing permissions for user: DESKTOP-M23J16K\CodexSandboxUsers
Bad owner or permissions on C:\Users\<当前用户>/.ssh/config
```

这说明默认 SSH 配置文件被 `CodexSandboxUsers` 或其他宽泛用户组授予了权限，OpenSSH 认为这个文件可能被其他用户修改，所以拒绝使用。

## Q：第一次配置 `simpletrack` 专用 SSH key 应该怎么做？

A：先生成一个只给 SimpleTrack / `simpletrack` GitHub 组织使用的 SSH key：

```powershell
ssh-keygen -t ed25519 -C "simpletrack" -f "$env:USERPROFILE\.ssh\id_ed25519_simpletrack"
```

然后读取公钥：

```powershell
Get-Content "$env:USERPROFILE\.ssh\id_ed25519_simpletrack.pub"
```

把输出的整行 `ssh-ed25519 ... simpletrack` 添加到 GitHub：

- 如果使用个人账号 `RoggerZ` 推送：添加到 GitHub 个人账号的 `Settings -> SSH and GPG keys`。
- 如果后续使用专门的组织/机器身份：添加到对应账号的 `Settings -> SSH and GPG keys`，或者按 GitHub 仓库策略配置 deploy key。

注意只提交公钥，不要提交或复制私钥 `id_ed25519_simpletrack`。

## Q：专用 SSH Host 应该怎么写？

A：在 `$env:USERPROFILE\.ssh\config_simpletrack` 中维护 Host 别名：

```sshconfig
Host github-simpletrack
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_simpletrack
  IdentitiesOnly yes
```

写好后先验证：

```powershell
ssh -T git@github-simpletrack
```

如果默认 `$env:USERPROFILE\.ssh\config` 权限异常导致这个命令仍然读取失败，可以直接指定配置文件验证：

```powershell
$sshConfig = "$($env:USERPROFILE -replace '\\','/')/.ssh/config_simpletrack"
ssh -F $sshConfig -T git@github-simpletrack
```

成功时会看到类似：

```text
Hi RoggerZ! You've successfully authenticated, but GitHub does not provide shell access.
```

## Q：当前 SimpleTrack 仓库实际用哪套 SSH 配置？

A：当前推荐继续使用独立配置文件：

```text
$env:USERPROFILE\.ssh\config_simpletrack
```

内容是：

```sshconfig
Host github-simpletrack
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_simpletrack
  IdentitiesOnly yes
```

这个配置已经验证可以识别到 GitHub 身份 `RoggerZ`。

## Q：为什么不直接依赖默认的 `~/.ssh/config`？

A：因为当前默认文件 `$env:USERPROFILE\.ssh\config` 的 ACL 已经异常，普通 PowerShell 进程无法读取，也无法用 `takeown` / `icacls` 修复：

```text
C:\Users\<当前用户>\.ssh\config: Access is denied.
```

在这种情况下，仓库级配置 `core.sshCommand` 更稳定，也不会影响其他 GitHub 账号或其他项目。

## Q：两个子仓库应该怎样固定使用 `config_simpletrack`？

A：在两个独立子仓库里写入仓库级 Git 配置：

```powershell
$sshConfig = "$($env:USERPROFILE -replace '\\','/')/.ssh/config_simpletrack"
git -C ".\src\analytics-core" config core.sshCommand "ssh -F $sshConfig"
git -C ".\src\simpletrack-saas" config core.sshCommand "ssh -F $sshConfig"
```

这样即使默认 `~/.ssh/config` 坏了，Git 推送也会绕过它，直接读取 `config_simpletrack`。

## Q：远程地址应该怎么写？

A：两个子仓库远程地址应该使用 `github-simpletrack` 这个 Host 别名：

```powershell
git -C ".\src\analytics-core" remote set-url origin git@github-simpletrack:simpletrack/analytics-core.git
git -C ".\src\simpletrack-saas" remote set-url origin git@github-simpletrack:simpletrack/simpletrack-saas.git
```

注意这里不是 `git@github.com:...`，而是 `git@github-simpletrack:...`。这样 Git 才会命中 `config_simpletrack` 里的专用身份。

## Q：如何验证当前配置是否正确？

A：先验证 SSH 身份：

```powershell
$sshConfig = "$($env:USERPROFILE -replace '\\','/')/.ssh/config_simpletrack"
ssh -F $sshConfig -T git@github-simpletrack
```

成功时会看到类似：

```text
Hi RoggerZ! You've successfully authenticated, but GitHub does not provide shell access.
```

再验证两个仓库的 Git 配置：

```powershell
git -C ".\src\analytics-core" config --get core.sshCommand
git -C ".\src\simpletrack-saas" config --get core.sshCommand

git -C ".\src\analytics-core" remote -v
git -C ".\src\simpletrack-saas" remote -v
```

期望结果：

```text
ssh -F C:/Users/<当前用户>/.ssh/config_simpletrack

origin  git@github-simpletrack:simpletrack/analytics-core.git
origin  git@github-simpletrack:simpletrack/simpletrack-saas.git
```

## Q：为什么我已经配置了访问密钥和代理，Codex 里还是 `Permission denied`？

A：因为这里有三层不同的问题，不能混在一起判断：

1. **GitHub 是否接受 key**：你在普通 PowerShell 里执行下面命令已经成功，说明 key 和 GitHub 授权是通的。

```powershell
ssh -F "$env:USERPROFILE\.ssh\config_simpletrack" -T git@github-simpletrack
```

看到下面结果就代表 GitHub 已接受这个 key：

```text
Hi RoggerZ! You've successfully authenticated, but GitHub does not provide shell access.
```

2. **当前进程能不能读取 SSH config / 私钥**：Codex 工具进程可能不是你的交互式 PowerShell 身份，而是 `desktop-m23j16k\codexsandboxoffline`。所以你手动 PowerShell 能读 `config_simpletrack`，不代表 Codex 沙箱进程也能读。

3. **SSH 是否真的走了 HTTP 代理**：`git config http.proxy` 只影响 HTTPS Git 请求，不会自动影响 `git@github...` 这种 SSH 请求。SSH 要走 HTTP 代理，需要在 SSH config 里配置 `ProxyCommand`，或者在 `GIT_SSH_COMMAND` / `core.sshCommand` 中显式指定。

因此当前判断是：**key 本身没问题；Codex 推送失败主要是当前沙箱读不到 `config_simpletrack` 或 `id_ed25519_simpletrack`，以及 SSH 不会自动吃 HTTP 代理。**

## Q：为什么 `icacls "$env:USERNAME:(R)"` 会报“无效参数”？

A：这是 PowerShell 字符串插值写法导致的。`$env:USERNAME:(R)` 这种写法容易被 PowerShell 或 `icacls` 拆错。正确写法是先把用户名放进变量，再用 `$()` 明确包起来：

```powershell
$me = "$env:USERDOMAIN\$env:USERNAME"

icacls "$env:USERPROFILE\.ssh\config_simpletrack" /grant:r "$($me):(R,W)" "SYSTEM:(F)" "Administrators:(F)"
icacls "$env:USERPROFILE\.ssh\id_ed25519_simpletrack" /grant:r "$($me):(R)" "SYSTEM:(F)" "Administrators:(F)"
```

如果目标是让 Codex 沙箱进程也能推送，需要额外给 Codex 当前进程身份只读权限。当前观察到的身份是：

```text
DESKTOP-M23J16K\codexsandboxoffline
```

可以用下面命令给它最小读权限：

```powershell
$codex = "DESKTOP-M23J16K\codexsandboxoffline"

icacls "$env:USERPROFILE\.ssh\config_simpletrack" /grant:r "$($codex):(R)"
icacls "$env:USERPROFILE\.ssh\id_ed25519_simpletrack" /grant:r "$($codex):(R)"
```

注意不要给 `Users`、`Everyone`、`Authenticated Users` 这类宽泛主体授权；OpenSSH 可能会因为权限过宽拒绝读取配置或私钥。

## Q：如果不想让 Codex 读取私钥，怎么继续推送？

A：可以直接在你已经认证成功的 PowerShell 里推送。当前待推送的子仓提交是：

```text
analytics-core: b820f7e
analytics-service: f60ac59
simpletrack-saas: 461900a
```

如果使用当前工作区，先确认/提交子仓；如果使用本轮生成的临时 clone，可以在 PowerShell 中执行：

```powershell
$base = "C:\Users\admin\Documents\src\uba\.tmp\commit-goal-20260509-092600"

git -C "$base\analytics-core" push origin main
git -C "$base\analytics-service" push origin main
git -C "$base\simpletrack-saas" push origin main
```

推送完成后，再回父仓更新 `src/analytics-core`、`src/analytics-service`、`src/simpletrack-saas` 的 gitlink 和实施进度文档。

## Q：如果仍然想修复默认 `$env:USERPROFILE\.ssh\config`，应该怎么做？

A：需要使用“以管理员身份运行”的 PowerShell 执行 ACL 修复。普通 PowerShell 可能没有权限。

```powershell
$config = "$env:USERPROFILE\.ssh\config"
$me = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name

takeown /F $config

icacls $config /setowner "$me"
icacls $config /inheritance:r

icacls $config /remove:g "DESKTOP-M23J16K\CodexSandboxUsers" "Users" "Authenticated Users" "Everyone"

icacls $config /grant:r "$($me):(R,W)" "SYSTEM:(F)" "Administrators:(F)"
```

修复后再执行：

```powershell
ssh -T git@github-simpletrack
```

## Q：SimpleTrack 项目最终推荐哪种做法？

A：推荐保留仓库级 `core.sshCommand + config_simpletrack`。

原因：

- 它已经验证可用。
- 它不依赖默认 SSH config 的 ACL 状态。
- 它不会影响个人账号、其他私有仓库或其他 GitHub Host。
- 它能让 `analytics-core` 和 `simpletrack-saas` 明确使用 `simpletrack` 组织对应的 SSH key。

默认 `$env:USERPROFILE\.ssh\config` 可以后续有空再修，不应该阻塞 SimpleTrack 当前开发和推送。
