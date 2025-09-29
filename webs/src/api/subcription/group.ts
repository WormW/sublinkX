import request from "@/utils/request";

// 获取分组列表
export function getGroups() {
  return request({
    url: "/api/v1/nodes/group/get",
    method: "get",
  });
}

// 添加分组
export function addGroup(data: { name: string }) {
  return request({
    url: "/api/v1/nodes/group/add",
    method: "post",
    data,
  });
}

// 更新分组
export function updateGroup(data: { id: number; name: string }) {
  return request({
    url: "/api/v1/nodes/group/update",
    method: "post",
    data,
  });
}

// 删除分组
export function deleteGroup(id: number) {
  return request({
    url: "/api/v1/nodes/group/delete",
    method: "delete",
    params: { id },
  });
}

// 设置节点分组
export function setNodeGroup(data: { name: string; group: string }) {
  return request({
    url: "/api/v1/nodes/group/set",
    method: "post",
    data,
  });
}

// 根据分组ID获取节点列表
export function getNodesByGroup(groupId: number) {
  return request({
    url: "/api/v1/nodes/group/nodes",
    method: "get",
    params: { groupId },
  });
}