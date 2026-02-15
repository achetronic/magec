<template>
  <canvas
    ref="canvasRef"
    class="w-full h-full"
    :class="store.centellaEnabled ? 'cursor-pointer' : 'cursor-not-allowed'"
    :style="{ opacity: store.centellaEnabled ? 1 : 0.5 }"
    @click="onClick"
  />
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useAppStore } from '../lib/stores/app.js'

const store = useAppStore()
const canvasRef = ref(null)

let ctx = null
let analyser = null
let resizeObserver = null
let animationId = null
let smoothedLevel = 0
let awakeness = 0
let smoothedRadius = 0
let time = 0
let particles = []

function onClick() {
  if (store.centellaEnabled) {
    store.toggleRecording()
  }
}

function setupCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return

  ctx = canvas.getContext('2d')

  const resize = () => {
    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width * devicePixelRatio
    canvas.height = rect.height * devicePixelRatio
    ctx.scale(devicePixelRatio, devicePixelRatio)
  }

  resize()
  resizeObserver = new ResizeObserver(resize)
  resizeObserver.observe(canvas)
}

function draw() {
  animationId = requestAnimationFrame(draw)

  const canvas = canvasRef.value
  if (!canvas || !analyser) return

  const w = canvas.offsetWidth
  const h = canvas.offsetHeight
  if (w === 0 || h === 0) return

  time += 0.008
  const centerX = w / 2
  const centerY = h / 2

  const maxDimension = Math.min(w, h, 384)
  const baseRadius = maxDimension * 0.25
  const maxRadius = maxDimension * 0.44

  const data = new Uint8Array(analyser.frequencyBinCount)
  analyser.getByteFrequencyData(data)

  let sum = 0
  const relevantBins = Math.min(48, data.length)
  for (let i = 0; i < relevantBins; i++) {
    const weight = 1 + (i < 16 ? 0.5 : 0)
    sum += data[i] * data[i] * weight
  }
  const raw = Math.sqrt(sum / relevantBins) / 255
  const target = Math.pow(raw, 0.7) * 1.4

  smoothedLevel += (target - smoothedLevel) * 0.25
  const level = Math.min(1, smoothedLevel)

  const isRec = store.isRecording
  const awakeTarget = isRec ? 1 : 0
  const awakeSpeed = isRec ? 0.15 : 0.03
  awakeness += (awakeTarget - awakeness) * awakeSpeed
  const awake = awakeness

  const dormantColor = { r: 251, g: 191, b: 36 }
  const awakeColor = { r: 239, g: 68, b: 68 }

  const color = {
    r: Math.round(dormantColor.r + (awakeColor.r - dormantColor.r) * awake),
    g: Math.round(dormantColor.g + (awakeColor.g - dormantColor.g) * awake),
    b: Math.round(dormantColor.b + (awakeColor.b - dormantColor.b) * awake)
  }

  const alphaBoost = 0.6 + awake * 0.4

  ctx.clearRect(0, 0, w, h)

  const ambientLevel = awake < 0.5 ? 0 : level
  const targetRadius = baseRadius + ambientLevel * (maxRadius - baseRadius)

  if (smoothedRadius === 0) smoothedRadius = baseRadius
  smoothedRadius += (targetRadius - smoothedRadius) * 0.08
  const radius = smoothedRadius

  const particleSpeed = 0.05 + awake * 0.15
  const particleDrift = 0.003 + awake * 0.007
  const particleDecay = 0.002 + awake * 0.006
  const spawnChance = 0.25 + awake * 0.4 + level * 0.3

  if (Math.random() < spawnChance) {
    const angle = Math.random() * Math.PI * 2
    const dist = Math.random() * radius * 0.7
    particles.push({
      x: centerX + Math.cos(angle) * dist,
      y: centerY + Math.sin(angle) * dist,
      vx: (Math.random() - 0.5) * particleSpeed,
      vy: (Math.random() - 0.5) * particleSpeed,
      life: 1,
      size: 1 + Math.random() * 2,
      drift: Math.random() * Math.PI * 2
    })
  }

  particles = particles.filter(p => {
    p.life -= particleDecay
    if (p.life <= 0) return false

    p.drift += particleDrift
    p.x += p.vx + Math.cos(p.drift) * particleSpeed
    p.y += p.vy + Math.sin(p.drift) * particleSpeed

    const dx = p.x - centerX
    const dy = p.y - centerY
    const dist = Math.sqrt(dx * dx + dy * dy)
    if (dist > radius * 0.85) {
      p.x = centerX + (dx / dist) * radius * 0.85
      p.y = centerY + (dy / dist) * radius * 0.85
    }

    const alpha = p.life * 0.6 * alphaBoost
    ctx.fillStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`

    ctx.beginPath()
    ctx.arc(p.x, p.y, p.size * p.life, 0, Math.PI * 2)
    ctx.fill()

    return true
  })

  const numSwirls = 3
  for (let s = 0; s < numSwirls; s++) {
    ctx.beginPath()
    const swirlOffset = (s / numSwirls) * Math.PI * 2
    const swirlRadius = radius * (0.3 + s * 0.15)

    for (let i = 0; i <= 60; i++) {
      const t = i / 60
      const angle = t * Math.PI * 2 + time * (1 + s * 0.3) + swirlOffset
      const wobble = Math.sin(t * Math.PI * 4 + time * 2) * (5 + level * 10)
      const r = swirlRadius + wobble

      const x = centerX + Math.cos(angle) * r
      const y = centerY + Math.sin(angle) * r

      if (i === 0) ctx.moveTo(x, y)
      else ctx.lineTo(x, y)
    }

    ctx.closePath()
    const swirlAlpha = (0.06 + level * 0.08) * alphaBoost
    ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${swirlAlpha})`
    ctx.lineWidth = 1
    ctx.stroke()
  }

  const coreGradient = ctx.createRadialGradient(centerX, centerY, 0, centerX, centerY, radius * 0.4)
  const coreAlpha = 0.12 + level * 0.08
  coreGradient.addColorStop(0, `rgba(${color.r}, ${color.g}, ${color.b}, ${coreAlpha * alphaBoost})`)
  coreGradient.addColorStop(1, `rgba(${color.r}, ${color.g}, ${color.b}, 0)`)
  ctx.fillStyle = coreGradient
  ctx.beginPath()
  ctx.arc(centerX, centerY, radius * 0.4, 0, Math.PI * 2)
  ctx.fill()

  const glowGradient = ctx.createRadialGradient(centerX, centerY, radius * 0.9, centerX, centerY, radius * 1.3)
  const glowAlpha = 0.15 * alphaBoost
  glowGradient.addColorStop(0, `rgba(${color.r}, ${color.g}, ${color.b}, ${glowAlpha})`)
  glowGradient.addColorStop(1, `rgba(${color.r}, ${color.g}, ${color.b}, 0)`)
  ctx.fillStyle = glowGradient
  ctx.beginPath()
  ctx.arc(centerX, centerY, radius * 1.3, 0, Math.PI * 2)
  ctx.fill()

  const breathe = Math.sin(time * 1.5) * 2
  ctx.beginPath()
  ctx.arc(centerX, centerY, radius + breathe, 0, Math.PI * 2)
  const strokeAlpha = 0.5 + awake * 0.35
  ctx.strokeStyle = `rgba(${color.r}, ${color.g}, ${color.b}, ${strokeAlpha})`
  ctx.lineWidth = 2
  ctx.stroke()
}

function tryAttachAnalyser() {
  const a = store.getAnalyser()
  if (a) {
    analyser = a
    draw()
    return true
  }
  return false
}

onMounted(() => {
  setupCanvas()

  if (!tryAttachAnalyser()) {
    const interval = setInterval(() => {
      if (tryAttachAnalyser()) {
        clearInterval(interval)
      }
    }, 200)

    setTimeout(() => clearInterval(interval), 15000)
  }
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (resizeObserver) resizeObserver.disconnect()
})
</script>
