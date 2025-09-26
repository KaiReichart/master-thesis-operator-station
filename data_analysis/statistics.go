package data_analysis

import (
	"math"
)

// FlightStatistics represents statistical analysis of flight data
type FlightStatistics struct {
	AirspeedStats       *DataStatistics `json:"airspeed_stats"`
	IndicatedAltitudeStats *DataStatistics `json:"indicated_altitude_stats"`
	AltitudeStats       *DataStatistics `json:"altitude_stats"`
	PressureAltitudeStats *DataStatistics `json:"pressure_altitude_stats"`
}

// DataStatistics represents statistical measures for a data series
type DataStatistics struct {
	Count      int     `json:"count"`
	Mean       float64 `json:"mean"`
	Variance   float64 `json:"variance"`
	StdDev     float64 `json:"std_dev"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Range      float64 `json:"range"`
	Median     float64 `json:"median"`
}

// CalculateFlightStatistics calculates comprehensive statistics for flight data
func CalculateFlightStatistics(flightData *FlightData) map[string]*FlightStatistics {
	result := make(map[string]*FlightStatistics)

	for aircraftLabel, positionData := range flightData.PositionData {
		if len(positionData) == 0 {
			continue
		}

		// Extract data series
		airspeeds := make([]float64, 0, len(positionData))
		indicatedAltitudes := make([]float64, 0, len(positionData))
		altitudes := make([]float64, 0, len(positionData))
		pressureAltitudes := make([]float64, 0, len(positionData))

		for _, point := range positionData {
			if point.Airspeed > 0 { // Only include positive airspeed values
				airspeeds = append(airspeeds, point.Airspeed)
			}
			if point.IndicatedAltitude != 0 { // Only include non-zero altitude values
				indicatedAltitudes = append(indicatedAltitudes, point.IndicatedAltitude)
			}
			if point.Altitude != 0 {
				altitudes = append(altitudes, point.Altitude)
			}
			if point.PressureAltitude != 0 {
				pressureAltitudes = append(pressureAltitudes, point.PressureAltitude)
			}
		}

		// Calculate statistics
		stats := &FlightStatistics{}
		
		if len(airspeeds) > 0 {
			stats.AirspeedStats = calculateDataStatistics(airspeeds)
		}
		if len(indicatedAltitudes) > 0 {
			stats.IndicatedAltitudeStats = calculateDataStatistics(indicatedAltitudes)
		}
		if len(altitudes) > 0 {
			stats.AltitudeStats = calculateDataStatistics(altitudes)
		}
		if len(pressureAltitudes) > 0 {
			stats.PressureAltitudeStats = calculateDataStatistics(pressureAltitudes)
		}

		result[aircraftLabel] = stats
	}

	return result
}

// calculateDataStatistics calculates comprehensive statistics for a data series
func calculateDataStatistics(data []float64) *DataStatistics {
	if len(data) == 0 {
		return nil
	}

	// Sort data for median calculation
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	quickSort(sortedData, 0, len(sortedData)-1)

	// Basic statistics
	count := len(data)
	sum := 0.0
	min := sortedData[0]
	max := sortedData[len(sortedData)-1]

	// Calculate mean
	for _, value := range data {
		sum += value
	}
	mean := sum / float64(count)

	// Calculate variance and standard deviation
	sumSquaredDiff := 0.0
	for _, value := range data {
		diff := value - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(count)
	stdDev := math.Sqrt(variance)

	// Calculate median
	var median float64
	if count%2 == 0 {
		median = (sortedData[count/2-1] + sortedData[count/2]) / 2
	} else {
		median = sortedData[count/2]
	}

	return &DataStatistics{
		Count:    count,
		Mean:     mean,
		Variance: variance,
		StdDev:   stdDev,
		Min:      min,
		Max:      max,
		Range:    max - min,
		Median:   median,
	}
}

// quickSort implements quicksort algorithm for sorting float64 slices
func quickSort(arr []float64, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

// partition is a helper function for quicksort
func partition(arr []float64, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// CalculateVarianceOverTime calculates variance for time windows
func CalculateVarianceOverTime(data []float64, windowSize int) []float64 {
	if len(data) < windowSize || windowSize <= 1 {
		return []float64{}
	}

	variances := make([]float64, 0, len(data)-windowSize+1)

	for i := 0; i <= len(data)-windowSize; i++ {
		window := data[i : i+windowSize]
		stats := calculateDataStatistics(window)
		if stats != nil {
			variances = append(variances, stats.Variance)
		}
	}

	return variances
}