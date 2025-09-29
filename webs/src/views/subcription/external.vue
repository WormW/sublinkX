<script setup lang='ts'>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getExternalSubscriptions,
  addExternalSubscription,
  updateExternalSubscription,
  deleteExternalSubscription,
  refreshExternalSubscription,
  type ExternalSubscription
} from "@/api/subcription/external"
import { GetGroup } from "@/api/subcription/node"
import { Refresh, Plus, Edit, Delete, Link, Timer, FolderOpened } from '@element-plus/icons-vue'

// 状态管理
const tableData = ref<ExternalSubscription[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const refreshingIds = ref<Set<number>>(new Set())
const allGroupNames = ref<string[]>([])

// 表单数据
const formData = ref<ExternalSubscription>({
  Name: '',
  URL: '',
  Enabled: true,
  UpdateInterval: 3600,
  GroupName: '',
  UserAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
})

// 表单规则
const formRules = {
  Name: [
    { required: true, message: '请输入订阅名称', trigger: 'blur' },
    { min: 1, max: 50, message: '名称长度应在1-50个字符', trigger: 'blur' }
  ],
  URL: [
    { required: true, message: '请输入订阅链接', trigger: 'blur' },
    { type: 'url' as const, message: '请输入有效的URL链接', trigger: 'blur' }
  ],
  UpdateInterval: [
    { required: true, message: '请输入更新间隔', trigger: 'blur' },
    { type: 'number' as const, min: 60, message: '更新间隔最少60秒', trigger: 'blur' }
  ]
}

// 获取订阅列表
const fetchSubscriptions = async () => {
  loading.value = true
  try {
    const { data } = await getExternalSubscriptions()
    tableData.value = data || []
  } catch (error) {
    ElMessage.error('获取订阅列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 获取分组列表
const fetchGroups = async () => {
  try {
    const { data } = await GetGroup()
    if (Array.isArray(data) && data.length > 0) {
      allGroupNames.value = data
    }
  } catch (error) {
    console.error('获取分组列表失败', error)
  }
}

// 打开添加对话框
const handleAdd = () => {
  dialogMode.value = 'add'
  formData.value = {
    Name: '',
    URL: '',
    Enabled: true,
    UpdateInterval: 3600,
    GroupName: '',
    UserAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (row: ExternalSubscription) => {
  dialogMode.value = 'edit'
  formData.value = { ...row }
  dialogVisible.value = true
}

// 保存订阅
const handleSave = async () => {
  try {
    // 准备发送的数据
    const data = {
      ...formData.value,
      ID: dialogMode.value === 'edit' ? formData.value.ID : undefined
    }

    if (dialogMode.value === 'add') {
      await addExternalSubscription(data)
      ElMessage.success('添加订阅成功')
    } else {
      await updateExternalSubscription(data)
      ElMessage.success('更新订阅成功')
    }
    dialogVisible.value = false
    await fetchSubscriptions()
  } catch (error) {
    ElMessage.error(dialogMode.value === 'add' ? '添加订阅失败' : '更新订阅失败')
    console.error(error)
  }
}

// 删除订阅
const handleDelete = async (row: ExternalSubscription) => {
  if (!row.ID) return

  await ElMessageBox.confirm(
    `确定要删除订阅 "${row.Name}" 吗？这将同时删除该订阅下的所有节点。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )

  try {
    await deleteExternalSubscription(row.ID)
    ElMessage.success('删除订阅成功')
    await fetchSubscriptions()
  } catch (error) {
    ElMessage.error('删除订阅失败')
    console.error(error)
  }
}

// 刷新订阅
const handleRefresh = async (row: ExternalSubscription) => {
  if (!row.ID) return

  refreshingIds.value.add(row.ID)
  try {
    const { data } = await refreshExternalSubscription(row.ID)
    ElMessage.success(`刷新成功，获取到 ${data} 个节点`)
    await fetchSubscriptions()
  } catch (error) {
    ElMessage.error('刷新订阅失败')
    console.error(error)
  } finally {
    refreshingIds.value.delete(row.ID)
  }
}

// 批量刷新
const handleBatchRefresh = async () => {
  const enabledSubs = tableData.value.filter(s => s.Enabled && s.ID)
  if (enabledSubs.length === 0) {
    ElMessage.warning('没有启用的订阅')
    return
  }

  ElMessage.info(`开始刷新 ${enabledSubs.length} 个订阅...`)

  for (const sub of enabledSubs) {
    if (sub.ID) {
      await handleRefresh(sub)
    }
  }

  ElMessage.success('批量刷新完成')
}

// 格式化时间
const formatDate = (dateStr?: string) => {
  if (!dateStr) return '从未更新'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '从未更新'
  return date.toLocaleString('zh-CN')
}

// 格式化更新间隔
const formatInterval = (seconds: number) => {
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}小时`
  return `${Math.floor(seconds / 86400)}天`
}

// 页面初始化
onMounted(() => {
  fetchSubscriptions()
  fetchGroups()
})
</script>

<template>
  <div class="app-container">
    <!-- 工具栏 -->
    <el-card shadow="never" class="mb-20">
      <div class="flex justify-between">
        <div>
          <el-button type="primary" :icon="Plus" @click="handleAdd">添加订阅</el-button>
          <el-button :icon="Refresh" @click="handleBatchRefresh">批量刷新</el-button>
        </div>
        <div>
          <el-button :icon="Refresh" circle @click="fetchSubscriptions" />
        </div>
      </div>
    </el-card>

    <!-- 数据表格 -->
    <el-card shadow="never">
      <el-table
        :data="tableData"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="Name" label="订阅名称" min-width="150">
          <template #default="{ row }">
            <div class="flex items-center">
              <el-icon class="mr-2"><Link /></el-icon>
              <span>{{ row.Name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="URL" label="订阅链接" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.URL" placement="top">
              <span class="truncate block max-w-xs">{{ row.URL }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="GroupName" label="关联分组" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.GroupName" type="info" size="small">
              <el-icon class="mr-1"><FolderOpened /></el-icon>
              {{ row.GroupName }}
            </el-tag>
            <span v-else class="text-gray-400">未设置</span>
          </template>
        </el-table-column>
        <el-table-column prop="Enabled" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.Enabled ? 'success' : 'danger'" size="small">
              {{ row.Enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="NodeCount" label="节点数" width="80" align="center">
          <template #default="{ row }">
            <el-tag type="info">{{ row.NodeCount || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="UpdateInterval" label="更新间隔" width="100">
          <template #default="{ row }">
            <div class="flex items-center">
              <el-icon class="mr-1"><Timer /></el-icon>
              <span>{{ formatInterval(row.UpdateInterval) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="LastUpdate" label="最后更新" min-width="160">
          <template #default="{ row }">
            <span class="text-sm text-gray-500">{{ formatDate(row.LastUpdate) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              :icon="Refresh"
              :loading="refreshingIds.has(row.ID)"
              @click="handleRefresh(row)"
            >
              刷新
            </el-button>
            <el-button
              size="small"
              :icon="Edit"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              :icon="Delete"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '添加外部订阅' : '编辑外部订阅'"
      width="600px"
    >
      <el-form :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="订阅名称" prop="Name">
          <el-input v-model="formData.Name" placeholder="请输入订阅名称" />
        </el-form-item>
        <el-form-item label="订阅链接" prop="URL">
          <el-input
            v-model="formData.URL"
            placeholder="请输入订阅链接（支持Base64编码）"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="关联分组">
          <el-select
            v-model="formData.GroupName"
            placeholder="请选择分组（可选）"
            clearable
          >
            <el-option
              v-for="group in allGroupNames"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="formData.Enabled" />
        </el-form-item>
        <el-form-item label="更新间隔" prop="UpdateInterval">
          <el-input-number
            v-model="formData.UpdateInterval"
            :min="60"
            :step="60"
            placeholder="秒"
          />
          <span class="ml-2 text-gray-500">秒（{{ formatInterval(formData.UpdateInterval) }}）</span>
        </el-form-item>
        <el-form-item label="User-Agent">
          <el-input
            v-model="formData.UserAgent"
            placeholder="自定义User-Agent（可选）"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.mb-20 {
  margin-bottom: 20px;
}
.flex {
  display: flex;
}
.justify-between {
  justify-content: space-between;
}
.items-center {
  align-items: center;
}
.mr-1 {
  margin-right: 4px;
}
.mr-2 {
  margin-right: 8px;
}
.ml-2 {
  margin-left: 8px;
}
.text-sm {
  font-size: 0.875rem;
}
.text-gray-400 {
  color: #9ca3af;
}
.text-gray-500 {
  color: #6b7280;
}
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.block {
  display: block;
}
.max-w-xs {
  max-width: 20rem;
}
</style>