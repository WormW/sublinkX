package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

// ExternalSubscription 外部订阅路由
func ExternalSubscription(r *gin.Engine) {
	subscription := r.Group("/api/external-subscription")
	{
		// 获取订阅列表
		subscription.GET("/list", api.ExternalSubscriptionList)
		// 添加订阅
		subscription.POST("/add", api.ExternalSubscriptionAdd)
		// 更新订阅
		subscription.POST("/update", api.ExternalSubscriptionUpdate)
		// 删除订阅
		subscription.DELETE("/delete", api.ExternalSubscriptionDelete)
		// 手动刷新订阅
		subscription.POST("/refresh", api.ExternalSubscriptionRefresh)
	}
}