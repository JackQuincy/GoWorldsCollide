package utils

import (
	"math"

	"github.com/JackQuincy/GoWorldsCollide/internal/random"
)

// TruncatedDiscreteDistribution samples a rounded normal distribution until the
// result lies within the optional inclusive bounds.
func TruncatedDiscreteDistribution(mean, stddev float64, minimum, maximum *int) int {
	return TruncatedDiscreteDistributionFrom(randomDefaultGaussian{}, mean, stddev, minimum, maximum)
}

// TruncatedDiscreteDistributionFrom samples using r.
func TruncatedDiscreteDistributionFrom(
	r Gaussian,
	mean, stddev float64,
	minimum, maximum *int,
) int {
	if minimum != nil && maximum != nil && *minimum > *maximum {
		panic("utils: distribution minimum exceeds maximum")
	}

	for {
		result := roundToEven(r.Gauss(mean, stddev))
		if minimum != nil && result < *minimum {
			continue
		}
		if maximum != nil && result > *maximum {
			continue
		}
		return result
	}
}

// Gaussian is the random behavior required by the truncated distribution.
type Gaussian interface {
	Gauss(mu, sigma float64) float64
}

type randomDefaultGaussian struct{}

func (randomDefaultGaussian) Gauss(mu, sigma float64) float64 {
	return random.Gauss(mu, sigma)
}

func roundToEven(value float64) int {
	return int(math.RoundToEven(value))
}
