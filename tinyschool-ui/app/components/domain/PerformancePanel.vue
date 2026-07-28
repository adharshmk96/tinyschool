<script setup lang="ts">
type Mode = 'student' | 'assignment' | 'exam' | 'class'
type RankedStudent = { id: string, name: string, score: number }
type TrendPoint = { label: string, value: number }
type PerformanceResult = { name: string, kind: string, score: number }

const props = withDefaults(defineProps<{
  subject?: string
  data?: unknown
  mode?: Mode
}>(), {
  mode: 'student'
})

const performance = computed(() => {
  if (!props.data || typeof props.data !== 'object')
    return null

  const value = props.data as Record<string, unknown>
  const number = (key: string) => Number(value[key] ?? 0)
  return {
    averageScore: number('averageScore'),
    classAverage: number('classAverage'),
    completionRate: number('completionRate'),
    completed: number('completed'),
    total: number('total'),
    standing: String(value.standing ?? 'Not ranked'),
    trend: Array.isArray(value.trend) ? value.trend.map(Number).filter(Number.isFinite) : [],
    highestScore: number('highestScore'),
    lowestScore: number('lowestScore'),
    medianScore: number('medianScore'),
    assignmentCompleted: number('assignmentCompleted'),
    assignmentTotal: number('assignmentTotal'),
    assignmentCompletion: number('assignmentCompletion'),
    examCompleted: number('examCompleted'),
    examTotal: number('examTotal'),
    examCompletion: number('examCompletion'),
    assignmentAverage: number('assignmentAverage'),
    examAverage: number('examAverage'),
    strongestArea: String(value.strongestArea ?? 'Waiting for scores'),
    scoreConsistency: number('scoreConsistency'),
    bestResult: value.bestResult && typeof value.bestResult === 'object' ? value.bestResult as PerformanceResult : null,
    topStudents: Array.isArray(value.topStudents) ? value.topStudents as RankedStudent[] : [],
    monthlyExamTrend: Array.isArray(value.monthlyExamTrend) ? value.monthlyExamTrend as TrendPoint[] : [],
    monthlyAssignmentTrend: Array.isArray(value.monthlyAssignmentTrend) ? value.monthlyAssignmentTrend as TrendPoint[] : []
  }
})

const heading = computed(() => {
  if (props.mode === 'assignment')
    return 'Assignment performance'
  if (props.mode === 'exam')
    return 'Exam performance'
  return props.mode === 'class' ? 'Class performance' : 'Performance overview'
})
const description = computed(() => {
  if (props.mode === 'assignment')
    return 'Score summary, submission progress and strongest results for this assignment.'
  if (props.mode === 'exam')
    return 'Score summary, marking progress and strongest results for this exam.'
  if (props.mode === 'class')
    return 'Assignment and exam performance for the current academic year.'
  return 'Current-year scores, completion and progress over time.'
})
</script>

<template>
  <TabPlaceholder
    v-if="!performance"
    icon="i-lucide-chart-no-axes-combined"
    title="No performance data available"
    description="Scores and completion insights will appear here after work has been marked."
  />
  <section v-else>
    <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <span class="grid size-9 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              name="i-lucide-chart-no-axes-combined"
              class="size-4"
            />
          </span>
          <h2 class="text-lg font-semibold text-highlighted">
            {{ heading }}
          </h2>
        </div>
        <p class="mt-2 text-sm leading-6 text-muted">
          {{ description }}
        </p>
      </div>
      <UBadge
        :label="mode === 'exam' ? 'Single exam' : mode === 'assignment' ? 'Single assignment' : 'Current year'"
        color="neutral"
        variant="subtle"
        icon="i-lucide-calendar-check"
        class="self-start"
      />
    </div>

    <div
      v-if="mode === 'exam' || mode === 'assignment'"
      class="grid gap-5 sm:grid-cols-2 xl:grid-cols-4"
    >
      <UCard>
        <p class="text-sm text-muted">
          Average score
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.averageScore }}%
        </p>
        <UBadge
          class="mt-3"
          color="success"
          variant="subtle"
          :label="performance.standing"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          {{ mode === 'assignment' ? 'Submission completion' : 'Score completion' }}
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.completed }} / {{ performance.total }}
        </p>
        <UProgress
          :model-value="performance.completionRate"
          class="mt-4"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Highest / lowest
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.highestScore }}% <span class="text-lg text-muted">/ {{ performance.lowestScore }}%</span>
        </p>
        <p class="mt-3 text-sm text-muted">
          Range across marked scores.
        </p>
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Median score
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.medianScore }}%
        </p>
        <p class="mt-3 text-sm text-muted">
          Middle result, less affected by outliers.
        </p>
      </UCard>
    </div>

    <div
      v-else-if="mode === 'class'"
      class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3"
    >
      <UCard>
        <p class="text-sm text-muted">
          Average score
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.averageScore }}%
        </p>
        <UBadge
          class="mt-3"
          color="success"
          variant="subtle"
          :label="performance.standing"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Assignments completed
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.assignmentCompleted }} / {{ performance.assignmentTotal }}
        </p>
        <UProgress
          :model-value="performance.assignmentCompletion"
          class="mt-4"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Exam scores completed
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.examCompleted }} / {{ performance.examTotal }}
        </p>
        <UProgress
          :model-value="performance.examCompletion"
          class="mt-4"
        />
      </UCard>
    </div>

    <div
      v-else
      class="grid gap-5 sm:grid-cols-2 xl:grid-cols-4"
    >
      <UCard>
        <p class="text-sm text-muted">
          Average score
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.averageScore }}%
        </p>
        <UBadge
          class="mt-3"
          color="success"
          variant="subtle"
          :label="performance.classAverage ? `Class average ${performance.classAverage}%` : 'Current average'"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Combined completion
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.completed }} / {{ performance.total }}
        </p>
        <UProgress
          :model-value="performance.completionRate"
          class="mt-4"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Strongest area
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.strongestArea }}
        </p>
        <p class="mt-3 text-sm text-muted">
          Compares marked assignment and exam averages.
        </p>
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Score consistency
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.scoreConsistency }}%
        </p>
        <p class="mt-3 text-sm text-muted">
          Higher means results stay closer to the combined average.
        </p>
      </UCard>
    </div>

    <div
      v-if="mode !== 'student'"
      class="mt-5 grid gap-5 lg:grid-cols-5"
    >
      <UCard class="lg:col-span-2">
        <h3 class="font-semibold text-highlighted">
          Top scoring students
        </h3>
        <div
          v-if="performance.topStudents.length"
          class="mt-4 space-y-3"
        >
          <div
            v-for="(student, index) in performance.topStudents"
            :key="student.id"
            class="flex items-center gap-3"
          >
            <span class="grid size-8 shrink-0 place-items-center rounded-full bg-primary/10 text-sm font-semibold text-primary">{{ index + 1 }}</span>
            <p class="min-w-0 flex-1 truncate text-sm font-medium text-highlighted">
              {{ student.name }}
            </p>
            <UBadge
              color="neutral"
              variant="subtle"
              :label="`${student.score}%`"
            />
          </div>
        </div>
        <p
          v-else
          class="mt-4 text-sm text-muted"
        >
          Top students will appear after scores are recorded.
        </p>
      </UCard>

      <UCard
        v-if="mode === 'class'"
        class="lg:col-span-3"
      >
        <h3 class="font-semibold text-highlighted">
          Monthly exam trend
        </h3>
        <div
          v-if="performance.monthlyExamTrend.length"
          class="mt-5 flex h-40 items-end gap-3"
        >
          <div
            v-for="point in performance.monthlyExamTrend"
            :key="point.label"
            class="flex min-w-0 flex-1 flex-col items-center justify-end gap-2"
          >
            <span class="text-xs font-medium text-highlighted">{{ point.value }}%</span>
            <div
              class="w-full rounded-t bg-primary/80"
              :style="{ height: `${Math.max(point.value, 3)}%` }"
              :title="`${point.label}: ${point.value}%`"
            />
            <span class="w-full truncate text-center text-xs text-muted">{{ point.label }}</span>
          </div>
        </div>
        <TabPlaceholder
          v-else
          icon="i-lucide-chart-spline"
          title="No exam trend yet"
          description="Monthly averages will appear after exam scores are recorded."
          badge="Waiting for scores"
        />
      </UCard>

      <UCard
        v-else
        class="lg:col-span-3"
      >
        <h3 class="font-semibold text-highlighted">
          Score spread
        </h3>
        <p class="mt-2 text-sm text-muted">
          The middle 50% of results centers around the {{ performance.medianScore }}% median.
        </p>
        <div class="mt-6">
          <div class="relative h-3 rounded-full bg-muted/20">
            <div
              class="absolute h-3 rounded-full bg-primary/70"
              :style="{ left: `${performance.lowestScore}%`, right: `${100 - performance.highestScore}%` }"
            />
            <div
              class="absolute -top-1 size-5 -translate-x-1/2 rounded-full border-4 border-white bg-primary shadow"
              :style="{ left: `${performance.medianScore}%` }"
            />
          </div>
          <div class="mt-3 flex justify-between text-xs text-muted">
            <span>{{ performance.lowestScore }}% low</span><span>{{ performance.highestScore }}% high</span>
          </div>
        </div>
      </UCard>
    </div>

    <div
      v-if="mode === 'student'"
      class="mt-5 grid gap-5 lg:grid-cols-2"
    >
      <UCard>
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="font-semibold text-highlighted">
              Assignments
            </h3>
            <p class="mt-1 text-sm text-muted">
              {{ performance.assignmentCompleted }} / {{ performance.assignmentTotal }} completed
            </p>
          </div>
          <p class="text-2xl font-semibold text-highlighted">
            {{ performance.assignmentAverage }}%
          </p>
        </div>
        <UProgress
          :model-value="performance.assignmentCompletion"
          class="mt-4"
        />
        <div
          v-if="performance.monthlyAssignmentTrend.length"
          class="mt-6 flex h-40 items-end gap-3"
        >
          <div
            v-for="point in performance.monthlyAssignmentTrend"
            :key="point.label"
            class="flex min-w-0 flex-1 flex-col items-center justify-end gap-2"
          >
            <span class="text-xs font-medium text-highlighted">{{ point.value }}%</span>
            <div
              class="w-full rounded-t bg-primary/80"
              :style="{ height: `${Math.max(point.value, 3)}%` }"
              :title="`${point.label}: ${point.value}%`"
            />
            <span class="w-full truncate text-center text-xs text-muted">{{ point.label }}</span>
          </div>
        </div>
        <TabPlaceholder
          v-else
          icon="i-lucide-clipboard-check"
          title="No assignment trend yet"
          description="Monthly assignment averages appear after scores are recorded."
          badge="Waiting for scores"
        />
      </UCard>

      <UCard>
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="font-semibold text-highlighted">
              Exams
            </h3>
            <p class="mt-1 text-sm text-muted">
              {{ performance.examCompleted }} / {{ performance.examTotal }} completed
            </p>
          </div>
          <p class="text-2xl font-semibold text-highlighted">
            {{ performance.examAverage }}%
          </p>
        </div>
        <UProgress
          :model-value="performance.examCompletion"
          class="mt-4"
        />
        <div
          v-if="performance.monthlyExamTrend.length"
          class="mt-6 flex h-40 items-end gap-3"
        >
          <div
            v-for="point in performance.monthlyExamTrend"
            :key="point.label"
            class="flex min-w-0 flex-1 flex-col items-center justify-end gap-2"
          >
            <span class="text-xs font-medium text-highlighted">{{ point.value }}%</span>
            <div
              class="w-full rounded-t bg-info/80"
              :style="{ height: `${Math.max(point.value, 3)}%` }"
              :title="`${point.label}: ${point.value}%`"
            />
            <span class="w-full truncate text-center text-xs text-muted">{{ point.label }}</span>
          </div>
        </div>
        <TabPlaceholder
          v-else
          icon="i-lucide-file-check-2"
          title="No exam trend yet"
          description="Monthly exam averages appear after scores are recorded."
          badge="Waiting for scores"
        />
      </UCard>

      <UCard class="lg:col-span-2">
        <div class="flex items-center gap-3">
          <span class="grid size-10 place-items-center rounded-lg bg-warning/10 text-warning">
            <UIcon
              name="i-lucide-trophy"
              class="size-5"
            />
          </span>
          <div v-if="performance.bestResult">
            <p class="text-sm text-muted">
              Best result
            </p>
            <p class="font-semibold text-highlighted">
              {{ performance.bestResult.name }} · {{ performance.bestResult.score }}%
            </p>
            <p class="text-xs text-muted">
              {{ performance.bestResult.kind }}
            </p>
          </div>
          <div v-else>
            <p class="font-semibold text-highlighted">
              No best result yet
            </p>
            <p class="text-sm text-muted">
              This insight appears after the first score.
            </p>
          </div>
        </div>
      </UCard>
    </div>
  </section>
</template>
