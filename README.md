🔎 Domain-Exporter

Domain Exporter 是一款 Prometheus Exporter，用于检测指定domain的到期时间。

🎯 目标

我们的目标是构建一个可以处理有可能到期的 domain 的 Prometheus Exporter，可以检测到哪些domain已经到期，从而提醒我们及时更新domain。

📈 功能

Domain Exporter 提供了以下功能：

通过 Whois API 自动检测 domain 到期时间
支持多个 domain 监控
支持 Prometheus 远程存储

🏁 如何使用

下载源码：git clone <https://github.com/zhuima/domain-exporter.git>
安装依赖：go mod tidy
启动服务：go run main.go

🤝 贡献

我们欢迎任何形式的贡献，比如 bug 报告，新功能提议或者代码贡献等，你可以 fork 这个项目，然后提交 pull request 来提交你的贡献。

📃 License

[MIT License](https://github.com/xxx/domain-exporter/blob/master/LICENSE)
