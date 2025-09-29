package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/node"
	"time"

	"gopkg.in/yaml.v3"
)

// SubscriptionService 订阅服务
type SubscriptionService struct {
	ticker *time.Ticker
	quit   chan bool
}

// NewSubscriptionService 创建新的订阅服务
func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		quit: make(chan bool),
	}
}

// StartAutoUpdate 启动自动更新
func (s *SubscriptionService) StartAutoUpdate(interval time.Duration) {
	// 默认间隔为30分钟
	if interval == 0 {
		interval = 30 * time.Minute
	}

	s.ticker = time.NewTicker(interval)

	go func() {
		// 启动时立即执行一次更新
		s.updateAllSubscriptions()

		// 定时执行更新
		for {
			select {
			case <-s.ticker.C:
				s.updateAllSubscriptions()
			case <-s.quit:
				s.ticker.Stop()
				return
			}
		}
	}()

	log.Println("订阅自动更新已启动，更新间隔:", interval)
}

// StopAutoUpdate 停止自动更新
func (s *SubscriptionService) StopAutoUpdate() {
	if s.quit != nil {
		close(s.quit)
		log.Println("订阅自动更新已停止")
	}
}

// updateAllSubscriptions 更新所有启用的订阅
func (s *SubscriptionService) updateAllSubscriptions() {
	log.Println("开始更新所有订阅...")

	// 获取所有启用的订阅
	subscriptions, err := models.GetEnabledExternalSubscriptions()
	if err != nil {
		log.Printf("获取订阅列表失败: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		log.Println("没有启用的订阅需要更新")
		return
	}

	// 更新每个订阅
	for _, subscription := range subscriptions {
		// 检查是否需要更新
		if s.shouldUpdate(&subscription) {
			log.Printf("正在更新订阅: %s", subscription.Name)

			nodeCount, err := UpdateSubscriptionNodes(&subscription)
			if err != nil {
				log.Printf("更新订阅 %s 失败: %v", subscription.Name, err)
				continue
			}

			log.Printf("订阅 %s 更新成功，获取到 %d 个节点", subscription.Name, nodeCount)
		}
	}

	log.Println("所有订阅更新完成")
}

// shouldUpdate 检查订阅是否需要更新
func (s *SubscriptionService) shouldUpdate(subscription *models.ExternalSubscription) bool {
	// 如果从未更新过，需要更新
	if subscription.LastUpdate.IsZero() {
		return true
	}

	// 计算距离上次更新的时间
	elapsed := time.Since(subscription.LastUpdate).Seconds()

	// 如果超过了设定的更新间隔，需要更新
	return elapsed >= float64(subscription.UpdateInterval)
}

// ClashConfig 定义Clash配置结构
type ClashConfig struct {
	Proxies []map[string]interface{} `yaml:"proxies"`
}

// UpdateSubscriptionNodes 从订阅链接获取并更新节点
func UpdateSubscriptionNodes(subscription *models.ExternalSubscription) (int, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", subscription.URL, nil)
	if err != nil {
		return 0, err
	}

	// 设置User-Agent
	if subscription.UserAgent != "" {
		req.Header.Set("User-Agent", subscription.UserAgent)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// 解析订阅内容
	content := string(body)
	log.Print("订阅内容长度: ", content)
	// 尝试Base64解码
	if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
		content = string(decoded)
	} else if decoded, err := base64.URLEncoding.DecodeString(content); err == nil {
		content = string(decoded)
	}

	// 清除旧的节点
	if err := subscription.ClearNodes(); err != nil {
		log.Println("清除旧节点失败:", err)
	}

	var nodes []models.Node
	var nodeCount int

	// 首先尝试解析为YAML格式（Clash配置）
	var clashConfig ClashConfig
	if err := yaml.Unmarshal([]byte(content), &clashConfig); err == nil && len(clashConfig.Proxies) > 0 {
		// 成功解析为YAML，处理Clash配置
		for _, proxy := range clashConfig.Proxies {
			// 将原始配置转换为JSON字符串保存
			configJSON, err := json.Marshal(proxy)
			if err != nil {
				log.Printf("序列化配置失败: %v", err)
				continue
			}

			link := convertClashProxyToLink(proxy)
			if link == "" {
				continue
			}

			// 创建节点
			tempNode := &models.Node{
				Link:   link,
				Config: string(configJSON), // 保存原始配置
			}

			// 解码节点名称
			decodedNode, err := decodeNodeName(tempNode)
			if err != nil {
				// 如果解码失败，尝试使用proxy中的名称
				if name, ok := proxy["name"].(string); ok {
					decodedNode = *tempNode
					decodedNode.Name = name
				} else {
					log.Printf("解码节点失败: %s, 错误: %v", link, err)
					continue
				}
			}

			// 添加订阅前缀到节点名称
			if subscription.Name != "" {
				decodedNode.Name = fmt.Sprintf("[%s] %s", subscription.Name, decodedNode.Name)
			}

			// 添加节点到数据库
			if err := decodedNode.Add(); err != nil {
				log.Printf("添加节点失败: %v", err)
				continue
			}

			// 强制使用订阅名称作为分组
			if subscription.Name != "" {
				groupNode := &models.GroupNode{Name: subscription.Name}
				if err := groupNode.Add(); err != nil {
					log.Printf("创建分组失败: %v", err)
				} else {
					if err := groupNode.Ass(&decodedNode); err != nil {
						log.Printf("关联节点到分组失败: %v", err)
					}
				}
			}

			nodes = append(nodes, decodedNode)
			nodeCount++
		}
	} else {
		// 不是YAML格式，按原有方式处理（按行分割）
		lines := strings.Split(content, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// 检查是否是有效的代理链接
			if !isValidProxyLink(line) {
				continue
			}

			// 创建节点
			tempNode := &models.Node{
				Link: line,
			}

			// 解码节点名称
			decodedNode, err := decodeNodeName(tempNode)
			if err != nil {
				log.Printf("解码节点失败: %s, 错误: %v", line, err)
				continue
			}

			// 添加订阅前缀到节点名称
			if subscription.Name != "" {
				decodedNode.Name = fmt.Sprintf("[%s] %s", subscription.Name, decodedNode.Name)
			}

			// 添加节点到数据库
			if err := decodedNode.Add(); err != nil {
				log.Printf("添加节点失败: %v", err)
				continue
			}

			// 强制使用订阅名称作为分组
			if subscription.Name != "" {
				groupNode := &models.GroupNode{Name: subscription.Name}
				if err := groupNode.Add(); err != nil {
					log.Printf("创建分组失败: %v", err)
				} else {
					if err := groupNode.Ass(&decodedNode); err != nil {
						log.Printf("关联节点到分组失败: %v", err)
					}
				}
			}

			nodes = append(nodes, decodedNode)
			nodeCount++
		}
	}

	// 关联节点到订阅
	if len(nodes) > 0 {
		if err := subscription.AssociateNodes(nodes); err != nil {
			log.Printf("关联节点到订阅失败: %v", err)
		}
	}

	// 更新订阅状态
	if err := subscription.UpdateStatus(nodeCount); err != nil {
		log.Printf("更新订阅状态失败: %v", err)
	}

	return nodeCount, nil
}

// convertClashProxyToLink 将Clash代理配置转换为链接
func convertClashProxyToLink(proxy map[string]interface{}) string {
	proxyType, ok := proxy["type"].(string)
	if !ok {
		return ""
	}

	switch proxyType {
	case "ss":
		return convertShadowsocksToLink(proxy)
	case "ssr":
		return convertShadowsocksRToLink(proxy)
	case "vmess":
		return convertVMESSToLink(proxy)
	case "vless":
		return convertVLESSToLink(proxy)
	case "trojan":
		return convertTrojanToLink(proxy)
	case "hysteria", "hysteria2":
		return convertHysteriaToLink(proxy)
	case "tuic":
		return convertTuicToLink(proxy)
	default:
		return ""
	}
}

// convertShadowsocksToLink 将SS配置转换为ss://链接
func convertShadowsocksToLink(proxy map[string]interface{}) string {
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	cipher, _ := proxy["cipher"].(string)
	password, _ := proxy["password"].(string)
	name, _ := proxy["name"].(string)

	if server == "" || port == 0 || cipher == "" || password == "" {
		return ""
	}

	// 构造SS链接
	// ss://base64(cipher:password)@server:port#name
	userInfo := fmt.Sprintf("%s:%s", cipher, password)
	encodedUserInfo := base64.StdEncoding.EncodeToString([]byte(userInfo))

	link := fmt.Sprintf("ss://%s@%s:%d", encodedUserInfo, server, port)
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}

	return link
}

// convertShadowsocksRToLink 将SSR配置转换为ssr://链接
func convertShadowsocksRToLink(proxy map[string]interface{}) string {
	// SSR链接格式: ssr://base64(server:port:protocol:cipher:obfs:base64(password)/?params)
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	cipher, _ := proxy["cipher"].(string)
	password, _ := proxy["password"].(string)
	protocol, _ := proxy["protocol"].(string)
	obfs, _ := proxy["obfs"].(string)
	name, _ := proxy["name"].(string)

	if server == "" || port == 0 || cipher == "" || password == "" {
		return ""
	}

	encodedPassword := base64.URLEncoding.EncodeToString([]byte(password))

	// 构建SSR链接内容
	ssrContent := fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		server, port, protocol, cipher, obfs, encodedPassword)

	if name != "" {
		params := fmt.Sprintf("/?remarks=%s", base64.URLEncoding.EncodeToString([]byte(name)))
		ssrContent += params
	}

	// 整体进行base64编码
	link := "ssr://" + base64.URLEncoding.EncodeToString([]byte(ssrContent))
	return link
}

// convertVMESSToLink 将VMess配置转换为vmess://链接
func convertVMESSToLink(proxy map[string]interface{}) string {
	// VMess链接格式: vmess://base64(json)
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	uuid, _ := proxy["uuid"].(string)
	alterId := 0
	switch a := proxy["alterId"].(type) {
	case float64:
		alterId = int(a)
	case int:
		alterId = a
	}
	cipher, _ := proxy["cipher"].(string)
	name, _ := proxy["name"].(string)
	network, _ := proxy["network"].(string)
	tls, _ := proxy["tls"].(string)

	if server == "" || port == 0 || uuid == "" {
		return ""
	}

	vmessConfig := map[string]interface{}{
		"v":    "2",
		"ps":   name,
		"add":  server,
		"port": port,
		"id":   uuid,
		"aid":  alterId,
		"scy":  cipher,
		"net":  network,
		"tls":  tls,
	}

	// 添加WebSocket配置
	if network == "ws" {
		if wsOpts, ok := proxy["ws-opts"].(map[string]interface{}); ok {
			if path, ok := wsOpts["path"].(string); ok {
				vmessConfig["path"] = path
			}
			if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
				if host, ok := headers["Host"].(string); ok {
					vmessConfig["host"] = host
				}
			}
		}
	}

	jsonBytes, err := json.Marshal(vmessConfig)
	if err != nil {
		return ""
	}

	link := "vmess://" + base64.StdEncoding.EncodeToString(jsonBytes)
	return link
}

// convertVLESSToLink 将VLESS配置转换为vless://链接
func convertVLESSToLink(proxy map[string]interface{}) string {
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	uuid, _ := proxy["uuid"].(string)
	name, _ := proxy["name"].(string)

	if server == "" || port == 0 || uuid == "" {
		return ""
	}

	link := fmt.Sprintf("vless://%s@%s:%d", uuid, server, port)

	// 添加查询参数
	params := url.Values{}
	if encryption, ok := proxy["encryption"].(string); ok {
		params.Set("encryption", encryption)
	}
	if flow, ok := proxy["flow"].(string); ok {
		params.Set("flow", flow)
	}
	if network, ok := proxy["network"].(string); ok {
		params.Set("type", network)
	}

	if len(params) > 0 {
		link += "?" + params.Encode()
	}

	if name != "" {
		link += "#" + url.QueryEscape(name)
	}

	return link
}

// convertTrojanToLink 将Trojan配置转换为trojan://链接
func convertTrojanToLink(proxy map[string]interface{}) string {
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	password, _ := proxy["password"].(string)
	name, _ := proxy["name"].(string)

	if server == "" || port == 0 || password == "" {
		return ""
	}

	link := fmt.Sprintf("trojan://%s@%s:%d", password, server, port)
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}

	return link
}

// convertHysteriaToLink 将Hysteria配置转换为hysteria://链接
func convertHysteriaToLink(proxy map[string]interface{}) string {
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	auth, _ := proxy["auth"].(string)
	name, _ := proxy["name"].(string)
	proxyType, _ := proxy["type"].(string)

	if server == "" || port == 0 {
		return ""
	}

	var link string
	if proxyType == "hysteria2" {
		link = fmt.Sprintf("hy2://%s@%s:%d", auth, server, port)
	} else {
		link = fmt.Sprintf("hy://%s@%s:%d", auth, server, port)
	}

	if name != "" {
		link += "#" + url.QueryEscape(name)
	}

	return link
}

// convertTuicToLink 将TUIC配置转换为tuic://链接
func convertTuicToLink(proxy map[string]interface{}) string {
	server, _ := proxy["server"].(string)
	port := 0
	switch p := proxy["port"].(type) {
	case float64:
		port = int(p)
	case int:
		port = p
	}
	uuid, _ := proxy["uuid"].(string)
	password, _ := proxy["password"].(string)
	name, _ := proxy["name"].(string)

	if server == "" || port == 0 || uuid == "" || password == "" {
		return ""
	}

	link := fmt.Sprintf("tuic://%s:%s@%s:%d", uuid, password, server, port)
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}

	return link
}

// GenerateYAMLConfig 生成YAML配置文件
func GenerateYAMLConfig(nodes []models.Node) (string, error) {
	// 准备代理配置列表
	var proxies []map[string]interface{}

	for _, node := range nodes {
		// 如果有原始配置，直接使用
		if node.Config != "" {
			var proxy map[string]interface{}
			if err := json.Unmarshal([]byte(node.Config), &proxy); err == nil {
				// 使用节点的名称替换代理配置中的name
				if node.Name != "" {
					proxy["name"] = node.Name
				}
				proxies = append(proxies, proxy)
				continue
			}
		}

		// 如果没有原始配置，从链接解析（向后兼容）
		if proxy := parseProxyFromLink(node); proxy != nil {
			// 确保使用节点的名称
			if node.Name != "" {
				proxy["name"] = node.Name
			}
			proxies = append(proxies, proxy)
		}
	}

	// 构建完整的Clash配置
	clashConfig := map[string]interface{}{
		"proxies": proxies,
		"proxy-groups": []map[string]interface{}{
			{
				"name":    "🚀 节点选择",
				"type":    "select",
				"proxies": getProxyNames(proxies),
			},
			{
				"name":     "♻️ 自动选择",
				"type":     "url-test",
				"proxies":  getProxyNames(proxies),
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
			{
				"name":     "🔯 故障转移",
				"type":     "fallback",
				"proxies":  getProxyNames(proxies),
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
			{
				"name":    "🎯 直连",
				"type":    "select",
				"proxies": []string{"DIRECT"},
			},
			{
				"name":    "🛑 拒绝",
				"type":    "select",
				"proxies": []string{"REJECT"},
			},
		},
		"rules": []string{
			"DOMAIN-SUFFIX,local,🎯 直连",
			"IP-CIDR,127.0.0.0/8,🎯 直连",
			"IP-CIDR,172.16.0.0/12,🎯 直连",
			"IP-CIDR,192.168.0.0/16,🎯 直连",
			"IP-CIDR,10.0.0.0/8,🎯 直连",
			"IP-CIDR,100.64.0.0/10,🎯 直连",
			"MATCH,🚀 节点选择",
		},
	}

	// 转换为YAML
	yamlBytes, err := yaml.Marshal(clashConfig)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// getProxyNames 获取代理名称列表
func getProxyNames(proxies []map[string]interface{}) []string {
	var names []string
	for _, proxy := range proxies {
		if name, ok := proxy["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

// parseProxyFromLink 从链接解析代理配置（用于向后兼容）
func parseProxyFromLink(node models.Node) map[string]interface{} {
	u, err := url.Parse(node.Link)
	if err != nil {
		return nil
	}

	// 根据协议类型解析
	switch u.Scheme {
	case "ss":
		// 解析shadowsocks链接
		return parseShadowsocksLink(node.Link, node.Name)
	case "vmess":
		// 解析VMess链接
		return parseVMessLink(node.Link, node.Name)
	case "trojan":
		// 解析Trojan链接
		return parseTrojanLink(node.Link, node.Name)
	// 可以根据需要添加更多协议的解析
	default:
		return nil
	}
}

// parseShadowsocksLink 解析SS链接
func parseShadowsocksLink(link, defaultName string) map[string]interface{} {
	// 这里实现SS链接的反向解析
	// 简化示例，实际需要完整的解析逻辑
	return map[string]interface{}{
		"type": "ss",
		"name": defaultName,
		// 其他字段需要从链接解析
	}
}

// parseVMessLink 解析VMess链接
func parseVMessLink(link, defaultName string) map[string]interface{} {
	// 这里实现VMess链接的反向解析
	return map[string]interface{}{
		"type": "vmess",
		"name": defaultName,
		// 其他字段需要从链接解析
	}
}

// parseTrojanLink 解析Trojan链接
func parseTrojanLink(link, defaultName string) map[string]interface{} {
	// 这里实现Trojan链接的反向解析
	return map[string]interface{}{
		"type": "trojan",
		"name": defaultName,
		// 其他字段需要从链接解析
	}
}

// isValidProxyLink 检查是否是有效的代理链接
func isValidProxyLink(link string) bool {
	validSchemes := []string{
		"ss://", "ssr://", "vmess://", "vless://",
		"trojan://", "hysteria://", "hy://", "hy2://",
		"hysteria2://", "tuic://",
	}

	for _, scheme := range validSchemes {
		if strings.HasPrefix(link, scheme) {
			return true
		}
	}
	return false
}

// decodeNodeName 解码节点名称
func decodeNodeName(nd *models.Node) (models.Node, error) {
	if nd.Name == "" {
		u, err := url.Parse(nd.Link)
		if err != nil {
			log.Println(err)
			return *nd, err
		}
		switch {
		case u.Scheme == "ss":
			ss, err := node.DecodeSSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = ss.Name
		case u.Scheme == "ssr":
			ssr, err := node.DecodeSSRURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = ssr.Qurey.Remarks
		case u.Scheme == "trojan":
			trojan, err := node.DecodeTrojanURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = trojan.Name
		case u.Scheme == "vmess":
			vmess, err := node.DecodeVMESSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = vmess.Ps
		case u.Scheme == "vless":
			vless, err := node.DecodeVLESSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = vless.Name
		case u.Scheme == "hy" || u.Scheme == "hysteria":
			hy, err := node.DecodeHYURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = hy.Name
		case u.Scheme == "hy2" || u.Scheme == "hysteria2":
			hy2, err := node.DecodeHY2URL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = hy2.Name
		case u.Scheme == "tuic":
			tuic, err := node.DecodeTuicURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = tuic.Name
		}
	}
	return *nd, nil
}
