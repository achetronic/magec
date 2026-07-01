<template>
  <div
    class="flow-node group absolute select-none"
    :class="{ 'is-selected': selected, 'is-entry': isEntry }"
    :style="nodeStyle"
    :data-node-id="node.id"
    @pointerdown.stop="onBodyPointerDown"
  >
    <!-- Entry flag -->
    <div
      v-if="isEntry"
      class="absolute -top-2.5 left-3 z-10 flex items-center gap-1 px-1.5 py-px rounded-full bg-rose-500/20 border border-rose-500/40"
    >
      <svg class="w-2.5 h-2.5 text-rose-300" viewBox="0 0 24 24" fill="currentColor">
        <path d="M3 3h13l-2 4 2 4H3z" />
      </svg>
      <span class="text-[8px] font-bold uppercase tracking-wider text-rose-300">Entry</span>
    </div>

    <div
      class="rounded-xl border bg-piedra-900/95 backdrop-blur-sm shadow-lg transition-all flex flex-col h-full"
      :class="[
        borderClass,
        selected ? selectedRingClass : hoverBorderClass,
      ]"
    >
      <!-- Header (drag handle) -->
      <div
        class="flow-node-drag flex items-center gap-2 px-3 py-2 cursor-grab active:cursor-grabbing rounded-t-xl"
        :class="headerBgClass"
      >
        <div class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0" :class="iconBgClass">
          <svg class="w-3.5 h-3.5" :class="iconColorClass" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.6">
            <path stroke-linecap="round" stroke-linejoin="round" :d="iconPath" />
          </svg>
        </div>
        <span class="flex-1 text-[9px] font-bold uppercase tracking-wider truncate" :class="labelColorClass">{{ typeLabel }}</span>
        <button
          @pointerdown.stop @click.stop="$emit('remove')"
          class="p-0.5 rounded opacity-0 group-hover:opacity-100 hover:bg-lava-500/20 transition-all flex-shrink-0"
          title="Delete node"
        >
          <svg class="w-3 h-3 text-arena-500 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- AGENT body -->
      <div v-if="node.type === 'agent'" class="px-3 py-2.5 space-y-2">
        <button
          @pointerdown.stop @click.stop="pickerOpen = !pickerOpen"
          class="w-full flex items-center gap-1.5 text-left text-xs font-medium outline-none cursor-pointer"
          :class="node.agentId ? 'text-arena-100' : 'text-arena-500 italic'"
        >
          <span class="truncate flex-1">{{ agentName }}</span>
          <svg class="w-3 h-3 flex-shrink-0 text-arena-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
          </svg>
        </button>
        <button
          @pointerdown.stop @click.stop="$emit('update', { ...node, responseAgent: !node.responseAgent })"
          class="flex items-center gap-1.5 text-[9px] font-medium px-1.5 py-0.5 rounded-md transition-all"
          :class="node.responseAgent ? 'bg-green-500/15 text-green-400' : 'text-arena-600 hover:text-arena-400 hover:bg-piedra-800'"
          :title="node.responseAgent ? 'This agent contributes to the final response' : 'Include this output in the flow response'"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.6">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.076-4.076a1.526 1.526 0 0 1 1.037-.443 48.2 48.2 0 0 0 5.887-.512c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.4 48.4 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" />
          </svg>
          Response
        </button>

        <!-- agent picker dropdown -->
        <Transition name="dropdown">
          <div v-if="pickerOpen" class="absolute z-50 left-2 right-2 top-full mt-1 bg-piedra-800 border border-piedra-700/60 rounded-xl shadow-2xl overflow-hidden" @pointerdown.stop>
            <div v-if="agents.length" class="py-1 max-h-44 overflow-y-auto">
              <button
                v-for="a in agents" :key="a.id"
                @click.stop="pickAgent(a.id)"
                class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
                :class="a.id === node.agentId ? 'bg-sol-500/10' : 'hover:bg-piedra-700/60'"
              >
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-medium truncate" :class="a.id === node.agentId ? 'text-sol-300' : 'text-arena-100'">{{ a.name || a.id }}</div>
                  <div v-if="a.description" class="text-[9px] text-arena-500 truncate">{{ a.description }}</div>
                </div>
                <svg v-if="a.id === node.agentId" class="w-3.5 h-3.5 text-sol-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              </button>
            </div>
            <div v-else class="px-3 py-4 text-[10px] text-arena-500 italic text-center">No agents available</div>
          </div>
        </Transition>
      </div>

      <!-- ROUTER body -->
      <div v-else-if="node.type === 'router'" class="px-2.5 py-2 space-y-1.5">
        <p class="px-0.5 text-[9px] text-arena-500 leading-snug">
          First rule whose <span class="text-atlantico-300 font-medium">CEL</span> guard is true wins. <code class="text-arena-400">state.x</code>, <code class="text-arena-400">iterations</code> available.
        </p>
        <!-- rules -->
        <div
          v-for="(rule, i) in node.rules" :key="'r' + i"
          class="relative flex items-center gap-1.5 rounded-lg bg-piedra-800/70 border border-piedra-700/50 pl-2 pr-3 py-1"
        >
          <input
            :value="rule.when"
            @input="updateRule(i, 'when', $event.target.value)"
            @pointerdown.stop
            placeholder="state.score >= 0.8"
            class="flex-1 min-w-0 bg-transparent text-[10px] font-mono text-arena-200 placeholder:text-arena-600 outline-none"
          />
          <span class="text-arena-600 text-[10px]">→</span>
          <input
            :value="rule.route"
            @input="updateRule(i, 'route', $event.target.value)"
            @pointerdown.stop
            placeholder="route"
            class="w-14 bg-transparent text-[10px] font-mono font-medium text-atlantico-300 placeholder:text-arena-600 outline-none"
          />
          <button @pointerdown.stop @click.stop="removeRule(i)" class="p-0.5 rounded opacity-0 group-hover:opacity-100 hover:bg-lava-500/20 transition-all">
            <svg class="w-2.5 h-2.5 text-arena-500 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
          <!-- per-route output port -->
          <span
            class="flow-port flow-port-out"
            :data-port="'out:' + node.id + '|' + (rule.route || '')"
            :title="'Drag to connect route: ' + (rule.route || '?')"
            @pointerdown.stop.prevent="$emit('start-edge', { nodeId: node.id, route: rule.route })"
          />
        </div>
        <button @pointerdown.stop @click.stop="addRule" class="w-full flex items-center justify-center gap-1 py-1 rounded-lg border border-dashed border-piedra-700/60 text-[9px] text-arena-500 hover:text-atlantico-300 hover:border-atlantico-500/40 transition-colors">
          <svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" /></svg>
          Add rule
        </button>
        <!-- default route -->
        <div class="relative flex items-center gap-1.5 rounded-lg bg-piedra-800/40 border border-piedra-700/40 pl-2 pr-3 py-1">
          <span class="flex-1 text-[9px] font-medium text-arena-500 uppercase tracking-wide">otherwise</span>
          <span class="text-arena-600 text-[10px]">→</span>
          <input
            :value="node.defaultRoute"
            @input="$emit('update', { ...node, defaultRoute: $event.target.value })"
            @pointerdown.stop
            placeholder="default"
            class="w-14 bg-transparent text-[10px] font-mono font-medium text-atlantico-300 placeholder:text-arena-600 outline-none"
          />
          <span
            class="flow-port flow-port-out"
            :data-port="'out:' + node.id + '|' + (node.defaultRoute || '')"
            :title="'Drag to connect default route: ' + (node.defaultRoute || '?')"
            @pointerdown.stop.prevent="$emit('start-edge', { nodeId: node.id, route: node.defaultRoute })"
          />
        </div>
      </div>

      <!-- JOIN body -->
      <div v-else-if="node.type === 'join'" class="px-3 py-2.5">
        <p class="text-[9px] text-arena-500 leading-snug">Waits for <span class="text-purple-300 font-medium">all</span> incoming branches, then forwards their combined output.</p>
      </div>

      <!-- PARALLEL body -->
      <div v-else-if="node.type === 'parallel'" class="px-3 py-2.5 space-y-2">
        <button
          @pointerdown.stop @click.stop="pickerOpen = !pickerOpen"
          class="w-full flex items-center gap-1.5 text-left text-xs font-medium outline-none cursor-pointer"
          :class="node.agentId ? 'text-arena-100' : 'text-arena-500 italic'"
        >
          <span class="truncate flex-1">{{ agentName }}</span>
          <svg class="w-3 h-3 flex-shrink-0 text-arena-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
          </svg>
        </button>
        <div class="flex items-center gap-1.5">
          <span class="text-[9px] text-arena-500">runs per list item, max</span>
          <input
            type="number" min="0"
            :value="node.maxConcurrency || 0"
            @input="$emit('update', { ...node, maxConcurrency: Math.max(0, parseInt($event.target.value) || 0) })"
            @pointerdown.stop
            class="w-12 bg-piedra-800 border border-piedra-700/50 rounded px-1.5 py-0.5 text-[10px] font-mono text-arena-200 outline-none focus:border-lava-500/50"
          />
          <span class="text-[9px] text-arena-600">(0 = unlimited)</span>
        </div>

        <Transition name="dropdown">
          <div v-if="pickerOpen" class="absolute z-50 left-2 right-2 top-full mt-1 bg-piedra-800 border border-piedra-700/60 rounded-xl shadow-2xl overflow-hidden" @pointerdown.stop>
            <div v-if="agents.length" class="py-1 max-h-44 overflow-y-auto">
              <button
                v-for="a in agents" :key="a.id"
                @click.stop="pickAgent(a.id)"
                class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
                :class="a.id === node.agentId ? 'bg-lava-500/10' : 'hover:bg-piedra-700/60'"
              >
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-medium truncate" :class="a.id === node.agentId ? 'text-lava-300' : 'text-arena-100'">{{ a.name || a.id }}</div>
                  <div v-if="a.description" class="text-[9px] text-arena-500 truncate">{{ a.description }}</div>
                </div>
              </button>
            </div>
            <div v-else class="px-3 py-4 text-[10px] text-arena-500 italic text-center">No agents available</div>
          </div>
        </Transition>
      </div>

      <!-- SUBFLOW body -->
      <div v-else-if="node.type === 'subflow'" class="px-3 py-2.5 space-y-2">
        <button
          @pointerdown.stop @click.stop="pickerOpen = !pickerOpen"
          class="w-full flex items-center gap-1.5 text-left text-xs font-medium outline-none cursor-pointer"
          :class="node.flowId ? 'text-arena-100' : 'text-arena-500 italic'"
        >
          <span class="truncate flex-1">{{ flowName }}</span>
          <svg class="w-3 h-3 flex-shrink-0 text-arena-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
          </svg>
        </button>
        <p class="text-[9px] text-arena-500 leading-snug">Embeds another flow as a nested step.</p>

        <Transition name="dropdown">
          <div v-if="pickerOpen" class="absolute z-50 left-2 right-2 top-full mt-1 bg-piedra-800 border border-piedra-700/60 rounded-xl shadow-2xl overflow-hidden" @pointerdown.stop>
            <div v-if="flows.length" class="py-1 max-h-44 overflow-y-auto">
              <button
                v-for="fl in flows" :key="fl.id"
                @click.stop="pickFlow(fl.id)"
                class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
                :class="fl.id === node.flowId ? 'bg-rose-500/10' : 'hover:bg-piedra-700/60'"
              >
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-medium truncate" :class="fl.id === node.flowId ? 'text-rose-300' : 'text-arena-100'">{{ fl.name || fl.id }}</div>
                  <div v-if="fl.description" class="text-[9px] text-arena-500 truncate">{{ fl.description }}</div>
                </div>
              </button>
            </div>
            <div v-else class="px-3 py-4 text-[10px] text-arena-500 italic text-center">No other flows available</div>
          </div>
        </Transition>
      </div>

      <!-- EXPRESSION body -->
      <div v-else-if="node.type === 'expression'" class="px-2.5 py-2 space-y-1.5 flex-1 flex flex-col min-h-0">
        <div class="flex items-center gap-1 px-0.5">
          <span class="text-[9px] text-arena-500">CEL over <code class="text-indigo-300">input</code> and <code class="text-indigo-300">state</code></span>
          <span class="relative group/tip">
            <svg class="w-3 h-3 text-arena-600 hover:text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.6"><path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" /></svg>
            <span class="absolute z-50 left-4 top-0 hidden group-hover/tip:block w-52 p-2 rounded-lg bg-piedra-950 border border-piedra-700/60 text-[9px] text-arena-400 leading-relaxed shadow-xl">
              The result becomes this node's output. Use a split to produce a list for a For-Each; list helpers .map() and .filter() are available.
            </span>
          </span>
        </div>
        <textarea
          :value="node.expression"
          @input="$emit('update', { ...node, expression: $event.target.value })"
          @pointerdown.stop
          rows="2" spellcheck="false"
          :placeholder="exprPlaceholder"
          class="w-full flex-1 min-h-[2.5rem] bg-piedra-800 border border-piedra-700/50 rounded-lg px-2 py-1 text-[10px] font-mono text-arena-200 placeholder:text-arena-600 outline-none focus:border-indigo-500/50 resize-none"
        />
        <div class="flex items-center gap-1.5">
          <span class="text-[9px] text-arena-500 whitespace-nowrap">save as</span>
          <input
            :value="node.outputKey"
            @input="$emit('update', { ...node, outputKey: $event.target.value })"
            @pointerdown.stop
            placeholder="state key (optional)"
            class="flex-1 min-w-0 bg-piedra-800 border border-piedra-700/50 rounded px-1.5 py-0.5 text-[10px] font-mono text-arena-200 placeholder:text-arena-600 outline-none focus:border-indigo-500/50"
          />
        </div>
      </div>

      <!-- TEMPLATE body -->
      <div v-else-if="node.type === 'template'" class="px-2.5 py-2 space-y-1.5 flex-1 flex flex-col min-h-0">
        <div class="flex items-center gap-1 px-0.5">
          <span class="text-[9px] text-arena-500">text with <code class="text-teal-300">{{ mustacheInput }}</code> placeholders</span>
          <span class="relative group/tip">
            <svg class="w-3 h-3 text-arena-600 hover:text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.6"><path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" /></svg>
            <span class="absolute z-50 left-4 top-0 hidden group-hover/tip:block w-52 p-2 rounded-lg bg-piedra-950 border border-piedra-700/60 text-[9px] text-arena-400 leading-relaxed shadow-xl">
              Reference input, input.field or state.key inside double braces. Renders a string — handy to build the next agent's prompt.
            </span>
          </span>
        </div>
        <textarea
          :value="node.template"
          @input="$emit('update', { ...node, template: $event.target.value })"
          @pointerdown.stop
          rows="3" spellcheck="false"
          :placeholder="tmplPlaceholder"
          class="w-full flex-1 min-h-[3rem] bg-piedra-800 border border-piedra-700/50 rounded-lg px-2 py-1 text-[10px] font-mono text-arena-200 placeholder:text-arena-600 outline-none focus:border-teal-500/50 resize-none"
        />
        <div class="flex items-center gap-1.5">
          <span class="text-[9px] text-arena-500 whitespace-nowrap">save as</span>
          <input
            :value="node.outputKey"
            @input="$emit('update', { ...node, outputKey: $event.target.value })"
            @pointerdown.stop
            placeholder="state key (optional)"
            class="flex-1 min-w-0 bg-piedra-800 border border-piedra-700/50 rounded px-1.5 py-0.5 text-[10px] font-mono text-arena-200 placeholder:text-arena-600 outline-none focus:border-teal-500/50"
          />
        </div>
      </div>
    </div>

    <!-- input port (all node types). The drop target is the whole node (see
         FlowCanvas onCanvasPointerUp hit-test); this port is the visual cue. -->
    <span
      class="flow-port flow-port-in"
      :class="connectingActive ? 'flow-port-target' : ''"
      :data-port="'in:' + node.id"
      title="Drop a connection here"
    />

    <!-- single output port for agent / join -->
    <span
      v-if="node.type !== 'router'"
      class="flow-port flow-port-out flow-port-out-single"
      :data-port="'out:' + node.id + '|'"
      title="Drag to connect"
      @pointerdown.stop.prevent="$emit('start-edge', { nodeId: node.id, route: '' })"
    />

    <!-- resize handle (bottom-right) -->
    <span
      class="flow-resize opacity-0 group-hover:opacity-100"
      :class="{ 'opacity-100': resizing }"
      title="Drag to resize"
      @pointerdown.stop.prevent="onResizeDown"
    >
      <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.2">
        <path stroke-linecap="round" d="M9 3L3 9M9 6L6 9" />
      </svg>
    </span>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const NODE_W = 210

const props = defineProps({
  node:              { type: Object, required: true },
  agents:            { type: Array, default: () => [] },
  flows:             { type: Array, default: () => [] },
  isEntry:           { type: Boolean, default: false },
  selected:          { type: Boolean, default: false },
  connectingActive:  { type: Boolean, default: false },
})

const emit = defineEmits(['update', 'remove', 'start-edge', 'end-edge', 'pointerdown-body'])

const pickerOpen = ref(false)

// ── size / resize ────────────────────────────────────────────────────────────
const MIN_W = 160
const MIN_H = 90

// nodeStyle places and sizes the node. Width falls back to the default; height
// is only applied once the operator has resized the node (node.h set), so
// un-resized nodes keep their natural content height.
const nodeStyle = computed(() => {
  const s = {
    left: props.node.x + 'px',
    top: props.node.y + 'px',
    width: (props.node.w || NODE_W) + 'px',
  }
  if (props.node.h) s.height = props.node.h + 'px'
  return s
})

const resizing = ref(false)
let resizeStart = null // { px, py, w, h }

function onResizeDown(e) {
  const el = e.currentTarget.closest('.flow-node')
  resizing.value = true
  resizeStart = {
    px: e.clientX,
    py: e.clientY,
    w: props.node.w || (el ? el.offsetWidth : NODE_W),
    h: props.node.h || (el ? el.offsetHeight : MIN_H),
  }
  window.addEventListener('pointermove', onResizeMove)
  window.addEventListener('pointerup', onResizeUp)
}

function onResizeMove(e) {
  if (!resizeStart) return
  const w = Math.max(MIN_W, Math.round(resizeStart.w + (e.clientX - resizeStart.px)))
  const h = Math.max(MIN_H, Math.round(resizeStart.h + (e.clientY - resizeStart.py)))
  emit('update', { ...props.node, w, h })
}

function onResizeUp() {
  resizing.value = false
  resizeStart = null
  window.removeEventListener('pointermove', onResizeMove)
  window.removeEventListener('pointerup', onResizeUp)
}

function onBodyPointerDown(e) {
  emit('pointerdown-body', e)
}

// ── agent ────────────────────────────────────────────────────────────────
const agentName = computed(() => {
  const a = props.agents.find(a => a.id === props.node.agentId)
  return a?.name || props.node.agentId || 'Select agent...'
})
function pickAgent(id) {
  pickerOpen.value = false
  emit('update', { ...props.node, agentId: id })
}

// ── subflow ────────────────────────────────────────────────────────────────
const flowName = computed(() => {
  const fl = props.flows.find(f => f.id === props.node.flowId)
  return fl?.name || props.node.flowId || 'Select flow...'
})
function pickFlow(id) {
  pickerOpen.value = false
  emit('update', { ...props.node, flowId: id })
}

// ── transform node placeholders ──────────────────────────────────────────────
// Literal double-brace text is built at runtime so Vue does not treat it as an
// interpolation in the template source.
const mustacheInput = '{{ input }}'
const exprPlaceholder = 'input.split(",")'
const tmplPlaceholder = 'Summarise for {{ state.lang }}:\n{{ input }}'

// ── router rules ───────────────────────────────────────────────────────────
function updateRule(i, key, val) {
  const rules = (props.node.rules || []).map((r, idx) => idx === i ? { ...r, [key]: val } : r)
  emit('update', { ...props.node, rules })
}
function addRule() {
  const rules = [...(props.node.rules || []), { when: '', route: '' }]
  emit('update', { ...props.node, rules })
}
function removeRule(i) {
  emit('update', { ...props.node, rules: (props.node.rules || []).filter((_, idx) => idx !== i) })
}

// ── appearance per type ─────────────────────────────────────────────────────
// Class strings are written out literally (no runtime interpolation) so
// Tailwind's JIT scanner picks them up — `border-${color}-500/30` would be
// invisible to the scanner and produce no CSS.
const TYPE = {
  agent: {
    label: 'Agent',
    icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z',
    border: 'border-sol-500/30',
    hoverBorder: 'hover:border-sol-500/60',
    selectedRing: 'border-sol-500/70 ring-2 ring-sol-500/25',
    headerBg: 'bg-sol-500/8',
    iconBg: 'bg-sol-500/15',
    iconColor: 'text-sol-400',
    labelColor: 'text-sol-400',
  },
  router: {
    label: 'Router',
    icon: 'M3 12h4l3-9 4 18 3-9h4',
    border: 'border-atlantico-500/30',
    hoverBorder: 'hover:border-atlantico-500/60',
    selectedRing: 'border-atlantico-500/70 ring-2 ring-atlantico-500/25',
    headerBg: 'bg-atlantico-500/8',
    iconBg: 'bg-atlantico-500/15',
    iconColor: 'text-atlantico-400',
    labelColor: 'text-atlantico-400',
  },
  join: {
    label: 'Join',
    icon: 'M7 4v5a5 5 0 005 5 5 5 0 005-5V4M12 14v6',
    border: 'border-purple-500/30',
    hoverBorder: 'hover:border-purple-500/60',
    selectedRing: 'border-purple-500/70 ring-2 ring-purple-500/25',
    headerBg: 'bg-purple-500/8',
    iconBg: 'bg-purple-500/15',
    iconColor: 'text-purple-400',
    labelColor: 'text-purple-400',
  },
  parallel: {
    label: 'Parallel',
    icon: 'M4 6h16M4 12h16M4 18h16',
    border: 'border-lava-500/30',
    hoverBorder: 'hover:border-lava-500/60',
    selectedRing: 'border-lava-500/70 ring-2 ring-lava-500/25',
    headerBg: 'bg-lava-500/8',
    iconBg: 'bg-lava-500/15',
    iconColor: 'text-lava-400',
    labelColor: 'text-lava-400',
  },
  subflow: {
    label: 'Subflow',
    icon: 'M9 4H5a1 1 0 00-1 1v4m0 6v4a1 1 0 001 1h4m6-16h4a1 1 0 011 1v4m0 6v4a1 1 0 01-1 1h-4M9 9h6v6H9z',
    border: 'border-rose-500/30',
    hoverBorder: 'hover:border-rose-500/60',
    selectedRing: 'border-rose-500/70 ring-2 ring-rose-500/25',
    headerBg: 'bg-rose-500/8',
    iconBg: 'bg-rose-500/15',
    iconColor: 'text-rose-400',
    labelColor: 'text-rose-400',
  },
  expression: {
    label: 'Expression',
    icon: 'M8 9l3 3-3 3m5 0h3M4 4h16a1 1 0 011 1v14a1 1 0 01-1 1H4a1 1 0 01-1-1V5a1 1 0 011-1z',
    border: 'border-indigo-500/30',
    hoverBorder: 'hover:border-indigo-500/60',
    selectedRing: 'border-indigo-500/70 ring-2 ring-indigo-500/25',
    headerBg: 'bg-indigo-500/8',
    iconBg: 'bg-indigo-500/15',
    iconColor: 'text-indigo-400',
    labelColor: 'text-indigo-400',
  },
  template: {
    label: 'Template',
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    border: 'border-teal-500/30',
    hoverBorder: 'hover:border-teal-500/60',
    selectedRing: 'border-teal-500/70 ring-2 ring-teal-500/25',
    headerBg: 'bg-teal-500/8',
    iconBg: 'bg-teal-500/15',
    iconColor: 'text-teal-400',
    labelColor: 'text-teal-400',
  },
}
const cfg              = computed(() => TYPE[props.node.type] || TYPE.agent)
const typeLabel       = computed(() => cfg.value.label)
const iconPath        = computed(() => cfg.value.icon)
const borderClass     = computed(() => cfg.value.border)
const hoverBorderClass = computed(() => cfg.value.hoverBorder)
const selectedRingClass = computed(() => cfg.value.selectedRing)
const headerBgClass   = computed(() => cfg.value.headerBg)
const iconBgClass     = computed(() => cfg.value.iconBg)
const iconColorClass  = computed(() => cfg.value.iconColor)
const labelColorClass = computed(() => cfg.value.labelColor)

function portClass() {
  return ''
}
</script>

<style scoped>
.flow-node { transition: box-shadow 0.15s ease; }

/* resize handle, bottom-right corner */
.flow-resize {
  position: absolute;
  right: 2px;
  bottom: 2px;
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-arena-500);
  cursor: nwse-resize;
  transition: color 0.12s ease, opacity 0.12s ease;
  z-index: 6;
}
.flow-resize:hover { color: var(--color-arena-300); }
.flow-resize svg { width: 10px; height: 10px; }

/* connection ports */
.flow-port {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 9999px;
  background: var(--color-piedra-700);
  border: 2px solid var(--color-piedra-500);
  cursor: crosshair;
  transition: all 0.12s ease;
  z-index: 5;
}
.flow-port:hover {
  background: var(--color-sol-400);
  border-color: var(--color-sol-300);
  transform: scale(1.25);
}
.flow-port-in {
  left: -7px;
  top: 16px;
}
.flow-port-in.flow-port-target {
  background: var(--color-rose-400, #fb7185);
  border-color: var(--color-rose-300, #fda4af);
  box-shadow: 0 0 0 4px rgba(251, 113, 133, 0.2);
}
.flow-port-out { right: -7px; }
.flow-port-out-single { top: 16px; }
.flow-port-out:not(.flow-port-out-single) {
  position: absolute;
  right: -13px;
  top: 50%;
  transform: translateY(-50%);
}
.flow-port-out:not(.flow-port-out-single):hover {
  transform: translateY(-50%) scale(1.25);
}

.dropdown-enter-active { transition: all 0.15s ease-out; }
.dropdown-leave-active { transition: all 0.1s ease-in; }
.dropdown-enter-from,
.dropdown-leave-to     { opacity: 0; transform: translateY(-4px) scale(0.97); }
</style>
