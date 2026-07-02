package util

import "math"

// ComputeRMS 计算振动采样序列的均方根值 (RMS)，反映振动能量整体水平。
func ComputeRMS(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sumSq float64
	for _, v := range samples {
		sumSq += v * v
	}
	return math.Sqrt(sumSq / float64(len(samples)))
}

// ComputePeak 计算振动采样序列的峰值（绝对值最大值）。
func ComputePeak(samples []float64) float64 {
	var peak float64
	for _, v := range samples {
		if av := math.Abs(v); av > peak {
			peak = av
		}
	}
	return peak
}

// CrestFactor 计算峰值因子 = 峰值 / RMS，是轴承早期点蚀等冲击性故障的敏感指标。
// RMS 为 0（无振动或数据缺失）时返回 0，避免除零。
func CrestFactor(peak, rms float64) float64 {
	if rms == 0 {
		return 0
	}
	return peak / rms
}

// DominantFrequency 通过简化离散傅里叶变换（DFT）估算振动信号主频。
// 采用直接求和法而非 FFT，适合边缘设备低频小批量采样场景；采样点数较大时性能可接受即可，
// 平台侧数据量通常为几十到几百点，无需引入 FFT 依赖。
// sampleRateHz <= 0 或采样点数不足时返回 0。
func DominantFrequency(samples []float64, sampleRateHz float64) float64 {
	n := len(samples)
	if n < 2 || sampleRateHz <= 0 {
		return 0
	}

	maxMag := -1.0
	bestK := 0
	// k=0 为直流分量，跳过；k 取到 n/2 满足奈奎斯特采样定理
	for k := 1; k <= n/2; k++ {
		var re, im float64
		for t, v := range samples {
			angle := 2 * math.Pi * float64(k) * float64(t) / float64(n)
			re += v * math.Cos(angle)
			im -= v * math.Sin(angle)
		}
		mag := math.Hypot(re, im)
		if mag > maxMag {
			maxMag = mag
			bestK = k
		}
	}
	return float64(bestK) * sampleRateHz / float64(n)
}

// ClassifyBearingFault 基于主频与峰值因子的简化启发式规则，粗略判断轴承故障特征类型。
// 精确的外圈/内圈/滚珠故障频率需要结合轴承几何参数和转速计算特征频率（BPFO/BPFI/BSF），
// 此处在缺少转速遥测的情况下采用频段+冲击强度的简化代理指标，仅作为预警参考，非精确诊断。
func ClassifyBearingFault(dominantFreqHz, crestFactor float64) string {
	const impactThreshold = 4.0 // 峰值因子超过该值视为存在冲击性振动分量
	if crestFactor < impactThreshold || dominantFreqHz <= 0 {
		return "none"
	}
	switch {
	case dominantFreqHz < 150:
		return "outer_race"
	case dominantFreqHz < 300:
		return "inner_race"
	default:
		return "ball"
	}
}
