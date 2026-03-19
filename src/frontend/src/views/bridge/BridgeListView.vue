<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">桥梁管理</h1>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>添加桥梁
      </el-button>
    </div>
    
    <el-card class="table-card">
      <el-table :data="bridgeList" v-loading="loading" stripe>
        <el-table-column type="index" width="60" label="序号" />
        <el-table-column prop="bridge_name" label="桥梁名称" min-width="150" />
        <el-table-column prop="bridge_code" label="桥梁编号" width="120" />
        <el-table-column prop="bridge_type" label="类型" width="100" />
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看</el-button>
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
      :title="isEdit ? '编辑桥梁' : '添加桥梁'"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="桥梁名称" prop="bridge_name">
              <el-input v-model="form.bridge_name" placeholder="请输入桥梁名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="桥梁编号" prop="bridge_code">
              <el-input v-model="form.bridge_code" placeholder="请输入桥梁编号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="桥梁类型" prop="bridge_type">
              <el-select v-model="form.bridge_type" placeholder="请选择类型" style="width: 100%">
                <el-option label="梁桥" value="梁桥" />
                <el-option label="拱桥" value="拱桥" />
                <el-option label="斜拉桥" value="斜拉桥" />
                <el-option label="悬索桥" value="悬索桥" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="建造年份" prop="build_year">
              <el-input-number v-model="form.build_year" :min="1900" :max="2100" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="详细地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入详细地址" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="经度" prop="longitude">
              <el-input-number v-model="form.longitude" :precision="6" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="纬度" prop="latitude">
              <el-input-number v-model="form.latitude" :precision="6" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="长度(m)" prop="length">
              <el-input-number v-model="form.length" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="宽度(m)" prop="width">
              <el-input-number v-model="form.width" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getBridgeList, createBridge, updateBridge, deleteBridge } from '@/api/bridge'
import type { Bridge } from '@/types'

const router = useRouter()

const loading = ref(false)
const bridgeList = ref<Bridge[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  bridge_id: '',
  bridge_name: '',
  bridge_code: '',
  bridge_type: '',
  build_year: new Date().getFullYear(),
  address: '',
  longitude: 116.4074,
  latitude: 39.9042,
  length: 0,
  width: 0,
  remark: ''
})

const rules: FormRules = {
  bridge_name: [{ required: true, message: '请输入桥梁名称', trigger: 'blur' }],
  bridge_code: [{ required: true, message: '请输入桥梁编号', trigger: 'blur' }],
  bridge_type: [{ required: true, message: '请选择桥梁类型', trigger: 'change' }],
  address: [{ required: true, message: '请输入详细地址', trigger: 'blur' }],
  longitude: [{ required: true, message: '请输入经度', trigger: 'blur' }],
  latitude: [{ required: true, message: '请输入纬度', trigger: 'blur' }],
  build_year: [{ required: true, message: '请输入建造年份', trigger: 'blur' }],
  length: [{ required: true, message: '请输入长度', trigger: 'blur' }],
  width: [{ required: true, message: '请输入宽度', trigger: 'blur' }]
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    '正常': 'success',
    '异常': 'danger',
    '维修中': 'warning'
  }
  return map[status] || 'info'
}

const loadData = async () => {
  loading.value = true
  try {
    const data = await getBridgeList({ page: page.value, page_size: pageSize.value })
    bridgeList.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    bridge_id: '',
    bridge_name: '',
    bridge_code: '',
    bridge_type: '',
    build_year: new Date().getFullYear(),
    address: '',
    longitude: 116.4074,
    latitude: 39.9042,
    length: 0,
    width: 0,
    remark: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row: Bridge) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleView = (row: Bridge) => {
  router.push(`/bridges/${row.bridge_id}`)
}

const handleDelete = async (row: Bridge) => {
  try {
    await ElMessageBox.confirm('确定要删除该桥梁吗？', '提示', { type: 'warning' })
    await deleteBridge(row.bridge_id)
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
          await updateBridge(form.bridge_id, form)
          ElMessage.success('更新成功')
        } else {
          await createBridge(form)
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
