package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"doh-client/config"
)

func main() {
	//加载配置
	config, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return
	}
	log.Printf("Loaded config: %v", config)

	// 初始化日志文件
	logFile, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Log Initialization Failed: > %s", err)
		log.Printf("Failed to open log file: %s", config.LogFilePath)
		log.Printf("Please check the log file path and permissions.")
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.Println("Log Initialization Complete")
	}

	// 监测日志文件大小并旋转
	go func() {
		for {
			time.Sleep(600 * time.Second) // 每10分钟检查一次
			//time.Sleep(10 * time.Second) // 每10秒检查一次 test
			info, err := logFile.Stat()
			if err == nil && info.Size() > config.MaxLogSize {
				if err := rotateLogFile(config.LogFilePath); err != nil {
					log.Printf("Log Rotation Failed: %s", err)
				}
			}
		}
	}()

	// 创建 TCP 服务器
	go func() {
		tcpAddr, err := net.ResolveTCPAddr("tcp", config.DnsAddr)
		if err != nil {
			log.Fatalf("Failed to resolve TCP address: %v", err)
		}

		tcpConn, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Fatalf("Failed to listen on TCP port: %v", err)
		}
		defer tcpConn.Close()

		log.Printf("TCP DNS server started, listening on %s", config.DnsAddr)

		// 处理 TCP 连接
		for {
			conn, err := tcpConn.Accept()
			if err != nil {
				log.Printf("Failed to accept TCP connection: %v", err)
				continue
			}

			go handleTCP(config, conn.(*net.TCPConn))
		}
	}()

	//创建 UDP 服务器(独立于主线程)(预留)
	/*go func() {
		udpAddr, err := net.ResolveUDPAddr("udp", config.DnsAddr)
		if err != nil {
			log.Fatalf("Failed to resolve UDP address: %v", err)
		}

		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Fatalf("Failed to listen on UDP port: %v", err)
		}
		defer udpConn.Close()

		log.Printf("UDP DNS server started, listening on %s", config.DnsAddr)

		// 处理 UDP 请求
		for {
			buf := make([]byte, 512)
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Failed to read from UDP connection: %v", err)
				continue
			}
			go handleUDP(config, udpConn, buf[:n], addr)
		}
	}()*/

	//创建HTTP服务器(健康检查接口)(保留-待用)
	/*http.HandleFunc("/health", healtCheck)
	http.ListenAndServe(":6253", nil)*/

	//创建UDP服务器(于主线程开启)
	addr, err := net.ResolveUDPAddr("udp", config.DnsAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	//使用handleUDP函数处理DNS请求
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port: %v", err)
	}
	defer conn.Close()
	log.Printf("UDP DNS server started, listening on %s", config.DnsAddr)
	for {
		buf := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Failed to read from UDP connection: %v", err)
			continue
		}
		go handleUDP(config, conn, buf[:n], addr)
	}
}

// 发送请求到 DOH 服务器 (简洁版) (保留-待用)
/*func Requst2DOH(config *config.Config, request []byte) ([]byte, error) {
	// 发送请求到 DOH 服务器
	resp, err := http.Post(config.DohServer, "application/dns-message", bytes.NewReader(request))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 读取 DOH 服务器的响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}*/

func Requst2DOH(config *config.Config, request []byte) ([]byte, error) {
	// 创建自定义 HTTP 客户端
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout: 5 * time.Second, // 设置连接超时时间
			}
			// 检查是否是 IPv6 地址，并添加方括号
			if strings.Contains(config.DohServerIP, ":") {
				addr = "[" + config.DohServerIP + "]:443"
			} else {
				addr = config.DohServerIP + ":443"
			}
			return dialer.DialContext(ctx, "tcp", addr)
		},
	}
	// 创建 HTTP 客户端
	client := &http.Client{
		Transport: transport,
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", config.DohServer, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// 设置自定义 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36 DoH-Client/1.0.0")
	req.Header.Set("Content-Type", "application/dns-message")

	// 发送请求到 DOH 服务器
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to DOH server: %v", err)
	}
	defer resp.Body.Close()

	// 读取 DOH 服务器的响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from DOH server: %v", err)
	}

	return respBody, nil
}

// 处理 UDP DNS 请求
func handleUDP(config *config.Config, conn *net.UDPConn, buf []byte, addr *net.UDPAddr) {

	// 记录请求到日志文件
	log.Printf("Received DNS(UDP) request from %s: %x\n", addr, buf)

	//检测请求是否来自监听IP
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		log.Printf("Failed to split host and port: %v", err)
		return
	}
	if host != config.ListenAddr && config.ListenAddr != "0.0.0.0" {
		log.Printf("Rejected DNS request from %s", host)
		return
	}

	// 调用 Requst2DOH 函数，发送请求到 DOH 服务器
	var respBody []byte
	respBody, err = Requst2DOH(config, buf)
	if err != nil {
		log.Printf("Failed to send request to DOH server: %v", err)
		return
	}

	// 记录响应到日志文件
	log.Printf("Received DOH response: %x\n", respBody)

	// 将 DOH 响应转换为 UDP DNS 格式并返回给客户端
	_, err = conn.WriteToUDP(respBody, addr)
	if err != nil {
		log.Printf("Failed to write response to UDP connection: %v", err)
		return
	}
}

func handleTCP(cfg *config.Config, conn *net.TCPConn) {
	defer conn.Close()

	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Printf("Failed to read message length from TCP connection: %v", err)
		return
	}

	msgLen := binary.BigEndian.Uint16(buf)
	message := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, message); err != nil {
		log.Printf("Failed to read message from TCP connection: %v", err)
		return
	}

	// 记录请求到日志文件
	log.Printf("Received DNS(TCP) request from %s: %x\n", conn.RemoteAddr(), message)

	respBody, err := Requst2DOH(cfg, message)
	if err != nil {
		log.Printf("Failed to send request to DOH server: %v", err)
		return
	}

	_, err = sendTCPDNSResponse(conn, respBody)
	if err != nil {
		log.Printf("Failed to write response to TCP connection: %v", err)
	}
}

func sendTCPDNSResponse(conn *net.TCPConn, respBody []byte) (int, error) {
	msgLen := make([]byte, 2)
	binary.BigEndian.PutUint16(msgLen, uint16(len(respBody)))

	tcpDNSResp := append(msgLen, respBody...)

	return conn.Write(tcpDNSResp)
}

// 健康检查接口(预留)
/*func healtCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}*/

// 日志文件滚动
func rotateLogFile(logFilePath string) error {
	// 关闭当前日志文件
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s", logFilePath)
	}
	defer logFile.Close()

	// 创建新的.gz文件
	newLogFilePath := logFilePath + "-" + time.Now().Format("20060102-150405") + ".tar.gz"
	outFile, err := os.Create(newLogFilePath)
	if err != nil {
		return fmt.Errorf("failed to create gz file: %s", newLogFilePath)
	}
	defer outFile.Close()

	// 自定义压缩等级，例如 gzip.BestCompression
	compressionLevel := gzip.BestCompression
	gzWriter, err := gzip.NewWriterLevel(outFile, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gz writer with level: %v", err)
	}
	defer func() {
		if err := gzWriter.Close(); err != nil {
			log.Printf("failed to close gzWriter: %v", err)
		}
	}()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			log.Printf("failed to close tarWriter: %v", err)
		}
	}()

	// 获取日志文件信息
	logFileStat, err := logFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %s", logFilePath)
	}

	// 创建tar头部
	logFileHeader := &tar.Header{
		Name:    filepath.Base(logFilePath),
		Size:    logFileStat.Size(),
		Mode:    0644,
		ModTime: logFileStat.ModTime(),
	}

	if err := tarWriter.WriteHeader(logFileHeader); err != nil {
		return fmt.Errorf("failed to write log file header: %s", logFilePath)
	}

	// 复制日志文件内容到tar文件
	if _, err := io.Copy(tarWriter, logFile); err != nil {
		return fmt.Errorf("failed to copy log file: %s", logFilePath)
	}

	// 清空原日志文件
	if err := os.Truncate(logFilePath, 0); err != nil {
		return fmt.Errorf("failed to truncate log file: %s", logFilePath)
	}

	return nil
}
