package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/1340691923/xwl_bi/cmd/sinker/internal/runner"
)

var (
	configFileDir  string
	configFileName string
	configFileExt  string
)

func init() {
	flag.StringVar(&configFileDir, "configFileDir", "config", "配置文件目录")
	flag.StringVar(&configFileName, "configFileName", "config", "配置文件名")
	flag.StringVar(&configFileExt, "configFileExt", "json", "配置文件扩展名")
	flag.Usage = printMainUsage
}

// main 只保留 sinker 入口级结构：
// 1. 定义命令行参数
// 2. 解析 protect / diagnostic 子命令
// 3. 把真正的运行时装配交给 internal/runner
func main() {
	if handled, exitCode := maybeRunProtectCLI(os.Args[1:]); handled {
		os.Exit(exitCode)
	}
	if handled, exitCode := maybeRunDiagnosticCLI(os.Args[1:]); handled {
		os.Exit(exitCode)
	}
	flag.Parse()
	runner.Run(configFileDir, configFileName, configFileExt)
}

func printMainUsage() {
	output := flag.CommandLine.Output()
	_, _ = fmt.Fprintln(output, "sinker 用法:")
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintln(output, "1. 服务模式")
	_, _ = fmt.Fprintln(output, "   sinker -configFileDir config -configFileName config -configFileExt json")
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintln(output, "2. diagnostic CLI")
	_, _ = fmt.Fprintln(output, "   sinker diagnostic --help")
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintln(output, "3. protect CLI")
	_, _ = fmt.Fprintln(output, "   sinker protect --help")
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintln(output, "服务模式参数:")
	flag.PrintDefaults()
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprint(output, diagnosticHelpText())
	_, _ = fmt.Fprint(output, protectHelpText())
}
