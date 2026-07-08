<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="max-w-2xl space-y-6">
    <!-- Header -->
    <div>
      <h3 class="text-sm font-semibold text-arena-100">Maintenance</h3>
      <p class="text-xs text-arena-500 mt-1">Housekeeping tools for the configuration store.</p>
    </div>

    <!-- Dead references cleanup -->
    <Card color="blue">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 bg-blue-500/10">
          <Icon name="refresh" size="md" class="text-blue-400" />
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-[13px] font-medium text-arena-100">Clean Up Dead References</h4>
          <p class="text-xs text-arena-500 mt-0.5">
            Finds references pointing at agents, flows, backends or other entities that no longer exist
            (left behind by deletes performed before referential integrity) and removes them.
          </p>
          <button
            @click="scan"
            :disabled="scanning"
            class="mt-3 px-4 py-1.5 bg-blue-500/15 hover:bg-blue-500/25 text-blue-300 text-xs font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ scanning ? 'Scanning...' : 'Scan & Clean' }}
          </button>
        </div>
      </div>
    </Card>

    <ReferenceListDialog
      ref="dialogRef"
      :title="`${dead.length} dead ${dead.length === 1 ? 'reference' : 'references'} found`"
      subtitle="These point at entities that no longer exist. Cleaning removes the references, never the entities holding them."
      :rows="rows"
      :confirm-label="cleaning ? 'Cleaning...' : 'Clean up'"
      confirm-class="bg-blue-500/15 hover:bg-blue-500/25 text-blue-300"
      :busy="cleaning"
      @confirm="clean"
    />
  </div>
</template>

<script setup>
import { computed, inject, ref } from 'vue'
import { integrityApi } from '../../lib/api/index.js'
import { useDataStore } from '../../lib/stores/data.js'
import Card from '../../components/Card.vue'
import Icon from '../../components/Icon.vue'
import ReferenceListDialog from '../../components/ReferenceListDialog.vue'

const toast = inject('toast')
const store = useDataStore()

const dialogRef = ref(null)
const dead = ref([])
const scanning = ref(false)
const cleaning = ref(false)

const rows = computed(() => dead.value.map((ref) => ({
  referrerType: ref.referrerType,
  referrerName: ref.referrerName,
  referrerId: ref.referrerId,
  detail: `${ref.field} → ${shortId(ref.targetId)}`,
  dot: 'bg-blue-400',
})))

function shortId(id) {
  return id.length > 12 ? id.slice(0, 8) + '…' : id
}

async function scan() {
  scanning.value = true
  try {
    const resp = await integrityApi.deadReferences()
    dead.value = resp.references || []
    if (dead.value.length === 0) {
      toast.success('No dead references found')
      return
    }
    dialogRef.value?.open()
  } catch (e) {
    toast.error(e.message)
  } finally {
    scanning.value = false
  }
}

async function clean() {
  cleaning.value = true
  try {
    const resp = await integrityApi.cleanDeadReferences()
    dialogRef.value?.close()
    toast.success(`Removed ${resp.removed} dead ${resp.removed === 1 ? 'reference' : 'references'}`)
    await store.refresh()
  } catch (e) {
    toast.error(e.message)
  } finally {
    cleaning.value = false
  }
}
</script>
