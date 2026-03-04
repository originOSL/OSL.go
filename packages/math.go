// name: math
// description: Advanced mathematical utilities
// author: roturbot
// requires: math, math/cmplx, math/rand

type Math struct{}

func (Math) abs(x any) float64 {
	return math.Abs(OSLcastNumber(x))
}

func (Math) ceil(x any) float64 {
	return math.Ceil(OSLcastNumber(x))
}

func (Math) floor(x any) float64 {
	return math.Floor(OSLcastNumber(x))
}

func (Math) round(x any) float64 {
	return math.Round(OSLcastNumber(x))
}

func (Math) trunc(x any) float64 {
	return math.Trunc(OSLcastNumber(x))
}

func (Math) sqrt(x any) float64 {
	return math.Sqrt(OSLcastNumber(x))
}

func (Math) cbrt(x any) float64 {
	return math.Cbrt(OSLcastNumber(x))
}

func (Math) pow(base any, exp any) float64 {
	return math.Pow(OSLcastNumber(base), OSLcastNumber(exp))
}

func (Math) exp(x any) float64 {
	return math.Exp(OSLcastNumber(x))
}

func (Math) log(x any) float64 {
	return math.Log(OSLcastNumber(x))
}

func (Math) log10(x any) float64 {
	return math.Log10(OSLcastNumber(x))
}

func (Math) log2(x any) float64 {
	return math.Log2(OSLcastNumber(x))
}

func (Math) sin(x any) float64 {
	return math.Sin(OSLcastNumber(x))
}

func (Math) cos(x any) float64 {
	return math.Cos(OSLcastNumber(x))
}

func (Math) tan(x any) float64 {
	return math.Tan(OSLcastNumber(x))
}

func (Math) asin(x any) float64 {
	return math.Asin(OSLcastNumber(x))
}

func (Math) acos(x any) float64 {
	return math.Acos(OSLcastNumber(x))
}

func (Math) atan(x any) float64 {
	return math.Atan(OSLcastNumber(x))
}

func (Math) atan2(y any, x any) float64 {
	return math.Atan2(OSLcastNumber(y), OSLcastNumber(x))
}

func (Math) sinh(x any) float64 {
	return math.Sinh(OSLcastNumber(x))
}

func (Math) cosh(x any) float64 {
	return math.Cosh(OSLcastNumber(x))
}

func (Math) tanh(x any) float64 {
	return math.Tanh(OSLcastNumber(x))
}

func (Math) min(a any, b any) float64 {
	return math.Min(OSLcastNumber(a), OSLcastNumber(b))
}

func (Math) max(a any, b any) float64 {
	return math.Max(OSLcastNumber(a), OSLcastNumber(b))
}

func (Math) clamp(value any, min any, max any) float64 {
	val := OSLcastNumber(value)
	minVal := OSLcastNumber(min)
	maxVal := OSLcastNumber(max)

	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

func (Math) lerp(start any, end any, t any) float64 {
	startVal := OSLcastNumber(start)
	endVal := OSLcastNumber(end)
	tVal := OSLcastNumber(t)

	return startVal + (endVal-startVal)*tVal
}

func (Math) sum(numbers []any) float64 {
	result := 0.0
	for _, n := range numbers {
		result += OSLcastNumber(n)
	}
	return result
}

func (Math) avg(numbers []any) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return math.sum(numbers) / float64(len(numbers))
}

func (Math) median(numbers []any) float64 {
	nums := make([]float64, len(numbers))
	for i, n := range numbers {
		nums[i] = OSLcastNumber(n)
	}

	sort.Float64s(nums)
	mid := len(nums) / 2

	if len(nums)%2 == 0 {
		return (nums[mid-1] + nums[mid]) / 2
	}
	return nums[mid]
}

func (Math) mode(numbers []any) []any {
	counts := make(map[float64]int)
	nums := make([]float64, len(numbers))

	for i, n := range numbers {
		val := OSLcastNumber(n)
		nums[i] = val
		counts[val]++
	}

	maxCount := 0
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}

	var modes []any
	for val, count := range counts {
		if count == maxCount {
			modes = append(modes, val)
		}
	}

	return modes
}

func (Math) stdDev(numbers []any) float64 {
	if len(numbers) < 2 {
		return 0
	}

	mean := math.avg(numbers)
	variance := 0.0

	for _, n := range numbers {
		diff := OSLcastNumber(n) - mean
		variance += diff * diff
	}

	variance /= float64(len(numbers) - 1)
	return math.Sqrt(variance)
}

func (Math) variance(numbers []any) float64 {
	if len(numbers) < 2 {
		return 0
	}

	mean := math.avg(numbers)
	variance := 0.0

	for _, n := range numbers {
		diff := OSLcastNumber(n) - mean
		variance += diff * diff
	}

	return variance / float64(len(numbers))
}

func (Math) range(numbers []any) float64 {
	if len(numbers) == 0 {
		return 0
	}

	minVal := OSLcastNumber(numbers[0])
	maxVal := minVal

	for _, n := range numbers {
		val := OSLcastNumber(n)
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	return maxVal - minVal
}

func (Math) factorial(n any) int {
	nVal := int(OSLcastNumber(n))
	if nVal < 0 {
		return 0
	}
	if nVal == 0 || nVal == 1 {
		return 1
	}

	result := 1
	for i := 2; i <= nVal; i++ {
		result *= i
	}

	return result
}

func (Math) fibonacci(n any) int {
	nVal := int(OSLcastNumber(n))
	if nVal <= 0 {
		return 0
	}
	if nVal == 1 {
		return 1
	}

	a, b := 0, 1
	for i := 2; i <= nVal; i++ {
		a, b = b, a+b
	}

	return b
}

func (Math) gcd(a any, b any) int {
	aVal := int(OSLcastNumber(a))
	bVal := int(OSLcastNumber(b))

	for bVal != 0 {
		aVal, bVal = bVal, aVal%bVal
	}

	return aVal
}

func (Math) lcm(a any, b any) int {
	aVal := int(OSLcastNumber(a))
	bVal := int(OSLcastNumber(b))

	if aVal == 0 || bVal == 0 {
		return 0
	}

	return (aVal * bVal) / math.gcd(aVal, bVal)
}

func (Math) isPrime(n any) bool {
	nVal := int(OSLcastNumber(n))
	if nVal < 2 {
		return false
	}
	if nVal == 2 {
		return true
	}
	if nVal%2 == 0 {
		return false
	}

	for i := 3; i*i <= nVal; i += 2 {
		if nVal%i == 0 {
			return false
		}
	}

	return true
}

func (Math) primes(count any) []any {
	countVal := int(OSLcastNumber(count))
	primes := []any{2}
	candidate := 3

	for len(primes) < countVal {
		isPrime := true
		for _, p := range primes {
			if int(OSLcastNumber(p))*int(OSLcastNumber(p)) > candidate {
				break
			}
			if candidate%int(OSLcastNumber(p)) == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			primes = append(primes, candidate)
		}
		candidate += 2
	}

	return primes
}

func (Math) degrees(x any) float64 {
	rads := OSLcastNumber(x)
	return rads * (180.0 / math.Pi)
}

func (Math) radians(x any) float64 {
	degs := OSLcastNumber(x)
	return degs * (math.Pi / 180.0)
}

func (Math) random(min any, max any) float64 {
	minVal := OSLcastNumber(min)
	maxVal := OSLcastNumber(max)

	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	return minVal + (rand.Float64() * (maxVal - minVal))
}

func (Math) randomInt(min any, max any) int {
	minVal := int(OSLcastNumber(min))
	maxVal := int(OSLcastNumber(max))

	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	return minVal + rand.Intn(maxVal-minVal+1)
}

func (Math) randomChoice(choices []any) any {
	if len(choices) == 0 {
		return nil
	}
	return choices[math.randomInt(0, len(choices)-1)]
}

func (Math) randomSeed(seed any) {
	seedVal := int64(OSLcastNumber(seed))
	rand.Seed(seedVal)
}

func (Math) hypot(x any, y any) float64 {
	return math.Hypot(OSLcastNumber(x), OSLcastNumber(y))
}

func (Math) distance(x1 any, y1 any, x2 any, y2 any) float64 {
	dx := OSLcastNumber(x2) - OSLcastNumber(x1)
	dy := OSLcastNumber(y2) - OSLcastNumber(y1)
	return math.Sqrt(dx*dx + dy*dy)
}

func (Math) mod(a any, b any) float64 {
	aVal := OSLcastNumber(a)
	bVal := OSLcastNumber(b)

	if bVal == 0 {
		return 0
	}

	return math.Mod(aVal, bVal)
}

func (Math) isNan(x any) bool {
	return math.IsNaN(OSLcastNumber(x))
}

func (Math) isInf(x any) int {
	return math.IsInf(OSLcastNumber(x), 0)
}

func (Math) sign(x any) int {
	val := OSLcastNumber(x)

	if val > 0 {
		return 1
	} else if val < 0 {
		return -1
	}
	return 0
}

func (Math) pi() float64 {
	return math.Pi
}

func (Math) e() float64 {
	return math.E
}

func (Math) phi() float64 {
	return 1.618033988749895
}

func (Math) toFixed(x any, decimals any) string {
	return fmt.Sprintf("%.*f", OSLcastInt(decimals), OSLcastNumber(x))
}

func (Math) toPercent(x any, total any) float64 {
	value := OSLcastNumber(x)
	totalVal := OSLcastNumber(total)

	if totalVal == 0 {
		return 0
	}

	return (value / totalVal) * 100
}

var math = Math{}
