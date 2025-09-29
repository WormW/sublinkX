<template>
  <div>
    <el-card>
      <template #header>
        <div class="header-wrapper">
          <span>分组管理</span>
          <el-button type="primary" @click="showAddDialog">添加分组</el-button>
        </div>
      </template>

      <el-table :data="groupList" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="分组名称" />
        <el-table-column prop="nodeCount" label="节点数量" width="120" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="editGroup(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteGroupConfirm(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑分组对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form :model="groupForm" :rules="rules" ref="groupFormRef" label-width="100px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { getGroups, addGroup, updateGroup, deleteGroup } from '@/api/subcription/group';

// 分组列表
const groupList = ref<any[]>([]);
const loading = ref(false);

// 对话框相关
const dialogVisible = ref(false);
const dialogTitle = ref('添加分组');
const isEdit = ref(false);
const editId = ref(0);

// 表单相关
const groupFormRef = ref<FormInstance>();
const groupForm = reactive({
  name: ''
});

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入分组名称', trigger: 'blur' },
    { min: 1, max: 50, message: '分组名称长度在 1 到 50 个字符', trigger: 'blur' }
  ]
};

// 获取分组列表
const fetchGroups = async () => {
  loading.value = true;
  try {
    const response = await getGroups();
    const result = response as any;
    if (result.code === '00000') {
      groupList.value = result.data || [];
    } else {
      ElMessage.error(result.msg || '获取分组列表失败');
    }
  } catch (error) {
    ElMessage.error('获取分组列表失败');
  } finally {
    loading.value = false;
  }
};

// 显示添加对话框
const showAddDialog = () => {
  dialogTitle.value = '添加分组';
  isEdit.value = false;
  editId.value = 0;
  groupForm.name = '';
  dialogVisible.value = true;
  // 清除之前的验证状态
  groupFormRef.value?.clearValidate();
};

// 编辑分组
const editGroup = (row: any) => {
  dialogTitle.value = '编辑分组';
  isEdit.value = true;
  editId.value = row.id;
  groupForm.name = row.name;
  dialogVisible.value = true;
  // 清除之前的验证状态
  groupFormRef.value?.clearValidate();
};

// 提交表单
const submitForm = async () => {
  if (!groupFormRef.value) return;

  await groupFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEdit.value) {
          // 编辑分组
          const response = await updateGroup({
            id: editId.value,
            name: groupForm.name
          });
          const result = response as any;
          if (result.code === '00000') {
            ElMessage.success('更新成功');
            dialogVisible.value = false;
            fetchGroups();
          } else {
            ElMessage.error(result.msg || '更新失败');
          }
        } else {
          // 添加分组
          const response = await addGroup({
            name: groupForm.name
          });
          const result = response as any;
          if (result.code === '00000') {
            ElMessage.success('添加成功');
            dialogVisible.value = false;
            fetchGroups();
          } else {
            ElMessage.error(result.msg || '添加失败');
          }
        }
      } catch (error) {
        ElMessage.error('操作失败');
      }
    }
  });
};

// 删除分组确认
const deleteGroupConfirm = (row: any) => {
  ElMessageBox.confirm(
    `确定要删除分组 "${row.name}" 吗？删除后该分组下的节点将不再关联到此分组。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const response = await deleteGroup(row.id);
      const result = response as any;
      if (result.code === '00000') {
        ElMessage.success('删除成功');
        fetchGroups();
      } else {
        ElMessage.error(result.msg || '删除失败');
      }
    } catch (error) {
      ElMessage.error('删除失败');
    }
  }).catch(() => {
    // 用户取消删除
  });
};

// 组件挂载时获取数据
onMounted(() => {
  fetchGroups();
});
</script>

<style scoped>
.header-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>