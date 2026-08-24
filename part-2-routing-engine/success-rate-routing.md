# 第 8 章 Success-Rate Routing

> 状态：章节骨架

## 章节定位

Provider 的总体 Authorization Rate 对单笔交易往往没有预测力。本章围绕交易细分、统计窗口与反馈延迟，构建可在线使用的 Expected Success Rate，而不是把一张按 Provider 聚合的排行榜当作智能路由。

## 本章将涵盖

* Provider、Issuer Country、BIN/IIN、Card Brand、Currency、Payment Method 等分桶维度。
* 成功与失败口径、软拒绝与技术错误是否进入分母。
* Sliding Window、衰减权重、最小样本量与置信度。
* 高维稀疏、Cold Start、数据延迟和时间段效应。
* 分层回退与 Bayesian Smoothing 等稳健估计方法。
* 指标存储、流式聚合、缓存和在线读取的时效预算。

## 推导路线

本章会从全局平均成功率开始，通过欧洲 Visa 与巴西 Mastercard 的反例拆解聚合偏差；随后不断细分，但又遇到样本稀疏和噪声，最终得到带样本量、置信度、窗口和回退层级的统计信号。

## 与前后章节的关系

上一章给出成本信号，本章给出成功概率信号。下一章会把 Eligibility、Rule、Cost、Success、Risk 与 Health 串成完整决策流水线，并说明成功率最高的路径也未必是风险调整后的最佳路径。
