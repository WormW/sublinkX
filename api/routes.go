package api

import (
	"github.com/gin-gonic/gin"
)

// Route 定义路由结构
type Route struct {
	Path      string  `json:"path"`
	Name      string  `json:"name,omitempty"`
	Component string  `json:"component"`
	Meta      RouteMeta `json:"meta,omitempty"`
	Children  []Route `json:"children,omitempty"`
}

type RouteMeta struct {
	Title     string   `json:"title,omitempty"`
	Icon      string   `json:"icon,omitempty"`
	Hidden    bool     `json:"hidden,omitempty"`
	AlwaysShow bool    `json:"alwaysShow,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// GetRoutes 返回动态路由
func GetRoutes(c *gin.Context) {
	routes := []Route{
		{
			Path:      "/subcription",
			Component: "Layout",
			Meta: RouteMeta{
				Title: "订阅管理",
				Icon:  "link",
			},
			Children: []Route{
				{
					Path:      "subs",
					Name:      "Subs",
					Component: "subcription/subs",
					Meta: RouteMeta{
						Title: "订阅列表",
						Icon:  "list",
					},
				},
				{
					Path:      "nodes",
					Name:      "Nodes",
					Component: "subcription/nodes",
					Meta: RouteMeta{
						Title: "节点列表",
						Icon:  "node",
					},
				},
				{
					Path:      "groups",
					Name:      "Groups",
					Component: "subcription/groups",
					Meta: RouteMeta{
						Title: "分组管理",
						Icon:  "folder",
					},
				},
				{
					Path:      "external",
					Name:      "External",
					Component: "subcription/external",
					Meta: RouteMeta{
						Title: "外部订阅",
						Icon:  "cloud",
					},
				},
				{
					Path:      "template",
					Name:      "Template",
					Component: "subcription/template",
					Meta: RouteMeta{
						Title: "订阅模板",
						Icon:  "template",
					},
				},
			},
		},
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"data": routes,
		"msg":  "获取路由成功",
	})
}