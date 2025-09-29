package api

import (
	"fmt"
	"strconv"
	"sublink/models"
	"sublink/services"

	"github.com/gin-gonic/gin"
)

// 添加外部订阅
func ExternalSubscriptionAdd(c *gin.Context) {
	var req struct {
		Name           string `json:"Name"`
		URL            string `json:"URL"`
		Enabled        bool   `json:"Enabled"`
		UpdateInterval int    `json:"UpdateInterval"`
		GroupName      string `json:"GroupName"`
		UserAgent      string `json:"UserAgent"`
	}

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	if req.Name == "" || req.URL == "" {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "订阅名称和链接不能为空",
		})
		return
	}

	// 设置默认值
	if req.UpdateInterval == 0 {
		req.UpdateInterval = 3600
	}
	if req.UserAgent == "" {
		req.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	subscription := &models.ExternalSubscription{
		Name:           req.Name,
		URL:            req.URL,
		Enabled:        req.Enabled,
		UpdateInterval: req.UpdateInterval,
		GroupName:      req.GroupName,
		UserAgent:      req.UserAgent,
	}

	if err := subscription.Add(); err != nil {
		c.JSON(400, gin.H{
			"code": "40002",
			"msg":  fmt.Sprintf("添加订阅失败: %s", err.Error()),
		})
		return
	}

	// 立即尝试更新订阅内容
	go services.UpdateSubscriptionNodes(subscription)

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "添加订阅成功",
		"data": subscription,
	})
}

// 更新外部订阅
func ExternalSubscriptionUpdate(c *gin.Context) {
	var req struct {
		ID             int    `json:"ID"`
		Name           string `json:"Name"`
		URL            string `json:"URL"`
		Enabled        bool   `json:"Enabled"`
		UpdateInterval int    `json:"UpdateInterval"`
		GroupName      string `json:"GroupName"`
		UserAgent      string `json:"UserAgent"`
	}

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	if req.ID == 0 {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "订阅ID不能为空",
		})
		return
	}

	// 设置默认值
	if req.UpdateInterval == 0 {
		req.UpdateInterval = 3600
	}
	if req.UserAgent == "" {
		req.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	oldSubscription := &models.ExternalSubscription{ID: req.ID}
	newSubscription := &models.ExternalSubscription{
		Name:           req.Name,
		URL:            req.URL,
		Enabled:        req.Enabled,
		UpdateInterval: req.UpdateInterval,
		GroupName:      req.GroupName,
		UserAgent:      req.UserAgent,
	}

	if err := oldSubscription.Update(newSubscription); err != nil {
		c.JSON(400, gin.H{
			"code": "40003",
			"msg":  fmt.Sprintf("更新订阅失败: %s", err.Error()),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新订阅成功",
	})
}

// 删除外部订阅
func ExternalSubscriptionDelete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "订阅ID不能为空",
		})
		return
	}

	subscriptionID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{
			"code": "40002",
			"msg":  "订阅ID格式错误",
		})
		return
	}

	subscription := &models.ExternalSubscription{ID: subscriptionID}
	if err := subscription.Delete(); err != nil {
		c.JSON(400, gin.H{
			"code": "40003",
			"msg":  fmt.Sprintf("删除订阅失败: %s", err.Error()),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除订阅成功",
	})
}

// 获取外部订阅列表
func ExternalSubscriptionList(c *gin.Context) {
	subscriptions, err := models.GetExternalSubscriptionList()
	if err != nil {
		c.JSON(500, gin.H{
			"code": "50001",
			"msg":  "获取订阅列表失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "获取订阅列表成功",
		"data": subscriptions,
	})
}

// 手动更新订阅
func ExternalSubscriptionRefresh(c *gin.Context) {
	var req struct {
		ID int `json:"id"`
	}

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	if req.ID == 0 {
		c.JSON(400, gin.H{
			"code": "40001",
			"msg":  "订阅ID不能为空",
		})
		return
	}

	subscription, err := models.GetExternalSubscription(req.ID)
	if err != nil {
		c.JSON(400, gin.H{
			"code": "40003",
			"msg":  "订阅不存在",
		})
		return
	}

	// 更新订阅节点
	nodeCount, err := services.UpdateSubscriptionNodes(subscription)
	if err != nil {
		c.JSON(400, gin.H{
			"code": "40004",
			"msg":  fmt.Sprintf("更新订阅失败: %s", err.Error()),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  fmt.Sprintf("更新订阅成功，获取到 %d 个节点", nodeCount),
		"data": nodeCount,
	})
}