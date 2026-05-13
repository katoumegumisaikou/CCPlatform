package mqtt

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/mochi-mqtt/server/v2"
)

// Publisher 是 MQTT 下行消息发布器，用于向机器人发送控制指令和配置。
// 通过 Broker 的 Publish 方法直接发布消息到指定 Topic。
type Publisher struct {
	server *mqtt.Server // MQTT Broker 实例
}

// NewPublisher 创建 Publisher 实例。
func NewPublisher(server *mqtt.Server) *Publisher {
	return &Publisher{server: server}
}

// SendCommand 下发控制指令到指定机器人。
// Topic: tdw/robot/{robotID}/cmd
// 指令示例: start(启动清扫), stop(停止), return(回仓), reset(复位)
func (p *Publisher) SendCommand(robotID, cmd string, params map[string]interface{}) error {
	topic := fmt.Sprintf("tdw/robot/%s/cmd", robotID)
	payload := map[string]interface{}{
		"cmd":    cmd,
		"params": params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal cmd: %w", err)
	}

	if err := p.server.Publish(topic, data, false, 0); err != nil {
		return fmt.Errorf("publish cmd: %w", err)
	}
	log.Printf("[MQTT] Published cmd to %s: %s", topic, string(data))
	return nil
}

// SendConfig 下发配置参数到指定机器人。
// Topic: tdw/robot/{robotID}/config
// 用于远程更新机器人的工作参数、清扫模式等配置。
func (p *Publisher) SendConfig(robotID string, cfg map[string]interface{}) error {
	topic := fmt.Sprintf("tdw/robot/%s/config", robotID)
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := p.server.Publish(topic, data, false, 0); err != nil {
		return fmt.Errorf("publish config: %w", err)
	}
	log.Printf("[MQTT] Published config to %s: %s", topic, string(data))
	return nil
}
