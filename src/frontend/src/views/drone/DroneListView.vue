<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">无人机管理</h1>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>添加无人机
      </el-button>
    </div>
    
    <el-card class="table-card">
      <el-table :data="droneList" v-loading="loading" stripe>
        <el-table-column type="index" width="60" label="序号" />
        <el-table-column prop="drone_name" label="无人机名称" min-width="150" />
        <el-table-column prop="drone_code" label="编号" width="120" />
        <el-table-column prop="model" label="型号" width="150" />
        <el-table-column prop="camera_param" label="摄像头参数" min-width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>
    
    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑无人机' : '添加无人机'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="无人机名称" prop="drone_name">
          <el-input v-model="form.drone_name" placeholder="请输入无人机名称" />
        </el-form-item>
        <el-form-item label="编号" prop="drone_code">
          <el-input v-model="form.drone_code" placeholder="请输入编号" />
        </el-form-item>
        <el-form-item label="型号" prop="model">
          <el-input v-model="form.model" placeholder="请输入型号" />
        </el-form-item>
        <el-form-item label="摄像头参数" prop="camera_param">
          <el-input v-model="form.camera_param" placeholder="请输入摄像头参数" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getDroneList, createDrone, updateDrone, deleteDrone } from '@/api/drone'
import type { Drone } from '@/types'

const loading = ref(false)
const droneList = ref<Drone[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  drone_id: '',
  drone_name: '',
  drone_code: '',
  model: '',
  camera_param: '',
  remark: ''
})

const rules: FormRules = {
  drone_name: [{ required: true, message: '请输入无人机名称', trigger: 'blur' }],
  drone_code: [{ required: true, message: '请输入编号', trigger: 'blur' }],
  model: [{ required: true, message: '请输入型号', trigger: 'blur' }],
  camera_param: [{ required: true, message: '请输入摄像头参数', trigger: 'blur' }]
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    '在线': 'success',
    '离线': 'info',
    '维修中': 'warning',
    '故障': 'danger'
  }
  return map[status] || 'info'
}

const loadData = async () => {
  loading.value = true
  try {
    const data = await getDroneList({ page: page.value, page_size: pageSize.value })
    droneList.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    drone_id: '',
    drone_name: '',
    drone_code: '',
    model: '',
    camera_param: '',
    remark: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row: Drone) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row: Drone) => {
  try {
    await ElMessageBox.confirm('确定要删除该无人机吗？', '提示', { type: 'warning' })
    await deleteDrone(row.drone_id)
    ElMessage.success('删除成功')
    loadData()
  } catch {
    // 取消删除
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value) {
          await updateDrone(form.drone_id, form)
          ElMessage.success('更新成功')
        } else {
          await createDrone(form)
          ElMessage.success('添加成功')
        }
        dialogVisible.value = false
        loadData()
      } finally {
        submitting.value = false
      }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0;
}

.table-card {
  border-radius: 12px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
