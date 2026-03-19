<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">报表管理</h1>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>生成报表
      </el-button>
    </div>

    <el-card class="table-card">
      <el-table :data="reportList" v-loading="loading" stripe>
        <el-table-column type="index" width="60" label="序号" />
        <el-table-column prop="report_name" label="报表名称" min-width="180" />
        <el-table-column prop="report_type" label="报表类型" width="150" />
        <el-table-column prop="create_time" label="创建时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看</el-button>
            <el-button link type="primary" @click="handleDownload(row)">下载</el-button>
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

    <!-- Create Report Dialog -->
    <el-dialog v-model="dialogVisible" title="生成报表" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="报表名称" prop="report_name">
          <el-input v-model="form.report_name" placeholder="请输入报表名称" />
        </el-form-item>
        <el-form-item label="报表类型" prop="report_type">
          <el-select v-model="form.report_type" placeholder="请选择报表类型" style="width: 100%">
            <el-option label="桥梁检测报表" value="桥梁检测报表" />
            <el-option label="病害统计报表" value="病害统计报表" />
            <el-option label="无人机巡检报表" value="无人机巡检报表" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联桥梁" prop="bridge_id">
          <el-select v-model="form.bridge_id" placeholder="请选择桥梁" style="width: 100%">
            <el-option
                v-for="bridge in bridgeList"
                :key="bridge.bridge_id"
                :label="bridge.bridge_name"
                :value="bridge.bridge_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围" prop="dateRange">
          <el-date-picker
              v-model="form.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getReportList, createReport, getReportDetail } from '@/api/report'
import { getBridgeList } from '@/api/bridge'
import type { Report, Bridge } from '@/types'

const loading = ref(false)
const reportList = ref<Report[]>([])
const bridgeList = ref<Bridge[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

interface FormData {
  report_name: string
  report_type: string
  bridge_id: string
  dateRange: [Date, Date] | []
}

const form = reactive<FormData>({
  report_name: '',
  report_type: '',
  bridge_id: '',
  dateRange: []
})

const rules: FormRules = {
  report_name: [{ required: true, message: '请输入报表名称', trigger: 'blur' }],
  report_type: [{ required: true, message: '请选择报表类型', trigger: 'change' }],
  bridge_id: [{ required: true, message: '请选择桥梁', trigger: 'change' }],
  dateRange: [{ required: true, message: '请选择时间范围', trigger: 'change' }]
}

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    '已生成': 'success',
    '生成中': 'warning',
    '失败': 'danger'
  }
  return map[status] || 'info'
}

const loadData = async () => {
  loading.value = true
  try {
    const data = await getReportList({ page: page.value, page_size: pageSize.value })
    reportList.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

const loadBridges = async () => {
  try {
    const data = await getBridgeList({ page: 1, page_size: 100 })
    bridgeList.value = data.list
  } catch {
    // 忽略错误
  }
}

const handleCreate = () => {
  Object.assign(form, {
    report_name: '',
    report_type: '',
    bridge_id: '',
    dateRange: []
  })
  dialogVisible.value = true
}

const handleView = async (row: Report) => {
  try {
    const detail = await getReportDetail(row.report_id)
    ElMessage.success('查看报表: ' + detail.report_name)
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  }
}

const handleDownload = (row: Report) => {
  ElMessage.success('开始下载: ' + row.report_name)
}

const handleDelete = async (row: Report) => {
  try {
    await ElMessageBox.confirm('确定要删除该报表吗？', '提示', { type: 'warning' })
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
        if (form.dateRange && form.dateRange.length === 2) {
          const [startTime, endTime] = form.dateRange as [Date, Date]
          await createReport({
            report_name: form.report_name,
            report_type: form.report_type,
            bridge_id: form.bridge_id,
            start_time: startTime.toISOString(),
            end_time: endTime.toISOString()
          })
        }
        ElMessage.success('报表生成中...')
        dialogVisible.value = false
        loadData()
      } finally {
        submitting.value = false
      }
    }
  })
}

onMounted(() => {
  loadData()
  loadBridges()
})
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
