import request from "@/utils/request";

// 外部订阅接口类型定义
export interface ExternalSubscription {
  ID?: number;
  Name: string;
  URL: string;
  Enabled: boolean;
  UpdateInterval: number;
  LastUpdate?: string;
  NodeCount?: number;
  GroupName?: string;
  UserAgent?: string;
}

// 获取外部订阅列表
export function getExternalSubscriptions() {
  return request({
    url: "/api/external-subscription/list",
    method: "get",
  });
}

// 添加外部订阅
export function addExternalSubscription(data: ExternalSubscription) {
  return request({
    url: "/api/external-subscription/add",
    method: "post",
    data: data,
  });
}

// 更新外部订阅
export function updateExternalSubscription(data: ExternalSubscription) {
  return request({
    url: "/api/external-subscription/update",
    method: "post",
    data: data,
  });
}

// 删除外部订阅
export function deleteExternalSubscription(id: number) {
  return request({
    url: "/api/external-subscription/delete",
    method: "delete",
    params: { id },
  });
}

// 刷新外部订阅
export function refreshExternalSubscription(id: number) {
  return request({
    url: "/api/external-subscription/refresh",
    method: "post",
    data: { id },
  });
}