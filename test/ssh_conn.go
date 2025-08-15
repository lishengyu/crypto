package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSH连接测试函数
func testSSHConnection(server, user, password string, timeout int, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	startTime := time.Now()
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	// 尝试连接
	_, err := ssh.Dial("tcp", server, config)

	time.Sleep(10 * time.Second)

	duration := time.Since(startTime).Milliseconds()
	if err != nil {
		results <- fmt.Sprintf("❌ 连接失败 | 耗时: %dms | 错误: %v", duration, err)
	} else {
		results <- fmt.Sprintf("✅ 连接成功 | 耗时: %dms", duration)
	}
}

func main() {
	// 解析命令行参数
	server := flag.String("server", "", "SSH服务器地址 (格式: host:port)")
	user := flag.String("user", "root", "SSH用户名")
	password := flag.String("password", "", "SSH密码")
	concurrency := flag.Int("c", 1, "并发连接数")
	timeout := flag.Int("t", 5, "超时时间 (秒)")

	flag.Parse()

	// 验证必要参数
	if *server == "" {
		fmt.Println("错误: 必须指定服务器地址")
		flag.Usage()
		os.Exit(1)
	}
	if *password == "" {
		fmt.Println("错误: 必须指定密码")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("\n🚀 开始 SSH 连接测试\n")
	fmt.Printf("├─ 目标服务器: %s\n", *server)
	fmt.Printf("├─ 用户名: %s\n", *user)
	fmt.Printf("├─ 并发数: %d\n", *concurrency)
	fmt.Printf("└─ 超时设置: %d秒\n\n", *timeout)

	// 创建等待组和结果通道
	var wg sync.WaitGroup
	results := make(chan string, *concurrency)

	// 启动并发任务
	startTime := time.Now()
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go testSSHConnection(*server, *user, *password, *timeout, &wg, results)
	}

	// 关闭结果通道的goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集并显示结果
	var success, failure int
	for result := range results {
		if strings.Contains(result, "连接成功") {
			success++
		} else {
			failure++
		}
		fmt.Printf("连接 %d: %s\n", success+failure, result)
	}

	// 显示汇总统计
	totalTime := time.Since(startTime).Seconds()
	fmt.Printf("\n📊 测试结果汇总")
	fmt.Printf("\n├─ 成功连接: %d", success)
	fmt.Printf("\n├─ 失败连接: %d", failure)
	fmt.Printf("\n├─ 总连接数: %d", success+failure)
	fmt.Printf("\n├─ 成功率: %.2f%%", float64(success)*100/float64(*concurrency))
	fmt.Printf("\n├─ 总耗时: %.2f秒", totalTime)
	fmt.Printf("\n└─ 平均连接时长: %.2fms\n\n", float64(totalTime*1000)/float64(*concurrency))
}
