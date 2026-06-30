// Package mqtt 封装了 MQTT Broker 和消息处理逻辑。
//
// 使用 mochi-mqtt 库实现嵌入式 MQTT Broker，机器人直接连接到本平台的 1883 端口上报数据。
// 通过 Hook 机制拦截所有发布消息，根据 Topic 路径解析并处理不同类型的数据。
//
// 数据流:
//
//	机器人 → MQTT Publish → Broker Hook(OnPublish) → 解析JSON → 更新MySQL → WebSocket推送
package mqtt

import (
	"fmt"
	"log"
	"net"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// Server 是全局 MQTT Broker 实例，InitBroker 成功后可被其他模块引用。
var Server *mqtt.Server

// InitBroker 初始化并启动嵌入式 MQTT Broker。
// 创建 mochi-mqtt Server 实例，添加认证 Hook（当前允许所有连接），
// 配置 TCP 监听器，然后在后台 goroutine 中启动服务。
func InitBroker(port int) error {
	Server = mqtt.New(nil)

	// 添加认证 Hook：当前配置为允许所有连接（生产环境应替换为自定义认证）
	if err := Server.AddHook(new(auth.AllowHook), nil); err != nil {
		return fmt.Errorf("add auth hook: %w", err)
	}

	// 创建 TCP 监听器，监听所有网络接口的指定端口
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp1",
		Address: fmt.Sprintf("0.0.0.0:%d", port),
	})
	if err := Server.AddListener(tcp); err != nil {
		return fmt.Errorf("add tcp listener: %w", err)
	}

	// 在后台 goroutine 中启动 Broker 服务
	log.Printf("[MQTT] Broker starting on port %d", port)
	go func() {
		if err := Server.Serve(); err != nil {
			log.Printf("[MQTT] Broker error: %v", err)
		}
	}()

	return nil
}

// GetTCPPort 使用1883端口创建临时 TCP 监听器，获取系统分配的实际端口号（如果1883被占用），然后关闭监听器并返回端口号。
func GetTCPPort() (int, error) {
	ln, err := net.Listen("tcp", ":1883")
	if err != nil {
		return 1883, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}
