package main

import (
	"math"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

// 将点轨迹简化，这里的阈值其实就是可以简化为直线的的距离，设置为10比较合适
func simplifyDouglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// 找到距离最大的点
	maxDistance := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		distance := distanceToSegment(points[i], points[0], points[len(points)-1])
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	// 如果最大距离小于阈值，直接返回首尾点
	if maxDistance <= tolerance {
		return []Point{points[0], points[len(points)-1]}
	}

	// 递归简化
	leftPart := simplifyDouglasPeucker(points[:maxIndex+1], tolerance)
	rightPart := simplifyDouglasPeucker(points[maxIndex:], tolerance)

	return append(leftPart[:len(leftPart)-1], rightPart...)
}

// 比较：
// 投影点算法：
// 优点：简单直观，不涉及面积和开根号等复杂运算，计算相对位置 u 相对高效。
// 缺点：可能对数值精度较为敏感，特别是在点 p 与线段 start 或 end 非常接近的情况下，可能引入一些误差。

// 海伦公式：
// 优点：适用于计算点到任意线段的距离，不仅仅是投影点。在某些情况下，这个方法可能更准确。
// 缺点：涉及到开根号等运算，相对于点乘法可能计算开销更大。

// 选择建议：
// 如果只需计算点到线段上的投影点，而不需要计算点到线段的实际距离，那么点乘法可能更简单和高效。
// 如果需要更通用的点到线段的距离计算，海伦公式可能提供更准确的结果，但在计算开销方面可能更大。
// 总体而言，对于许多应用场景来说，点乘法足够简单和有效，但具体选择还是要根据实际需求和性能要求来决定。
// 计算点到首尾连线的的距离

func distanceToSegment(p, start, end Point) float64 {
	lineLength := distanceBetweenPoints(start, end)

	// 首尾重合，
	if lineLength == 0 {
		return distanceBetweenPoints(p, start)
	}

	// 使用点乘求解投影比例
	u := ((p.Latitude-start.Latitude)*(end.Latitude-start.Latitude) +
		(p.Longitude-start.Longitude)*(end.Longitude-start.Longitude)) / (lineLength * lineLength)

	//在这里，u 表示点 p 在线段 start 和 end 上的投影点相对于线段的位置。如果 u 小于0或大于1，表示投影点在线段的延长线上而不在线段上。
	//如果 u < 0，则表示投影点在线段 start 之前，那么最近的点就是线段的起点 start。
	//如果 u > 1，则表示投影点在线段 end 之后，那么最近的点就是线段的终点 end。
	//因此，这段代码的目的是找到距离点 p 最近的在线段上的点。如果投影点在线段之前或之后，就选择线段的起点或终点作为最近的点，
	//然后计算点 p 到这个最近点的距离，即 distanceBetweenPoints(p, closestPoint)。
	// <-----0    1 ------->
	if u < 0 || u > 1 {
		closestPoint := start
		if u > 0 {
			closestPoint = end
		}
		return distanceBetweenPoints(p, closestPoint)
	}

	// 0<=  u <= 1 先找到投影点，然后再计算距离
	intersection := Point{
		Latitude:  start.Latitude + u*(end.Latitude-start.Latitude),
		Longitude: start.Longitude + u*(end.Longitude-start.Longitude),
	}

	return distanceBetweenPoints(p, intersection)
}

// 计算2点之间距离
func distanceBetweenPoints(p1, p2 Point) float64 {
	// 使用简化的球面距离计算
	radius := 6371.0 // 地球半径，单位：千米
	dLat := degToRad(p2.Latitude - p1.Latitude)
	dLon := degToRad(p2.Longitude - p1.Longitude)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degToRad(p1.Latitude))*math.Cos(degToRad(p2.Latitude))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := radius * c
	return distance
}

// 角度转为弧度
func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180)
}
