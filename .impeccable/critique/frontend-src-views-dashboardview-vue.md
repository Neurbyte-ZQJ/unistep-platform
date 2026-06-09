---
target: "frontend/src/views/DashboardView.vue"
total_score: 16
p0_count: 1
p1_count: 3
date: 2026-06-06
---

# DashboardView.vue Critique

## Design Health Score: 16/40 (Poor)

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 1 | 无加载态、无错误态 |
| 2 | Match System / Real World | 3 | 角色标签和文案自然 |
| 3 | User Control and Freedom | 2 | 快捷入口有限 |
| 4 | Consistency and Standards | 2 | 颜色硬编码，未用 tokens |
| 5 | Error Prevention | 1 | API 失败无用户感知 |
| 6 | Recognition Rather Than Recall | 2 | 概览卡片无图标 |
| 7 | Flexibility and Efficiency | 1 | 无快捷键、无自定义 |
| 8 | Aesthetic and Minimalist Design | 2 | 卡片网格单调 |
| 9 | Error Recovery | 1 | 无重试、无降级 |
| 10 | Help and Documentation | 1 | 无帮助入口 |

## Priority Issues

- [P0] 无加载态和错误态
- [P1] 概览卡片是 hero-metric 模板
- [P1] 颜色全部硬编码，未使用设计系统 tokens
- [P1] 图表卡片视觉节奏单调
- [P2] Dashboard 标题区视觉弱
