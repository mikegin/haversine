package utils

import "math"

const (
	EARTH_RADIUS = 6372.8
)

func ReferenceHaversine(x0 float64, y0 float64, x1 float64, y1 float64, earthRadius float64) float64 {
	lat1 := y0
	lat2 := y1
	lon1 := x0
	lon2 := x1

	dLat := RadiansFromDegrees(lat2 - lat1)
	dLon := RadiansFromDegrees(lon2 - lon1)
	lat1 = RadiansFromDegrees(lat1)
	lat2 = RadiansFromDegrees(lat2)

	a := Square(math.Sin(dLat/2.0)) + math.Cos(lat1)*math.Cos(lat2)*Square(math.Sin(dLon/2))
	c := 2.0 * math.Asin(math.Sqrt(a))

	result := earthRadius * c

	return result
}

func RadiansFromDegrees(degrees float64) float64 {
	return 0.01745329251994329577 * degrees
}

func Square(a float64) float64 {
	return a * a
}
