package models

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

// ExternalSubscription 外部订阅链接模型
type ExternalSubscription struct {
	gorm.Model
	ID             int
	Name           string    `json:"Name" gorm:"uniqueIndex;not null"`     // 订阅名称
	URL            string    `json:"URL" gorm:"not null"`                   // 订阅链接
	Enabled        bool      `json:"Enabled" gorm:"default:true"`           // 是否启用
	UpdateInterval int       `json:"UpdateInterval" gorm:"default:3600"`    // 更新间隔（秒）
	LastUpdate     time.Time `json:"LastUpdate"`                            // 最后更新时间
	NodeCount      int       `json:"NodeCount"`                             // 节点数量
	GroupName      string    `json:"GroupName"`                             // 关联的分组名称
	UserAgent      string    `json:"UserAgent"`                             // 自定义User-Agent
	Nodes          []Node    `json:"Nodes" gorm:"many2many:subscription_nodes"` // 关联的节点
}

// 添加外部订阅
func (es *ExternalSubscription) Add() error {
	// 检查订阅是否已存在
	var existingSubscription ExternalSubscription
	result := DB.Model(es).Where("name = ? OR url = ?", es.Name, es.URL).First(&existingSubscription)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("添加外部订阅错误:", result.Error)
		return result.Error
	}
	if result.RowsAffected > 0 {
		return errors.New("订阅名称或链接已存在")
	}

	// 设置默认值
	if es.UpdateInterval == 0 {
		es.UpdateInterval = 3600 // 默认1小时更新一次
	}
	if es.UserAgent == "" {
		es.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	return DB.Create(es).Error
}

// 更新外部订阅
func (es *ExternalSubscription) Update(newES *ExternalSubscription) error {
	// 检查是否存在其他同名或同链接的订阅
	var existingSubscription ExternalSubscription
	result := DB.Model(&ExternalSubscription{}).
		Where("(name = ? OR url = ?) AND id != ?", newES.Name, newES.URL, es.ID).
		First(&existingSubscription)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return errors.New("订阅名称或链接已被其他订阅使用")
	}

	// 使用结构体更新，但排除ID字段
	return DB.Model(es).Where("id = ?", es.ID).Omit("ID", "CreatedAt", "DeletedAt").Updates(newES).Error
}

// 删除外部订阅
func (es *ExternalSubscription) Delete() error {
	// 先查询订阅信息
	result := DB.Model(es).Where("id = ?", es.ID).First(es)
	if result.Error != nil {
		return result.Error
	}

	// 预加载关联的节点
	DB.Model(es).Preload("Nodes").First(es)

	// 清除关联的节点
	if len(es.Nodes) > 0 {
		err := DB.Model(es).Association("Nodes").Clear()
		if err != nil {
			log.Println("清除订阅关联节点失败:", err)
			return err
		}

		// 删除关联的节点记录
		for _, node := range es.Nodes {
			if err := node.Del(); err != nil {
				log.Println("删除节点失败:", err)
			}
		}
	}

	// 删除订阅记录
	return DB.Delete(es).Error
}

// 获取单个外部订阅
func GetExternalSubscription(id int) (*ExternalSubscription, error) {
	var es ExternalSubscription
	result := DB.Model(&ExternalSubscription{}).Preload("Nodes").Where("id = ?", id).First(&es)
	if result.Error != nil {
		return nil, result.Error
	}
	return &es, nil
}

// 获取所有外部订阅列表
func GetExternalSubscriptionList() ([]ExternalSubscription, error) {
	var subscriptions []ExternalSubscription
	result := DB.Model(&ExternalSubscription{}).Preload("Nodes").Find(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

// 获取启用的外部订阅列表
func GetEnabledExternalSubscriptions() ([]ExternalSubscription, error) {
	var subscriptions []ExternalSubscription
	result := DB.Model(&ExternalSubscription{}).Where("enabled = ?", true).Preload("Nodes").Find(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

// 更新最后更新时间和节点数量
func (es *ExternalSubscription) UpdateStatus(nodeCount int) error {
	return DB.Model(es).Where("id = ?", es.ID).Updates(map[string]interface{}{
		"last_update": time.Now(),
		"node_count":  nodeCount,
	}).Error
}

// 清除订阅的所有节点
func (es *ExternalSubscription) ClearNodes() error {
	// 预加载关联的节点
	DB.Model(es).Preload("Nodes").First(es)

	if len(es.Nodes) > 0 {
		// 清除关联
		err := DB.Model(es).Association("Nodes").Clear()
		if err != nil {
			return err
		}

		// 删除节点记录
		for _, node := range es.Nodes {
			if err := node.Del(); err != nil {
				log.Println("删除节点失败:", err)
			}
		}
	}
	return nil
}

// 关联节点到订阅
func (es *ExternalSubscription) AssociateNodes(nodes []Node) error {
	if len(nodes) == 0 {
		return nil
	}
	return DB.Model(es).Association("Nodes").Append(nodes)
}