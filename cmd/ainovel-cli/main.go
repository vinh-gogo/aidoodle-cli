package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/entry/headless"
	"github.com/voocel/ainovel-cli/internal/entry/startup"
	"github.com/voocel/ainovel-cli/internal/entry/tui"
	"github.com/voocel/ainovel-cli/internal/eval"
	"github.com/voocel/ainovel-cli/internal/host/trend"
	"github.com/voocel/ainovel-cli/internal/rules"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// headlessMode 记录本次是否 headless 启动，供 die 决定错误退出时是否暂停。
var headlessMode bool

func main() {
	// 子命令在常规 flag 解析之前拦截：eval 是离线评测 harness，参数体系独立。
	if len(os.Args) > 1 && os.Args[1] == "eval" {
		os.Exit(eval.Command(os.Args[2:]))
	}

	opts, args, err := parseCLIOptions(os.Args[1:])
	if err != nil {
		die("flags: %v", err)
	}
	if opts.Version {
		buildversion.Print(os.Stdout, versionInfo())
		return
	}
	if opts.Update {
		if err := runSelfUpdate(opts.UpdateVersion); err != nil {
			fmt.Fprintf(os.Stderr, "update: %v\n", err)
			os.Exit(1)
		}
		return
	}
	headlessMode = opts.Headless

	// 首次引导
	if bootstrap.NeedsSetup() {
		if opts.Headless {
			die("error: chế độ headless không hỗ trợ hướng dẫn khởi tạo lần đầu, vui lòng chạy giao diện TUI một lần để hoàn tất cấu hình")
		}
		setupCfg, err := bootstrap.RunSetup()
		if err != nil {
			die("setup: %v", err)
		}
		// 引导完成后使用生成的配置继续
		runWithConfig(setupCfg, opts, args)
		return
	}

	// 加载配置
	cfg, err := bootstrap.LoadConfig()
	if err != nil {
		die("config: %v", err)
	}

	runWithConfig(cfg, opts, args)
}

// die 统一处理致命错误退出：打印到 stderr、落盘到 ~/.ainovel/last-error.log，
// 并在交互式终端（非 headless）下暂停等待回车——双击启动时控制台会随进程退出
// 立即关闭，不暂停的话错误一闪而过，正是 issue #37 里用户无从排查的根因。
func die(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, msg)
	if path := bootstrap.WriteStartupError(msg); path != "" {
		fmt.Fprintf(os.Stderr, "(Chi tiết lỗi đã được ghi vào %s)\n", path)
	}
	if !headlessMode && stdinIsTerminal() {
		fmt.Fprint(os.Stderr, "\nNhấn phím Enter để thoát...")
		fmt.Fscanln(os.Stdin)
	}
	os.Exit(1)
}

// stdinIsTerminal 判断标准输入是否连接到终端（字符设备）。双击启动 / 交互式终端
// 为 true；管道、重定向、CI 为 false。零依赖近似，足够区分要不要暂停。
func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func runWithConfig(cfg bootstrap.Config, opts cliOptions, args []string) {
	rules.EnsureHomeRulesDir()

	if len(args) > 0 {
		die("error: không còn hỗ trợ truyền trực tiếp nhu cầu tiểu thuyết qua dòng lệnh, vui lòng nhập trong ô nhập liệu của TUI sau khi khởi động")
	}

	if opts.Style != "" {
		cfg.Style = opts.Style
	}
	if opts.Dir != "" {
		cfg.OutputDir = opts.Dir
	}

	// FillDefaults 必须先于资产加载:OutputDir 是运行时字段,默认值在此归一——
	// 否则默认配置下 <书目录>/style/ 的本书级文风覆盖永远不会被加载。
	cfg.FillDefaults()
	bundle := assets.Load(cfg.Style, assets.DefaultLoadOptions(cfg.OutputDir))
	if _, ok := bundle.Styles[cfg.Style]; !ok {
		die("error: không tìm thấy phong cách %q. Các phong cách khả dụng: %s", cfg.Style, availableStyles(bundle.Styles))
	}
	if opts.Trends {
		fmt.Println("Đang thu thập các xu hướng tin tức hot (trends)...")
		snap, err := trend.RunIntake(context.Background(), cfg, cfg.OutputDir, nil)
		if err != nil {
			die("lỗi thu thập xu hướng: %v", err)
		}
		fmt.Printf("Đã thu thập %d xu hướng vào %s/meta/trends/\n", len(snap.Items), cfg.OutputDir)
		for i, it := range snap.Items {
			if i >= 10 {
				fmt.Printf("... và %d xu hướng khác.\n", len(snap.Items)-10)
				break
			}
			traffic := ""
			if it.Traffic != "" {
				traffic = fmt.Sprintf(" [%s]", it.Traffic)
			}
			fmt.Printf("  %d. %s%s (%s)\n", i+1, it.Title, traffic, it.Source)
		}
		if opts.Prompt == "" && opts.PromptFile == "" && !opts.Headless {
			return
		}
	}

	if opts.Review {
		cfg.AdvanceMode = "review"
	}

	if opts.Headless {
		if opts.Next {
			if err := headless.Run(cfg, bundle, headless.Options{Next: true}); err != nil {
				die("error: %v", err)
			}
			return
		}
		prompt, err := loadPrompt(opts)
		if err != nil {
			die("error: %v", err)
		}
		if prompt == "" && opts.Trends {
			prompt = "Viết series video TikTok doodle explainer các chủ đề hot xu hướng mới nhất bằng hình tượng người que thời đồ đá."
		}
		if err := headless.Run(cfg, bundle, headless.Options{Prompt: prompt}); err != nil {
			die("error: %v", err)
		}
		return
	}
	if opts.Prompt != "" || opts.PromptFile != "" {
		die("error: --prompt/--prompt-file chỉ có thể sử dụng trong chế độ --headless")
	}
	if err := tui.Run(cfg, bundle, versionInfo()); err != nil {
		die("error: %v", err)
	}
}

func availableStyles(m map[string]string) string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

type cliOptions struct {
	Headless      bool
	Trends        bool
	Review        bool
	Next          bool
	Prompt        string
	PromptFile    string
	Version       bool
	Update        bool
	UpdateVersion string
	Style         string
	Dir           string
}

// parseCLIOptions 提取 CLI flag，返回选项和剩余参数。
func parseCLIOptions(argv []string) (cliOptions, []string, error) {
	var opts cliOptions
	var args []string
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "--version", "-v":
			opts.Version = true
		case "version":
			if i+1 < len(argv) {
				return opts, nil, fmt.Errorf("version không nhận tham số")
			}
			opts.Version = true
		case "update":
			if opts.Update {
				return opts, nil, fmt.Errorf("update chỉ có thể chỉ định một lần")
			}
			opts.Update = true
			if i+1 < len(argv) {
				if strings.HasPrefix(argv[i+1], "-") {
					return opts, nil, fmt.Errorf("update chỉ nhận một tham số phiên bản tùy chọn")
				}
				opts.UpdateVersion = argv[i+1]
				i++
			}
			if i+1 < len(argv) {
				return opts, nil, fmt.Errorf("update chỉ nhận một tham số phiên bản tùy chọn")
			}
		case "--headless":
			opts.Headless = true
		case "--prompt":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("--prompt thiếu giá trị")
			}
			opts.Prompt = argv[i+1]
			i++
		case "--prompt-file":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("--prompt-file thiếu giá trị")
			}
			opts.PromptFile = argv[i+1]
			i++
		case "--style", "-s":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("--style thiếu giá trị")
			}
			opts.Style = argv[i+1]
			i++
		case "--dir", "-d":
			if i+1 >= len(argv) {
				return opts, nil, fmt.Errorf("--dir thiếu giá trị")
			}
			opts.Dir = argv[i+1]
			i++
		case "--trends":
			opts.Trends = true
		case "--review":
			opts.Review = true
		case "--next":
			opts.Next = true
		default:
			args = append(args, argv[i])
		}
	}
	if opts.Next && !opts.Headless {
		return opts, nil, fmt.Errorf("--next chỉ có thể sử dụng trong chế độ --headless (trong TUI vui lòng dùng lệnh /next)")
	}
	if opts.Prompt != "" && opts.PromptFile != "" {
		return opts, nil, fmt.Errorf("--prompt và --prompt-file không thể dùng đồng thời")
	}
	if opts.Version && (opts.Update || opts.Headless || opts.Prompt != "" || opts.PromptFile != "" || len(args) > 0) {
		return opts, nil, fmt.Errorf("version không thể dùng chung với các tham số khởi động khác")
	}
	if opts.Update && (opts.Headless || opts.Prompt != "" || opts.PromptFile != "" || len(args) > 0) {
		return opts, nil, fmt.Errorf("update không thể dùng chung với các tham số khởi động khác")
	}
	return opts, args, nil
}

func versionInfo() buildversion.Info {
	return buildversion.Resolve(buildversion.Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}

func runSelfUpdate(target string) error {
	info := versionInfo()
	result, err := buildversion.Update(context.Background(), buildversion.UpdateOptions{
		Repo:           buildversion.DefaultRepo,
		BinaryName:     "ainovel-cli",
		TargetVersion:  target,
		CurrentVersion: info.Version,
	})
	if err != nil {
		return err
	}
	if !result.Updated {
		fmt.Printf("ainovel-cli đã là phiên bản mới nhất %s\n", result.Version)
		return nil
	}
	fmt.Printf("ainovel-cli đã được cập nhật lên %s\n", result.Version)
	fmt.Printf("Vị trí cài đặt: %s\n", result.Path)
	return nil
}

func loadPrompt(opts cliOptions) (string, error) {
	return loadPromptFrom(opts, os.Stdin)
}

func loadPromptFrom(opts cliOptions, stdin io.Reader) (string, error) {
	if opts.PromptFile == "" {
		return strings.TrimSpace(opts.Prompt), nil
	}

	if opts.PromptFile == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("đọc prompt thất bại: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return startup.LoadPromptFile(opts.PromptFile)
}
